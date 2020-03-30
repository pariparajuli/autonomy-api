package store

import (
	"fmt"

	"github.com/bitmark-inc/autonomy-api/schema"
)

var (
	ErrRequestNotExist = fmt.Errorf("the request is either solved or not open for you")
)

// RequestHelp create a help entry
func (s *AutonomyStore) RequestHelp(accountNumber, subject, text string) (*schema.HelpRequest, error) {
	help := schema.HelpRequest{
		Requester: accountNumber,
		Subject:   subject,
		Text:      text,
	}

	if err := s.ormDB.Create(&help).Error; err != nil {
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
