package store

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitmark-inc/autonomy-api/schema"
)

type Report interface {
	GetNearbyReportingUserCount(reportType schema.ReportType, dist int, loc schema.Location, now time.Time) (int, error)
}

// GetNearbyReportingUserCount returns the number of users who have reported symptoms/behaviors
// in the specified area and within the specified day.
func (m *mongoDB) GetNearbyReportingUserCount(reportType schema.ReportType, dist int, loc schema.Location, now time.Time) (int, error) {
	var c *mongo.Collection
	switch reportType {
	case schema.ReportTypeSymptom:
		c = m.client.Database(m.database).Collection(schema.SymptomReportCollection)
	case schema.ReportTypeBehavior:
		c = m.client.Database(m.database).Collection(schema.BehaviorReportCollection)
	default:
		return 0, errors.New("invalid report type")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, todayStartAt, tomorrowStartAt := getStartTimeOfConsecutiveDays(now)

	pipeline := []bson.M{
		aggStageGeoProximity(dist, loc),
		aggStageReportedBetween(todayStartAt.Unix(), tomorrowStartAt.Unix()),
		{
			"$group": bson.M{
				"_id": "$profile_id",
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	cursor, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}

	if !cursor.Next(ctx) {
		return 0, nil
	}

	var result struct {
		Count int `bson:"count"`
	}
	if err := cursor.Decode(&result); err != nil {
		return 0, err
	}

	return result.Count, nil
}
