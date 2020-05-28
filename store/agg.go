package store

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/bitmark-inc/autonomy-api/schema"
)

func aggStageGeoProximity(maxDistance int, location schema.Location) bson.M {
	return bson.M{
		"$geoNear": bson.M{
			"near": bson.M{
				"type":        "Point",
				"coordinates": bson.A{location.Longitude, location.Latitude},
			},
			"distanceField": "dist",
			"maxDistance":   maxDistance,
			"spherical":     true,
			"includeLocs":   "location",
		},
	}
}

func aggStageReportedBetween(start, end int64) bson.M {
	return bson.M{
		"$match": bson.M{
			"ts": bson.M{
				"$gte": start,
				"$lt":  end,
			},
		},
	}
}

func aggStagePreventNullArray(fields ...string) bson.M {
	targets := bson.M{}
	for _, field := range fields {
		targets[field] = bson.M{"$ifNull": bson.A{specifyField(field), bson.A{}}}
	}
	return bson.M{"$project": targets}
}

func specifyField(fieldName string) string {
	return fmt.Sprintf("$%s", fieldName)
}

func getStartTimeOfConsecutiveDays(now time.Time) (yesterdayStartAt time.Time, todayStartAt time.Time, tomorrowStartAt time.Time) {
	todayStartAt = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterdayStartAt = todayStartAt.AddDate(0, 0, -1)
	tomorrowStartAt = todayStartAt.AddDate(0, 0, 1)
	return
}
