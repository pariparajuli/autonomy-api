package store

import (
	"fmt"

	"github.com/lib/pq"

	"github.com/bitmark-inc/autonomy-api/schema"
)

var (
	ErrRequestNotExist     = fmt.Errorf("the request is either solved or not open for you")
	ErrMultipleRequestMade = fmt.Errorf("making multiple requests is not allowed")
)

// RequestHelp create a help entry
func (s *AutonomyStore) RequestHelp(accountNumber, subject, needs, meetingPlace, contactInfo string) (*schema.HelpRequest, error) {
	help := schema.HelpRequest{
		Requester:    accountNumber,
		Subject:      subject,
		Needs:        needs,
		MeetingPlace: meetingPlace,
		ContactInfo:  contactInfo,
	}

	if err := s.ormDB.Create(&help).Error; err != nil {
		pqErr := err.(*pq.Error)
		if pqErr.Code == "23505" {
			return nil, ErrMultipleRequestMade
		}
		return nil, err
	}
	return &help, nil
}

// ListHelps first queries accounts within 50KM and returns lists of help
// requests by those accounts
func (s *AutonomyStore) ListHelps(accountNumber string, latitude, longitude float64, count int64) ([]schema.HelpRequest, error) {
	helps := []schema.HelpRequest{}

	// HARDCODED 50KM
	accounts, err := s.mongo.NearestDistance(50000, schema.Location{
		Latitude:  latitude,
		Longitude: longitude,
	})
	if err != nil {
		return nil, err
	}

	if err := s.ormDB.Raw(
		`SELECT * FROM help_requests
		JOIN unnest(?::text[]) WITH ORDINALITY account(requester, index) USING (requester)
		WHERE (requester = ? OR state = ?) AND created_at > now() - INTERVAL '12 hours'
		ORDER BY account.index;`, // HARDCODED: 12 hours of expiration
		pq.Array(accounts),
		accountNumber,
		schema.HELP_PENDING,
	).Scan(&helps).Error; err != nil {
		return nil, err
	}

	return helps, nil
}

func (s *AutonomyStore) GetHelp(helpID string) (*schema.HelpRequest, error) {
	var help schema.HelpRequest

	if err := s.ormDB.Where("id = ?", helpID).First(&help).Error; err != nil {
		return nil, err
	}

	return &help, nil
}

// AnswerHelp set a request to `RESPONDED`. A request could be updated only when
// its state is `PENDING` and the helper is not the same as the requester.
func (s *AutonomyStore) AnswerHelp(accountNumber string, helpID string) error {
	result := s.ormDB.Model(schema.HelpRequest{}).
		Where("id = ? AND requester != ? AND state = ?", helpID, accountNumber, schema.HELP_PENDING).
		Updates(map[string]interface{}{
			"state":  schema.HELP_RESPONDED,
			"helper": accountNumber,
		})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrRequestNotExist
	}

	return nil
}
