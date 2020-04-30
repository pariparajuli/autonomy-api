package store

import (
	"time"

	"github.com/bitmark-inc/autonomy-api/consts"
	"github.com/bitmark-inc/autonomy-api/schema"
	scoreUtil "github.com/bitmark-inc/autonomy-api/score"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	metricUpdateInterval = 5 * time.Minute
)

type Metric interface {
	CollectRawMetrics(location schema.Location) (*schema.Metric, error)
	SyncAccountMetrics(accountNumber string, coefficient *schema.ScoreCoefficient, location schema.Location) (*schema.Metric, error)
	SyncPOIMetrics(poiID primitive.ObjectID, location schema.Location) (*schema.Metric, error)
}

func (m *mongoDB) CollectRawMetrics(location schema.Location) (*schema.Metric, error) {
	behaviorData, err := m.NearestGoodBehavior(consts.CORHORT_DISTANCE_RANGE, location)
	if err != nil {
		return nil, err
	}

	behaviorScore, behaviorDelta, behaviorCount, _, err := scoreUtil.BehaviorScore(behaviorData)
	if err != nil {
		return nil, err
	}

	symptomScore, _, symptomCount, symptomDelta, err := m.NearestSymptomScore(consts.CORHORT_DISTANCE_RANGE, location)
	if err != nil {
		return nil, err
	}

	confirmedCount, confirmedDelta, err := m.GetConfirm(location)
	if err != nil {
		return nil, err
	}

	confirmedScore, err := m.ConfirmScore(location)
	if err != nil {
		return nil, err
	}

	return &schema.Metric{
		SymptomCount:   float64(symptomCount),
		SymptomDelta:   float64(symptomDelta),
		SymptomScore:   float64(symptomScore),
		BehaviorCount:  float64(behaviorCount),
		BehaviorDelta:  float64(behaviorDelta),
		BehaviorScore:  float64(behaviorScore),
		ConfirmedCount: float64(confirmedCount),
		ConfirmedDelta: float64(confirmedDelta),
		ConfirmedScore: float64(confirmedScore),
	}, nil
}

func (m *mongoDB) SyncAccountMetrics(accountNumber string, coefficient *schema.ScoreCoefficient, location schema.Location) (*schema.Metric, error) {
	var metric *schema.Metric
	rawMetrics, err := m.CollectRawMetrics(location)
	if err != nil {
		return nil, err
	}

	metric, err = scoreUtil.CalculateMetric(*rawMetrics)
	if err != nil {
		return nil, err
	}

	if coefficient != nil {
		metric.Score = scoreUtil.TotalScoreV1(*coefficient,
			metric.SymptomScore,
			metric.BehaviorScore,
			metric.ConfirmedScore,
		)
	} else {
		scoreUtil.DefaultTotalScore(metric.SymptomScore, metric.BehaviorScore, metric.ConfirmedScore)
	}

	if err := m.UpdateProfileMetric(accountNumber, metric); err != nil {
		return nil, err
	}

	return metric, nil
}

func (m *mongoDB) SyncPOIMetrics(poiID primitive.ObjectID, location schema.Location) (*schema.Metric, error) {
	var metric *schema.Metric
	rawMetrics, err := m.CollectRawMetrics(location)
	if err != nil {
		return nil, err
	}

	metric, err = scoreUtil.CalculateMetric(*rawMetrics)
	if err != nil {
		return nil, err
	}

	if err := m.UpdatePOIMetric(poiID, *metric); err != nil {
		return nil, err
	}

	return metric, nil
}
