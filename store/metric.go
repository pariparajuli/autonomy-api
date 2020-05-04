package store

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

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
	behaviorToday, behaviorYesterday, err := m.NearestGoodBehavior(consts.CORHORT_DISTANCE_RANGE, location)
	if err != nil {
		log.WithFields(log.Fields{
			"prefix": mongoLogPrefix,
			"error":  err,
		}).Error("nearest good behavior")
		return nil, err
	} else {
		log.WithFields(log.Fields{
			"prefix":            mongoLogPrefix,
			"behaviorToday":     behaviorToday,
			"behaviorYesterday": behaviorYesterday,
		}).Debug("nearest good behavior")
	}

	behaviorScore, behaviorDelta, behaviorCount, totalPeopleReport := scoreUtil.BehaviorScore(behaviorToday, behaviorYesterday)

	officialSymptomDistribution, officialSymptomCount, userCount, err := m.NearestSymptomScore(consts.NEARBY_DISTANCE_RANGE, location, true)
	log.Info(fmt.Sprintf("CollectRawMetrics: officialSymptomDistribution:%v , officialSymptomCount:%v, userCount:%v", officialSymptomDistribution, officialSymptomCount, totalPeopleReport))

	if err != nil {
		log.WithFields(log.Fields{
			"prefix": mongoLogPrefix,
			"error":  err,
		}).Error("nearest official symptom info")
		return nil, err
	} else {
		log.WithFields(log.Fields{
			"prefix":        mongoLogPrefix,
			"distribution":  officialSymptomDistribution,
			"symptom_count": officialSymptomCount,
			"user_count":    userCount,
		}).Debug("nearest official symptom info")
	}

	customSymptoms, err := m.AreaCustomizedSymptomList(consts.NEARBY_DISTANCE_RANGE, location)
	if err != nil {
		log.WithFields(log.Fields{
			"prefix": mongoLogPrefix,
			"error":  err,
		}).Error("nearest custom symptom info")
		return nil, err
	} else {
		log.WithFields(log.Fields{
			"prefix":        mongoLogPrefix,
			"symptom_count": len(customSymptoms),
		}).Debug("nearest customized symptom info")
	}

	confirmedCount, confirmDiff, confirmDiffPercent, err := m.GetConfirm(location)
	if err != nil {
		log.WithFields(log.Fields{
			"prefix": mongoLogPrefix,
			"error":  err,
		}).Error("confirm info")
		return nil, err
	} else {
		log.WithFields(log.Fields{
			"prefix":         mongoLogPrefix,
			"latest_count":   confirmedCount,
			"diff_yesterday": confirmDiff,
			"percent":        confirmDiffPercent,
		}).Debug("nearest official symptom info")

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
				BehaviorTotal:           behaviorCount,
				TotalPeople:             totalPeopleReport,
				MaxScorePerPerson:       schema.TotalOfficialBehaviorWeight,
				CustomizedBehaviorTotal: float64(behaviorToday.CustomizedCount),
				Score:                   behaviorScore,
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
		log.WithFields(log.Fields{
			"prefix":         mongoLogPrefix,
			"account_number": accountNumber,
			"error":          err,
		}).Error("collect raw metrics")
		return nil, err
	}

	log.WithFields(log.Fields{
		"prefix":         mongoLogPrefix,
		"account_number": accountNumber,
		"raw_metrics":    rawMetrics,
	}).Debug("collect raw metrics")

	metric, err = scoreUtil.CalculateMetric(*rawMetrics, &p.Metric)
	if err != nil {
		return nil, err
	}

	if coefficient != nil {
		// TODO: Called in both here metic.go and score/metric.go. Decide where to call
		//scoreUtil.SymptomScore(p.ScoreCoefficient.SymptomWeights, metric, &p.Metric)
		//	scoreUtil.ConfirmScore(metric)

		metric.Score = scoreUtil.TotalScoreV1(*coefficient,
			metric.Details.Symptoms.Score,
			metric.Details.Behaviors.Score,
			metric.Details.Confirm.Score,
		)
	} else {
		// TODO: Called in both here metic.go and score/metric.go. Decide where to call
		//	scoreUtil.SymptomScore(schema.DefaultSymptomWeights, metric, &p.Metric)
		//	scoreUtil.ConfirmScore(metric)

		scoreUtil.DefaultTotalScore(
			metric.Details.Symptoms.Score,
			metric.Details.Behaviors.Score,
			metric.Details.Confirm.Score)
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
