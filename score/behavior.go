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

func BehaviorScore(rawData NearestGoodBehaviorData) (float64, float64, float64, float64, error) {
	// Score Rule:  Self defined weight can not exceed more than 1/2 of weight, if it exceeds 1/2 of weight, it counts as 1/2 of weight
	topScore := float64(rawData.TotalRecordCount)*schema.TotalOfficialBehaviorWeight + rawData.CustomizedBehaviorWeight
	nearbyScore := rawData.OfficialBehaviorWeight + rawData.CustomizedBehaviorWeight
	if portion := float64(rawData.CustomizedBehaviorWeight) / float64(topScore); portion > 0.5 {
		nearbyScore = topScore/2 + rawData.OfficialBehaviorWeight
	}
	topScorePast := float64(rawData.PastTotalRecordCount)*schema.TotalOfficialBehaviorWeight + rawData.PastCustomizedBehaviorWeight
	nearbyScorePast := rawData.PastOfficialBehaviorWeight + rawData.PastCustomizedBehaviorWeight
	if portion := float64(rawData.CustomizedBehaviorWeight) / float64(topScorePast); portion > 0.5 {
		nearbyScorePast = topScorePast/2 + rawData.PastOfficialBehaviorWeight
	}
	score := 100 * nearbyScore / topScore
	scorePast := 100 * nearbyScorePast / topScorePast
	deltaInPercent := ((score - scorePast) / scorePast) * 100
	totalReportedCount := float64(rawData.OfficialBehaviorCount + rawData.CustomizedBehaviorCount)
	deltaReportedCountPast := totalReportedCount - (float64(rawData.PastCustomizedBehaviorCount + rawData.PastOfficialBehaviorCount))
	return score, deltaInPercent, totalReportedCount, deltaReportedCountPast, nil
}
