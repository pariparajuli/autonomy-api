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

var ErrInvalidLocation = fmt.Errorf("invalid location")

// CalculatePOIStateActivity calculates metrics by the location of a POI
func (s *ScoreUpdateWorker) CalculatePOIStateActivity(ctx context.Context, id string) (*schema.Metric, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Query poi for calculating state.", zap.String("poiID", id))

	poiID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	poi, err := s.mongo.GetPOI(poiID)
	if err != nil {
		return nil, err
	}

	if poi == nil || poi.Location == nil {
		return nil, ErrInvalidLocation
	}

	location := schema.Location{
		Latitude:  poi.Location.Coordinates[1],
		Longitude: poi.Location.Coordinates[0],
	}

	logger.Info("Calculate metric by location.", zap.Any("location", location))
	rawMetrics, err := s.mongo.CollectRawMetrics(location)
	if err != nil {
		return nil, err
	}

	return score.CalculateMetric(*rawMetrics, &poi.Metric)
}

// RefreshLocationStateActivity updates the metrics as well as the score if the POI id
// is not provided. Otherwise, it updates the score of POIs in the profile.
// It will return accounts whose score's color is changed.
func (s *ScoreUpdateWorker) RefreshLocationStateActivity(ctx context.Context, accountNumber, poiID string, metric schema.Metric) ([]string, error) {
	logger := activity.GetLogger(ctx)
	updatedAccounts := make([]string, 0)

	if poiID != "" {
		id, err := primitive.ObjectIDFromHex(poiID)
		if err != nil {
			return nil, err
		}

		if err := s.mongo.UpdatePOIMetric(id, metric); err != nil {
			return nil, err
		}

		profiles, err := s.mongo.GetProfilesByPOI(poiID)
		if err != nil {
			return nil, err
		}

		for _, profile := range profiles {
			if profile.ScoreCoefficient != nil {
				metric.Score = score.TotalScoreV1(*profile.ScoreCoefficient, metric.SymptomScore, metric.BehaviorScore, metric.ConfirmedScore)
			}

			if err := s.mongo.UpdateProfilePOIMetric(profile.AccountNumber, id, metric); err != nil {
				return nil, err
			}

			var changed bool
			if len(profile.PointsOfInterest) != 0 {
				changed = score.CheckScoreColorChange(profile.PointsOfInterest[0].Score, metric.Score)
			}

			if changed {
				logger.Debug("State color changed", zap.Any("old", profile.Metric.Score), zap.Any("new", metric.Score))
				updatedAccounts = append(updatedAccounts, profile.AccountNumber)
			}
		}
	} else { // poiID == ''
		profile, err := s.mongo.GetProfile(accountNumber)
		if err != nil {
			return nil, err
		}

		if profile.ScoreCoefficient != nil {
			metric.Score = score.TotalScoreV1(*profile.ScoreCoefficient, metric.SymptomScore, metric.BehaviorScore, metric.ConfirmedScore)
		}

		if err := s.mongo.UpdateProfileMetric(accountNumber, &metric); err != nil {
			return nil, err
		}

		var changed bool
		if profile.Metric.LastUpdate != 0 {
			changed = score.CheckScoreColorChange(profile.Metric.Score, metric.Score)
		}

		if changed {
			logger.Debug("State color changed", zap.Any("old", profile.Metric.Score), zap.Any("new", metric.Score))
			updatedAccounts = append(updatedAccounts, profile.AccountNumber)
		}
	}
	return updatedAccounts, nil
}

// NotifyLocationStateActivity is to send notification to end users for notifing the
// significant changes of location states.
func (s *ScoreUpdateWorker) NotifyLocationStateActivity(ctx context.Context, id string, accounts []string) error {
	logger := activity.GetLogger(ctx)
	if len(accounts) == 0 {
		logger.Warn("Send notification without accounts")
		return nil
	}

	if id == "" {
		return s.Background.NotifyAccountsByTemplate(accounts, background.CURRENT_LOCATION_STATUS_CHANGE,
			map[string]interface{}{
				"notification_type": "RISK_LEVEL_CHANGED",
			},
		)
	}

	return s.Background.NotifyAccountsByTemplate(accounts, background.SAVED_LOCATION_STATUS_CHANGE,
		map[string]interface{}{
			"notification_type": "RISK_LEVEL_CHANGED",
			"poi_id":            id,
		},
	)
}

// CalculateAccountStateActivity calculates metrics by a given account's location
func (s *ScoreUpdateWorker) CalculateAccountStateActivity(ctx context.Context, accountNumber string) (*schema.Metric, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Query account profile for calculating state.", zap.String("accountNumber", accountNumber))

	profile, err := s.mongo.GetProfile(accountNumber)
	if err != nil {
		return nil, err
	}
	logger.Info("Account profile.", zap.Any("profile", profile))

	if profile.Location == nil {
		return nil, ErrInvalidLocation
	}

	location := schema.Location{
		Latitude:  profile.Location.Coordinates[1],
		Longitude: profile.Location.Coordinates[0],
	}

	logger.Info("Calculate metric by location.", zap.Any("location", location))
	rawMetrics, err := s.mongo.CollectRawMetrics(location)
	if err != nil {
		return nil, err
	}

	return score.CalculateMetric(*rawMetrics, &profile.Metric)
}
