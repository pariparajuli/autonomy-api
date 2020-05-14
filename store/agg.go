package store

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/bitmark-inc/autonomy-api/schema"
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

func aggStagePreventNullArray(fields ...string) bson.M {
	targets := bson.M{}
	for _, field := range fields {
		targets[field] = bson.M{"$ifNull": bson.A{specifyField(field), bson.A{}}}
	}
	return bson.M{"$project": targets}
}

// aggStageReportedItemCount counts the length of items of each report type and adds them up.
/*
{
	"$project": {
		count: {
			$sum: {
				$add: [
					{$size: $addendFields[0]},
					{$size: $addendFields[1]},
					...
				]
			}
		}
	}
}
*/
func aggStageReportedItemCount(resultField string, addendFields ...string) bson.M {
	terms := bson.A{}
	for _, field := range addendFields {
		terms = append(terms, bson.M{"$size": specifyField(field)})
	}
	return bson.M{
		"$project": bson.M{
			resultField: bson.M{"$sum": bson.M{"$add": terms}},
		},
	}
}

// aggStageSumValues sums up the values of the specified field.
/*
{
	_id: null,
	resultField: {
		$sum: "$targetField"
	}
}
*/
func aggStageSumValues(targetField, resultField string) bson.M {
	return bson.M{
		"$group": bson.M{
			"_id": nil,
			resultField: bson.M{
				"$sum": specifyField(targetField),
			},
		},
	}
}

func specifyField(fieldName string) string {
	return fmt.Sprintf("$%s", fieldName)
}
