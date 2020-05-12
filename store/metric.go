package store

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bitmark-inc/autonomy-api/consts"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/score"
)

const (
	metricUpdateInterval = 5 * time.Minute
)

type Metric interface {
	CollectRawMetrics(location schema.Location) (*schema.Metric, error)
	SyncAccountMetrics(accountNumber string, coefficient *schema.ScoreCoefficient, location schema.Location) (*schema.Metric, error)
	SyncAccountPOIMetrics(accountNumber string, coefficient *schema.ScoreCoefficient, poiID primitive.ObjectID) (*schema.Metric, error)
	SyncPOIMetrics(poiID primitive.ObjectID, location schema.Location) (*schema.Metric, error)
}

// CollectRawMetrics will gather data from various of sources that is required to
// calculate an autonomy score
func (m *mongoDB) CollectRawMetrics(location schema.Location) (*schema.Metric, error) {
	// Processing behaviors data
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

	behaviorScore, behaviorDelta, behaviorCount, totalPeopleReport := score.BehaviorScore(behaviorToday, behaviorYesterday)

	// Processing symptoms data
	symptomToday, symptomYesterday, err := m.NearestSymptomScore(consts.NEARBY_DISTANCE_RANGE, location)
	log.Info(fmt.Sprintf("CollectRawMetrics: officialSymptomDistribution:%v , officialSymptomCount:%v, userCount:%v", symptomToday.WeightDistribution, symptomToday.OfficialCount, symptomToday.UserCount))
	if err != nil {
		log.WithFields(log.Fields{
			"prefix": mongoLogPrefix,
			"error":  err,
		}).Error("NearestSymptomScore outcome")
		return nil, err
	}

	// Processing confirmed case data
	confirmedCount, confirmDiff, confirmDiffPercent, err := m.GetCDSConfirm(location)
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
		}).Debug("confirm metric")

	}

	return &schema.Metric{
		BehaviorCount:  float64(behaviorCount),
		BehaviorDelta:  float64(behaviorDelta),
		ConfirmedCount: confirmedCount,
		ConfirmedDelta: confirmDiffPercent,
		//	SymptomCount:   sOfficialCount + sCustomizedCount,
		//		SymptomDelta:   sDeltaInPercent,
		Details: schema.Details{
			Confirm: schema.ConfirmDetail{
				Yesterday: confirmedCount - confirmDiff,
				Today:     confirmedCount,
			},
			Symptoms: schema.SymptomDetail{
				TodayData:     symptomToday,
				YesterdayData: symptomYesterday,
			},
			Behaviors: schema.BehaviorDetail{
				BehaviorTotal:           behaviorCount,
				TotalPeople:             totalPeopleReport,
				MaxScorePerPerson:       schema.TotalOfficialBehaviorWeight,
				CustomizedBehaviorTotal: float64(behaviorToday.CustomizedCount),
				Score: behaviorScore,
			},
		},
	}, nil
}

func (m *mongoDB) SyncAccountMetrics(accountNumber string, coefficient *schema.ScoreCoefficient, location schema.Location) (*schema.Metric, error) {
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

	metric := score.CalculateMetric(*rawMetrics, coefficient)

	if err := m.UpdateProfileMetric(accountNumber, metric); err != nil {
		return nil, err
	}

	return &metric, nil
}

func (m *mongoDB) SyncAccountPOIMetrics(accountNumber string, coefficient *schema.ScoreCoefficient, poiID primitive.ObjectID) (*schema.Metric, error) {
	profile, err := m.GetProfile(accountNumber)
	if nil != err {
		log.WithFields(log.Fields{
			"prefix":         mongoLogPrefix,
			"account_number": accountNumber,
		}).Error("get profile")
		return nil, err
	}

	for _, p := range profile.PointsOfInterest {
		if p.ID == poiID {
			poi, err := m.GetPOI(poiID)
			if err != nil {
				return nil, ErrPOINotFound
			}

			location := schema.Location{
				Longitude: poi.Location.Coordinates[0],
				Latitude:  poi.Location.Coordinates[1],
			}
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

			metric := score.CalculateMetric(*rawMetrics, coefficient)

			if err := m.UpdateProfilePOIMetric(accountNumber, poiID, metric); err != nil {
				return nil, err
			}
			return &metric, nil
		}
	}

	return nil, ErrPOINotFound
}

func (m *mongoDB) SyncPOIMetrics(poiID primitive.ObjectID, location schema.Location) (*schema.Metric, error) {
	rawMetrics, err := m.CollectRawMetrics(location)
	if err != nil {
		return nil, err
	}

	metric := score.CalculateMetric(*rawMetrics, nil)
	if err != nil {
		return nil, err
	}

	if err := m.UpdatePOIMetric(poiID, metric); err != nil {
		return nil, err
	}

	return &metric, nil
}
