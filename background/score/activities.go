package score

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bitmark-inc/autonomy-api/background"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/score"
)

func (s *ScoreUpdateWorker) CalculatePOIStateActivity(ctx context.Context, id string) (bool, error) {
	poiID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}
	poi, err := s.mongo.GetPOI(poiID)
	if err != nil {
		return false, err
	}

	confirmedScore, err := s.mongo.ConfirmScore(schema.Location{
		Latitude:  poi.Location.Coordinates[1],
		Longitude: poi.Location.Coordinates[0],
	})
	if err != nil {
		return false, err
	}

	totalScore := score.TotalScore(0, 0, confirmedScore)

	return s.mongo.RefreshPOIState(poiID, totalScore)
}

func (s *ScoreUpdateWorker) SendPOINotificationActivity(ctx context.Context, id string) error {
	accounts, err := s.mongo.GetAccountsByPOI(id)
	if err != nil {
		return err
	}
	return s.Background.NotifyAccountsByTemplate(accounts, background.SAVED_LOCATION_STATUS_CHANGE,
		map[string]interface{}{
			"notification_type": "RISK_LEVEL_CHANGED",
			"poi_id":            id,
		},
	)
}

func (s *ScoreUpdateWorker) CalculateAccountStateActivity(ctx context.Context, account_number string) (bool, error) {
	return s.mongo.RefreshAccountState(account_number)
}

func (s *ScoreUpdateWorker) SendAccountNotificationActivity(ctx context.Context, account_number string) error {
	return s.Background.NotifyAccountsByTemplate([]string{account_number}, background.CURRENT_LOCATION_STATUS_CHANGE,
		map[string]interface{}{
			"notification_type": "RISK_LEVEL_CHANGED",
		},
	)
}
