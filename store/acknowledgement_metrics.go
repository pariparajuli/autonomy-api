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
	GetPersonalReportedItemCount(reportType, accountNumber string) (int64, int64, error)
	GetCommunityAvgReportedItemCount(reportType string, meter int, loc schema.Location) (float64, float64, error)
}

func (m *mongoDB) GetPersonalReportedItemCount(reportType, accountNumber string) (int64, int64, error) {
	var c *mongo.Collection
	var fields []string
	switch reportType {
	case "symptom":
		c = m.client.Database(m.database).Collection(schema.SymptomReportCollection)
		fields = []string{"official_symptoms", "customized_symptoms"}
	case "behavior":
		c = m.client.Database(m.database).Collection(schema.BehaviorReportCollection)
		fields = []string{"official_behaviors", "customized_behaviors"}
	default:
		return 0, 0, errors.New("undefined report type")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	todayBeginTime := todayStartAt()

	// today
	pipeline := []bson.M{
		{"$match": bson.M{"account_number": accountNumber}},
		aggStageReportedToday(todayBeginTime),
		aggStagePreventNullArray(fields...),
		aggStageReportedItemCount("count", fields...),
		aggStageSumValues("count", "total"),
	}
	cur, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"prefix": mongoLogPrefix}).Error("failed to count user reported items")
		return 0, 0, err
	}
	type aggregatedItem struct {
		Total int64 `bson:"total"`
	}
	result := make([]aggregatedItem, 0)
	for cur.Next(ctx) {
		var item aggregatedItem
		if err = cur.Decode(&item); err != nil {
			log.WithError(err).WithFields(log.Fields{"prefix": mongoLogPrefix}).Error("failed to decode aggreaged result")
			return 0, 0, err
		}
		result = append(result, item)
	}
	countToday := int64(0)
	if len(result) > 0 {
		countToday = result[0].Total
	}

	// yesterday
	pipeline = []bson.M{
		{"$match": bson.M{"account_number": accountNumber}},
		aggStageReportedYesterday(todayBeginTime),
		aggStagePreventNullArray(fields...),
		aggStageReportedItemCount("count", fields...),
		aggStageSumValues("count", "total"),
	}

	cur, err = c.Aggregate(ctx, pipeline)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"prefix": mongoLogPrefix}).Error("failed to count user reported items")
		return 0, 0, err
	}
	result = make([]aggregatedItem, 0)
	for cur.Next(ctx) {
		var item aggregatedItem
		if err = cur.Decode(&item); err != nil {
			log.WithError(err).WithFields(log.Fields{"prefix": mongoLogPrefix}).Error("failed to decode aggreaged result")
			return 0, 0, err
		}
		result = append(result, item)
	}
	countYesterday := int64(0)
	if len(result) > 0 {
		countYesterday = result[0].Total
	}

	log.WithFields(log.Fields{"prefix": mongoLogPrefix, "cnt_today": countToday, "cnt_yesterday": countYesterday}).Debug("personal reported item count")

	return countToday, countYesterday, nil
}

func (m *mongoDB) GetCommunityAvgReportedItemCount(reportType string, meter int, loc schema.Location) (float64, float64, error) {
	var c *mongo.Collection
	var fields []string
	switch reportType {
	case "symptom":
		c = m.client.Database(m.database).Collection(schema.SymptomReportCollection)
		fields = []string{"official_symptoms", "customized_symptoms"}

	case "behavior":
		c = m.client.Database(m.database).Collection(schema.BehaviorReportCollection)
		fields = []string{"official_behaviors", "customized_behaviors"}
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
		aggStagePreventNullArray(fields...),
		aggStageReportedItemCount("count", fields...),
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
		aggStagePreventNullArray(fields...),
		aggStageReportedItemCount("count", fields...),
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

	log.WithFields(log.Fields{"prefix": mongoLogPrefix, "avg_today": avgToday, "avg_yesterday": avgYesterday}).Debug("community avg reported item count")

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
