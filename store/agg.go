package store

import (
	"github.com/bitmark-inc/autonomy-api/schema"
	"go.mongodb.org/mongo-driver/bson"
)

func matchReportedToday(todayBeginTime int64) bson.M {
	return bson.M{
		"ts": bson.M{
			"$gte": todayBeginTime,
		},
	}
}

func matchReportedYesterday(todayBeginTime int64) bson.M {
	return bson.M{
		"ts": bson.M{
			"$gte": todayBeginTime - 24*60*60,
			"$lt":  todayBeginTime,
		},
	}
}

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

func aggStageReportedToday(todayBeginTime int64) bson.M {
	return bson.M{
		"$match": matchReportedToday(todayBeginTime),
	}
}

func aggStageReportedYesterday(todayBeginTime int64) bson.M {
	return bson.M{
		"$match": matchReportedYesterday(todayBeginTime),
	}
}

func aggStageUserReportCount() bson.M {
	return bson.M{
		"$group": bson.M{
			"_id": "$account_number",
			"count": bson.M{
				"$sum": 1,
			},
		},
	}
}
