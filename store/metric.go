package store

import (
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bitmark-inc/autonomy-api/consts"
	"github.com/bitmark-inc/autonomy-api/schema"
	scoreUtil "github.com/bitmark-inc/autonomy-api/score"
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

	behaviorScore, behaviorDelta, behaviorCount, _ := scoreUtil.BehaviorScore(behaviorData)

	officialSymptomDistribution, officialSymptomCount, userCount, err := m.NearOfficialSymptomInfo(consts.NEARBY_DISTANCE_RANGE, location)
	if err != nil {
		return nil, err
	}

	customSymptoms, err := m.AreaCustomizedSymptomList(consts.NEARBY_DISTANCE_RANGE, location)
	if err != nil {
		return nil, err
	}

	confirmedCount, confirmDiff, confirmDiffPercent, err := m.GetConfirm(location)
	if err != nil {
		return nil, err
	}

	return &schema.Metric{
		BehaviorCount:  float64(behaviorCount),
		BehaviorDelta:  float64(behaviorDelta),
		ConfirmedCount: float64(confirmedCount),
		ConfirmedDelta: float64(confirmDiffPercent),
		SymptomCount:   officialSymptomCount + float64(len(customSymptoms)),
		Details: schema.Details{
			Confirm: schema.ConfirmDetail{
				Yesterday: float64(confirmedCount - confirmDiff),
				Today:     float64(confirmedCount),
			},
			Symptoms: schema.SymptomDetail{
				TotalPeople:        userCount,
				Symptoms:           officialSymptomDistribution,
				CustomSymptomCount: float64(len(customSymptoms)),
			},
			Behaviors: schema.BehaviorDetail{
				Score: behaviorScore,
			},
		},
	}, nil
}

func (m *mongoDB) SyncAccountMetrics(accountNumber string, coefficient *schema.ScoreCoefficient, location schema.Location) (*schema.Metric, error) {
	p, err := m.GetProfile(accountNumber)
	if nil != err {
		log.WithFields(log.Fields{
			"prefix":         mongoLogPrefix,
			"account_number": accountNumber,
			"lat":            location.Latitude,
			"lng":            location.Longitude,
		}).Error("get profile")
		return nil, err
	}

	var metric *schema.Metric
	rawMetrics, err := m.CollectRawMetrics(location)
	if err != nil {
		return nil, err
	}

	metric, err = scoreUtil.CalculateMetric(*rawMetrics, &p.Metric)
	if err != nil {
		return nil, err
	}

	if coefficient != nil {
		scoreUtil.SymptomScore(p.ScoreCoefficient.SymptomWeights, metric, &p.Metric)
		scoreUtil.ConfirmScore(metric)

		metric.Score = scoreUtil.TotalScoreV1(*coefficient,
			metric.SymptomCount,
			metric.BehaviorCount,
			metric.ConfirmedCount,
		)
	} else {
		scoreUtil.SymptomScore(schema.DefaultSymptomWeights, metric, &p.Metric)
		scoreUtil.ConfirmScore(metric)

		scoreUtil.DefaultTotalScore(metric.SymptomCount, metric.BehaviorCount, metric.ConfirmedCount)
	}

	if err := m.UpdateProfileMetric(accountNumber, metric); err != nil {
		return nil, err
	}

	return metric, nil
}

func (m *mongoDB) SyncPOIMetrics(poiID primitive.ObjectID, location schema.Location) (*schema.Metric, error) {
	var metric *schema.Metric
	poi, err := m.GetPOI(poiID)
	if err != nil {
		return nil, err
	}

	rawMetrics, err := m.CollectRawMetrics(location)
	if err != nil {
		return nil, err
	}

	metric, err = scoreUtil.CalculateMetric(*rawMetrics, &poi.Metric)
	if err != nil {
		return nil, err
	}

	if err := m.UpdatePOIMetric(poiID, *metric); err != nil {
		return nil, err
	}

	return metric, nil
}
