package score

import (
	"github.com/bitmark-inc/autonomy-api/schema"
)

type NearestGoodBehaviorData struct {
	TotalRecordCount             int32
	OfficialBehaviorWeight       float64
	OfficialBehaviorCount        int32
	CustomizedBehaviorWeight     float64
	CustomizedBehaviorCount      int32
	PastTotalRecordCount         int32
	PastOfficialBehaviorWeight   float64
	PastOfficialBehaviorCount    int32
	PastCustomizedBehaviorWeight float64
	PastCustomizedBehaviorCount  int32
}

func BehaviorScore(rawData NearestGoodBehaviorData) (float64, float64, float64, float64) {
	// Score Rule:  Self defined weight can not exceed more than 1/2 of weight, if it exceeds 1/2 of weight, it counts as 1/2 of weight
	topScore := float64(rawData.TotalRecordCount)*schema.TotalOfficialBehaviorWeight + rawData.CustomizedBehaviorWeight
	nearbyScore := rawData.OfficialBehaviorWeight + rawData.CustomizedBehaviorWeight
	topScorePast := float64(rawData.PastTotalRecordCount)*schema.TotalOfficialBehaviorWeight + rawData.PastCustomizedBehaviorWeight
	nearbyScorePast := rawData.PastOfficialBehaviorWeight + rawData.PastCustomizedBehaviorWeight
	if topScore <= 0 && topScorePast <= 0 {
		return 0, 0, 0, 0
	}

	if topScore > 0 {
		if portion := float64(rawData.CustomizedBehaviorWeight) / float64(topScore); portion > 0.5 {
			nearbyScore = topScore/2 + rawData.OfficialBehaviorWeight
		}
	}
	if topScorePast > 0 {
		if portion := float64(rawData.CustomizedBehaviorWeight) / float64(topScorePast); portion > 0.5 {
			nearbyScorePast = topScorePast/2 + rawData.PastOfficialBehaviorWeight
		}
	}
	score := float64(0)
	if score > 0 {
		score = 100 * nearbyScore / topScore
	}

	scorePast := float64(0)
	deltaInPercent := float64(100)
	if topScorePast > 0 {
		scorePast = 100 * nearbyScorePast / topScorePast
		deltaInPercent = ((score - scorePast) / scorePast) * 100
	}

	totalReportedCount := float64(rawData.OfficialBehaviorCount + rawData.CustomizedBehaviorCount)
	deltaReportedCountPast := totalReportedCount - (float64(rawData.PastCustomizedBehaviorCount + rawData.PastOfficialBehaviorCount))
	return score, deltaInPercent, totalReportedCount, deltaReportedCountPast
}
