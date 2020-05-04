package score

import (
	"math"

	"github.com/bitmark-inc/autonomy-api/schema"
)

type NearestGoodBehaviorData struct {
	TotalCount       int32   `json:"totalCount" bson:"totalCount`
	OfficialWeight   float64 `json:"officialWeight" bson:"officialWeight"`
	OfficialCount    int32   `json:"officialCount" bson:"officialCount"`
	CustomizedWeight float64 `json:"customizedWeight" bson:"customizedWeight"`
	CustomizedCount  int32   `json:"customizedCount" bson:"customizedCount"`
}

func BehaviorScore(rawDataToday NearestGoodBehaviorData, rawDataYesterday NearestGoodBehaviorData) (float64, float64, float64, float64) {
	// Score Rule:  Self defined weight can not exceed more than 1/2 of weight, if it exceeds 1/2 of weight, it counts as 1/2 of weight
	topScore := float64(rawDataToday.TotalCount)*schema.TotalOfficialBehaviorWeight + rawDataToday.CustomizedWeight
	nearbyScore := rawDataToday.OfficialWeight + rawDataToday.CustomizedWeight
	topScorePast := float64(rawDataYesterday.TotalCount)*schema.TotalOfficialBehaviorWeight + rawDataYesterday.CustomizedWeight

	if topScore <= 0 && topScorePast <= 0 {
		return 0, 0, 0, 0
	}

	if topScore > 0 {
		if portion := float64(rawDataToday.CustomizedWeight) / float64(topScore); portion > 0.5 {
			nearbyScore = topScore/2 + rawDataToday.OfficialWeight
		}
	}
	score := float64(0)
	if topScore > 0 {
		score = 100 * nearbyScore / topScore
	}

	totalReportedCountPast := float64(rawDataYesterday.CustomizedCount + rawDataYesterday.OfficialCount)
	totalReportedCount := float64(rawDataToday.OfficialCount + rawDataToday.CustomizedCount)
	deltaCountInPercent := float64(100)
	if totalReportedCountPast > 0 {
		deltaCountInPercent = ((totalReportedCount - totalReportedCountPast) / totalReportedCountPast) * 100
	}

	return score, deltaCountInPercent, totalReportedCount, totalReportedCountPast
}
