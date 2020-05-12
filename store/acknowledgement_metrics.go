package store

import (
	"context"
	"errors"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitmark-inc/autonomy-api/schema"
)

type groupedByUserResult struct {
	ID    string `bson:"_id"` // '$account_number'
	Count int64  `bson:"count"`
}

type AcknowledgementMetrics interface {
	GetPersonalReportCount(reportType, accountNumber string) (int64, int64, error)
	GetCommunityAvgReportCount(reportType string, meter int, loc schema.Location) (float64, float64, error)
}

func (m *mongoDB) GetPersonalReportCount(reportType, accountNumber string) (int64, int64, error) {
	var c *mongo.Collection
	switch reportType {
	case "symptom":
		c = m.client.Database(m.database).Collection(schema.SymptomReportCollection)
	case "behavior":
		c = m.client.Database(m.database).Collection(schema.BehaviorReportCollection)
	default:
		return 0, 0, errors.New("undefined report type")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	todayBeginTime := todayStartAt()

	// today
	filter := bson.M{
		"$and": []bson.M{
			{"account_number": accountNumber},
			matchReportedToday(todayBeginTime),
		},
	}
	countToday, err := c.CountDocuments(ctx, filter)
	if err != nil {
		return 0, 0, err
	}

	// yesterday
	filter = bson.M{
		"$and": []bson.M{
			{"account_number": accountNumber},
			matchReportedYesterday(todayBeginTime),
		},
	}
	countYesterday, err := c.CountDocuments(ctx, filter)
	if err != nil {
		return 0, 0, err
	}

	return countToday, countYesterday, nil
}

func (m *mongoDB) GetCommunityAvgReportCount(reportType string, meter int, loc schema.Location) (float64, float64, error) {
	var c *mongo.Collection
	switch reportType {
	case "symptom":
		c = m.client.Database(m.database).Collection(schema.SymptomReportCollection)
	case "behavior":
		c = m.client.Database(m.database).Collection(schema.BehaviorReportCollection)
	default:
		return 0, 0, errors.New("undefined report type")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	todayBeginTime := todayStartAt()

	// today
	pipeline := []bson.M{
		aggStageGeoProximity(meter, loc),
		aggStageReportedToday(todayBeginTime),
		aggStageUserReportCount(),
	}
	cur, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"prefix": mongoLogPrefix}).Error("failed to count user report")
		return 0, 0, err
	}
	result := make([]groupedByUserResult, 0)
	for cur.Next(ctx) {
		var item groupedByUserResult
		if err = cur.Decode(&item); err != nil {
			log.WithError(err).WithFields(log.Fields{"prefix": mongoLogPrefix}).Error("failed to decode aggreaged result")
			return 0, 0, err
		}
		result = append(result, item)
	}
	avgToday := calculateAvgCount(result)

	// yesterday
	pipeline = []bson.M{
		aggStageGeoProximity(meter, loc),
		aggStageReportedYesterday(todayBeginTime),
		aggStageUserReportCount(),
	}

	cur, err = c.Aggregate(ctx, pipeline)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"prefix": mongoLogPrefix}).Error("failed to count user report")
		return 0, 0, err
	}
	result = make([]groupedByUserResult, 0)
	for cur.Next(ctx) {
		var item groupedByUserResult
		if err = cur.Decode(&item); err != nil {
			log.WithError(err).WithFields(log.Fields{"prefix": mongoLogPrefix}).Error("failed to decode aggreaged result")
			return 0, 0, err
		}
		result = append(result, item)
	}
	avgYesterday := calculateAvgCount(result)

	log.WithFields(log.Fields{"prefix": mongoLogPrefix, "avg_today": avgToday, "avg_yesterday": avgYesterday}).Debug("community avg report")

	return avgToday, avgYesterday, nil
}

func calculateAvgCount(result []groupedByUserResult) (avg float64) {
	var reportCount, userCount int64
	for _, r := range result {
		reportCount += r.Count
		userCount++
	}
	if userCount > 0 {
		avg = float64(reportCount) / float64(userCount)
	}
	return
}
