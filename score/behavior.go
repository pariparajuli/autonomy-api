package score

import (
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/store"
)

func behaviorScore(mongo *store.MongoStore, radiusMeter int, location schema.Location) (float64, float64, float64, float64, error) {
	rawData, err := (*mongo).NearestGoodBehavior(radiusMeter, location)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	// Score Rule:  Self defined weight can not exceed more than 1/2 of weight, if it exceeds 1/2 of weight, it counts as 1/2 of weight
	topScore := float64(rawData.TotalRecordCount)*schema.TotalOfficialBehaviorWeight + rawData.CustomerizedBehaviorWeight
	nearbyScore := rawData.OfficialBehaviorWeight + rawData.CustomerizedBehaviorWeight
	if portion := float64(rawData.CustomerizedBehaviorWeight) / float64(topScore); portion > 0.5 {
		nearbyScore = topScore/2 + rawData.OfficialBehaviorWeight
	}
	topScorePast := float64(rawData.PastTotalRecordCount)*schema.TotalOfficialBehaviorWeight + rawData.PastCustomerizedBehaviorWeight
	nearbyScorePast := rawData.PastOfficialBehaviorWeight + rawData.PastCustomerizedBehaviorWeight
	if portion := float64(rawData.CustomerizedBehaviorWeight) / float64(topScorePast); portion > 0.5 {
		nearbyScorePast = topScorePast/2 + rawData.PastOfficialBehaviorWeight
	}
	score := 100 * nearbyScore / topScore
	scorePast := 100 * nearbyScorePast / topScorePast
	deltaInPercent := (score - scorePast/score) * 100
	totalReportedCount := float64(rawData.OfficialBehaviorCount + rawData.CustomerizedBehaviorCount)
	deltaReportedCountPast := totalReportedCount - (float64(rawData.PastCustomerizedBehaviorCount + rawData.PastOfficialBehaviorCount))
	return score, deltaInPercent, totalReportedCount, deltaReportedCountPast, nil
}
