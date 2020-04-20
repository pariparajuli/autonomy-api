package score

import (
	"context"

	"github.com/bitmark-inc/autonomy-api/background"
)

func (s *ScoreUpdateWorker) CalculatePOIStateActivity(ctx context.Context, id string) (bool, error) {
	return s.mongo.RefreshPOIState(id)
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
