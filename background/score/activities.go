package score

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"

	"github.com/bitmark-inc/autonomy-api/background"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/score"
)

var ErrInvalidLocation = fmt.Errorf("invalid location of POI")

func (s *ScoreUpdateWorker) CalculatePOIStateActivity(ctx context.Context, id string) (bool, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Query poi for calculating state.", zap.String("poiID", id))

	poiID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}

	poi, err := s.mongo.GetPOI(poiID)
	if err != nil {
		return false, err
	}
	logger.Info("Query poi for calculating state.", zap.Any("poi", poi))

	if poi == nil || poi.Location == nil {
		return false, ErrInvalidLocation
	}

	location := schema.Location{
		Latitude:  poi.Location.Coordinates[1],
		Longitude: poi.Location.Coordinates[0],
	}

	metric, err := score.CalculateMetric(s.mongo, location)
	if err != nil {
		return false, err
	}

	return s.mongo.RefreshPOIState(poiID, *metric)
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

func (s *ScoreUpdateWorker) CalculateAccountStateActivity(ctx context.Context, accountNumber string) (bool, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Query account profile for calculating state.", zap.String("accountNumber", accountNumber))

	profile, err := s.mongo.GetProfile(accountNumber)
	if err != nil {
		return false, err
	}
	logger.Info("Account profile.", zap.Any("profile", profile))

	if profile.Location == nil {
		return false, ErrInvalidLocation
	}

	location := schema.Location{
		Latitude:  profile.Location.Coordinates[1],
		Longitude: profile.Location.Coordinates[0],
	}

	logger.Info("Calculate metric by location.", zap.Any("location", location))
	metric, err := score.CalculateMetric(s.mongo, location)
	if err != nil {
		return false, err
	}

	return s.mongo.RefreshAccountState(accountNumber, *metric)
}

func (s *ScoreUpdateWorker) SendAccountNotificationActivity(ctx context.Context, accountNumber string) error {
	return s.Background.NotifyAccountsByTemplate([]string{accountNumber}, background.CURRENT_LOCATION_STATUS_CHANGE,
		map[string]interface{}{
			"notification_type": "RISK_LEVEL_CHANGED",
		},
	)
}
