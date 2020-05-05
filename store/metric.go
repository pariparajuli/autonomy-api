package store

import (
	"fmt"
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
	SyncAccountPOIMetrics(accountNumber string, coefficient *schema.ScoreCoefficient, poiID primitive.ObjectID) (*schema.Metric, error)
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

	symptomToday, symptomYesterday, err := m.NearestSymptomScore(consts.NEARBY_DISTANCE_RANGE, location)
	log.Info(fmt.Sprintf("CollectRawMetrics: officialSymptomDistribution:%v , officialSymptomCount:%v, userCount:%v", symptomToday.WeightDistribution, symptomToday.OfficialCount, symptomToday.UserCount))
	if err != nil {
		log.WithFields(log.Fields{
			"prefix": mongoLogPrefix,
			"error":  err,
		}).Error("NearestSymptomScore outcome")
		return nil, err
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
		//	SymptomCount:   sOfficialCount + sCustomizedCount,
		//		SymptomDelta:   sDeltaInPercent,
		Details: schema.Details{
			Confirm: schema.ConfirmDetail{
				Yesterday: float64(confirmedCount - confirmDiff),
				Today:     float64(confirmedCount),
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

	var symptomScore, sTotalweight, sMaxScorePerPerson, sDeltaInPercent, sOfficialCount, sCustomizedCount float64
	if coefficient != nil {
		// TODO: Called in both here metic.go and score/metric.go. Decide where to call
		// scoreUtil.SymptomScore(p.ScoreCoefficient.SymptomWeights, metric, &p.Metric)
		//scoreUtil.ConfirmScore(metric)
		// Symptom
		symptomScore, sTotalweight, sMaxScorePerPerson, sDeltaInPercent, sOfficialCount, sCustomizedCount =
			scoreUtil.SymptomScore(p.ScoreCoefficient.SymptomWeights, rawMetrics.Details.Symptoms.TodayData, rawMetrics.Details.Symptoms.YesterdayData)
		metric.Details.Symptoms.Score = symptomScore
		metric.Score = scoreUtil.TotalScoreV1(*coefficient,
			metric.Details.Symptoms.Score,
			metric.Details.Behaviors.Score,
			metric.Details.Confirm.Score,
		)
	} else {
		symptomScore, sTotalweight, sMaxScorePerPerson, sDeltaInPercent, sOfficialCount, sCustomizedCount =
			scoreUtil.SymptomScore(schema.DefaultSymptomWeights, rawMetrics.Details.Symptoms.TodayData, rawMetrics.Details.Symptoms.YesterdayData)
		metric.Details.Symptoms.Score = symptomScore
		scoreUtil.DefaultTotalScore(
			metric.Details.Symptoms.Score,
			metric.Details.Behaviors.Score,
			metric.Details.Confirm.Score)
	}
	metric.SymptomDelta = sDeltaInPercent
	metric.SymptomCount = sOfficialCount + sCustomizedCount
	metric.Details.Symptoms = schema.SymptomDetail{
		SymptomTotal:       sTotalweight,
		TotalPeople:        rawMetrics.Details.Symptoms.TodayData.UserCount,
		Symptoms:           rawMetrics.Details.Symptoms.TodayData.WeightDistribution,
		MaxScorePerPerson:  sMaxScorePerPerson,
		CustomizedWeight:   sCustomizedCount,
		CustomSymptomCount: sCustomizedCount,
		Score:              symptomScore,
	}
	log.WithFields(log.Fields{
		"prefix":             mongoLogPrefix,
		"SymptomDelta":       sDeltaInPercent,
		"SymptomCount":       sOfficialCount + sCustomizedCount,
		"SymptomTotal":       metric.Details.Symptoms.SymptomTotal,
		"TotalPeople":        metric.Details.Symptoms.TotalPeople,
		"MaxScorePerPerson":  metric.Details.Symptoms.MaxScorePerPerson,
		"CustomizedWeight":   metric.Details.Symptoms.CustomizedWeight,
		"CustomSymptomCount": metric.Details.Symptoms.CustomSymptomCount,
	}).Debug("SyncAccountMetrics symptom Info")
	if err := m.UpdateProfileMetric(accountNumber, metric); err != nil {
		return nil, err
	}

	return metric, nil
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
			symptomToday, symptomYesterday, err := m.NearestSymptomScore(consts.NEARBY_DISTANCE_RANGE, location)
			log.Info(fmt.Sprintf("CollectRawMetrics: officialSymptomDistribution:%v , officialSymptomCount:%v, userCount:%v", symptomToday.WeightDistribution, symptomToday.OfficialCount, symptomToday.UserCount))
			if err != nil {
				log.WithFields(log.Fields{
					"prefix": mongoLogPrefix,
					"error":  err,
				}).Error("NearestSymptomScore outcome")
				return nil, err
			}

			metric := poi.Metric

			var symptomScore, sTotalweight, sMaxScorePerPerson, sDeltaInPercent, sOfficialCount, sCustomizedCount float64
			if coefficient != nil {
				// TODO: Called in both here metic.go and score/metric.go. Decide where to call
				//scoreUtil.SymptomScore(p.ScoreCoefficient.SymptomWeights, metric, &p.Metric)
				//	scoreUtil.ConfirmScore(metric)
				// Symptom
				symptomScore, sTotalweight, sMaxScorePerPerson, sDeltaInPercent, sOfficialCount, sCustomizedCount = scoreUtil.SymptomScore(profile.ScoreCoefficient.SymptomWeights, symptomToday, symptomYesterday)
				metric.Details.Symptoms.Score = symptomScore
				metric.Score = scoreUtil.TotalScoreV1(*coefficient,
					metric.Details.Symptoms.Score,
					metric.Details.Behaviors.Score,
					metric.Details.Confirm.Score,
				)
			} else {
				// TODO: Called in both here metic.go and score/metric.go. Decide where to call
				//	scoreUtil.SymptomScore(schema.DefaultSymptomWeights, metric, &p.Metric)
				scoreUtil.ConfirmScore(&metric)
				symptomScore, sTotalweight, sMaxScorePerPerson, sDeltaInPercent, sOfficialCount, sCustomizedCount = scoreUtil.SymptomScore(schema.DefaultSymptomWeights, symptomToday, symptomYesterday)
				metric.Details.Symptoms.Score = symptomScore
				scoreUtil.DefaultTotalScore(
					metric.Details.Symptoms.Score,
					metric.Details.Behaviors.Score,
					metric.Details.Confirm.Score)
			}
			metric.SymptomDelta = sDeltaInPercent
			metric.SymptomCount = sOfficialCount + sCustomizedCount
			metric.Details.Symptoms = schema.SymptomDetail{
				SymptomTotal:       sTotalweight,
				TotalPeople:        symptomToday.UserCount,
				MaxScorePerPerson:  sMaxScorePerPerson,
				CustomizedWeight:   sCustomizedCount,
				CustomSymptomCount: sCustomizedCount,
				Symptoms:           symptomToday.WeightDistribution,
				Score:              symptomScore,
			}
			log.WithFields(log.Fields{
				"prefix":             mongoLogPrefix,
				"SymptomDelta":       sDeltaInPercent,
				"SymptomCount":       sOfficialCount + sCustomizedCount,
				"SymptomTotal":       metric.Details.Symptoms.SymptomTotal,
				"TotalPeople":        metric.Details.Symptoms.TotalPeople,
				"MaxScorePerPerson":  metric.Details.Symptoms.MaxScorePerPerson,
				"CustomizedWeight":   metric.Details.Symptoms.CustomizedWeight,
				"CustomSymptomCount": metric.Details.Symptoms.CustomSymptomCount,
			}).Debug("SyncAccountMetrics symptom Info")

			if err := m.UpdateProfilePOIMetric(accountNumber, poiID, metric); err != nil {
				return nil, err
			}
			return &metric, nil
		}
	}

	return nil, ErrPOINotFound
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
