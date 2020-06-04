package store

import (
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bitmark-inc/autonomy-api/consts"
	"github.com/bitmark-inc/autonomy-api/geo"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/score"
)

const (
	metricUpdateInterval = 5 * time.Minute
)

type Metric interface {
	CollectRawMetrics(location schema.Location) (*schema.Metric, error)
	SyncProfileIndividualMetrics(profileID string) (*schema.IndividualMetric, error)
	SyncAccountMetrics(accountNumber string, coefficient *schema.ScoreCoefficient, location schema.Location) (*schema.Metric, error)
	SyncAccountPOIMetrics(accountNumber string, coefficient *schema.ScoreCoefficient, poiID primitive.ObjectID) (*schema.Metric, error)
	SyncPOIMetrics(poiID primitive.ObjectID, location schema.Location) (*schema.Metric, error)
}

// CollectRawMetrics will gather data from various of sources that is required to
// calculate an autonomy score
func (m *mongoDB) CollectRawMetrics(location schema.Location) (*schema.Metric, error) {
	now := time.Now().UTC()
	todayStartAt := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	yesterdayStartAtUnix := todayStartAt.AddDate(0, 0, -1).Unix()
	todayStartAtUnix := todayStartAt.Unix()
	tomorrowStartAtUnix := todayStartAt.AddDate(0, 0, 1).Unix()

	behaviorDistrToday, err := m.FindNearbyBehaviorDistribution(consts.NEARBY_DISTANCE_RANGE, location, todayStartAtUnix, tomorrowStartAtUnix)
	if err != nil {
		return nil, err
	}
	behaviorDistrYesterday, err := m.FindNearbyBehaviorDistribution(consts.NEARBY_DISTANCE_RANGE, location, yesterdayStartAtUnix, todayStartAtUnix)
	if err != nil {
		return nil, err
	}
	behaviorReportTimes, err := m.FindNearbyBehaviorReportTimes(consts.NEARBY_DISTANCE_RANGE, location, todayStartAtUnix, tomorrowStartAtUnix)
	if err != nil {
		return nil, err
	}

	symptomDistToday, err := m.FindNearbySymptomDistribution(consts.NEARBY_DISTANCE_RANGE, location, todayStartAtUnix, tomorrowStartAtUnix)
	if err != nil {
		return nil, err
	}
	symptomDistYesterday, err := m.FindNearbySymptomDistribution(consts.NEARBY_DISTANCE_RANGE, location, yesterdayStartAtUnix, todayStartAtUnix)
	if err != nil {
		return nil, err
	}
	symptomUserCount, err := m.GetNearbyReportingUserCount(schema.ReportTypeSymptom, consts.NEARBY_DISTANCE_RANGE, location, now)
	if err != nil {
		return nil, err
	}

	if location.Country == "" {
		log.Info("fetch poi geo info from external service")
		var err error
		location, err = geo.PoliticalGeoInfo(location)
		if err != nil {
			log.WithError(err).WithField("location", location).Error("fail to fetch geo info")
			return nil, err
		}
	}

	// Processing confirmed case data
	activeCount, activeDiff, activeDiffPercent, err := m.GetCDSActive(location)
	if err == ErrNoConfirmDataset || err == ErrInvalidConfirmDataset || err == ErrPoliticalTypeGeoInfo {
		log.WithFields(log.Fields{
			"prefix":   mongoLogPrefix,
			"location": location,
			"err":      err,
		}).Warn("collect confirm raw metrics")
	} else if err != nil {
		log.WithFields(log.Fields{
			"prefix": mongoLogPrefix,
			"error":  err,
		}).Error("confirm info")
		return nil, err
	} else {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "activeCount": activeCount, "activeDiff": activeDiff, "activeDiffPercent": activeDiffPercent}).Debug("confirm info")
	}

	confirmData, err := m.ContinuousDataCDSConfirm(location, consts.ConfirmScoreWindowSize, 0)

	if err == ErrNoConfirmDataset || err == ErrInvalidConfirmDataset || err == ErrPoliticalTypeGeoInfo {
		log.WithFields(log.Fields{
			"prefix":   mongoLogPrefix,
			"location": location,
			"err":      err,
		}).Warn("collect continuous confirm raw metrics")
		confirmData = []schema.CDSScoreDataSet{}
	} else if err != nil {
		log.WithFields(log.Fields{
			"prefix": mongoLogPrefix,
			"error":  err,
		}).Error("continuous confirm info")
		return nil, err
	} else {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "activeCount": activeCount, "activeDiff": activeDiff, "activeDiffPercent": activeDiffPercent}).Debug("confirm info")
	}

	return &schema.Metric{
		ConfirmedCount: activeCount,
		ConfirmedDelta: activeDiffPercent,
		Details: schema.Details{
			Confirm: schema.ConfirmDetail{
				ContinuousData: confirmData,
			},
			Symptoms: schema.SymptomDetail{
				TotalPeople: float64(symptomUserCount),
				TodayData: schema.NearestSymptomData{
					WeightDistribution: symptomDistToday,
				},
				YesterdayData: schema.NearestSymptomData{
					WeightDistribution: symptomDistYesterday,
				},
			},
			Behaviors: schema.BehaviorDetail{
				ReportTimes:           behaviorReportTimes,
				TodayDistribution:     behaviorDistrToday,
				YesterdayDistribution: behaviorDistrYesterday,
			},
		},
	}, nil
}

// SyncProfileIndividualMetrics calculate individual metrics and save into profile
func (m *mongoDB) SyncProfileIndividualMetrics(profileID string) (*schema.IndividualMetric, error) {
	now := time.Now().UTC()

	symptomsToday, symptomsYesterday, err := m.GetSymptomCount(profileID, nil, 0, now)
	if err != nil {
		return nil, err
	}

	symptomsDelta := score.ChangeRate(float64(symptomsToday), float64(symptomsYesterday))

	behaviorsToday, behaviorsYesterday, err := m.GetBehaviorCount(profileID, nil, 0, now)
	if err != nil {
		return nil, err
	}

	behaviorsDelta := score.ChangeRate(float64(behaviorsToday), float64(behaviorsYesterday))

	metric := schema.IndividualMetric{
		SymptomCount:  float64(symptomsToday),
		SymptomDelta:  symptomsDelta,
		BehaviorCount: float64(behaviorsToday),
		BehaviorDelta: behaviorsDelta,
	}

	if err := m.UpdateProfileIndividualMetric(profileID, metric); err != nil {
		return nil, err
	}

	return &metric, nil
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
				AddressComponent: schema.AddressComponent{
					Country: poi.Country,
					State:   poi.State,
					County:  poi.County,
				},
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
