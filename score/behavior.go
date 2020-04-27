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
	topScore := float64(rawData.TotalRecordCount)*schema.TotalDefaultBehaviorWeight + rawData.SelfDefinedBehaviorWeight
	nearbyScore := rawData.DefaultBehaviorWeight + rawData.SelfDefinedBehaviorWeight
	if portion := float64(rawData.SelfDefinedBehaviorWeight) / float64(topScore); portion > 0.5 {
		nearbyScore = topScore/2 + rawData.DefaultBehaviorWeight
	}
	topScorePast := float64(rawData.PastTotalRecordCount)*schema.TotalDefaultBehaviorWeight + rawData.PastSelfDefinedBehaviorWeight
	nearbyScorePast := rawData.PastDefaultBehaviorWeight + rawData.PastSelfDefinedBehaviorWeight
	if portion := float64(rawData.SelfDefinedBehaviorWeight) / float64(topScorePast); portion > 0.5 {
		nearbyScorePast = topScorePast/2 + rawData.PastDefaultBehaviorWeight
	}
	score := 100 * nearbyScore / topScore
	scorePast := 100 * nearbyScorePast / topScorePast
	deltaInPercent := (score - scorePast/score) * 100
	totalReportedCount := float64(rawData.DefaultBehaviorCount + rawData.SelfDefinedBehaviorCount)
	deltaReportedCountPast := totalReportedCount - (float64(rawData.PastSelfDefinedBehaviorCount + rawData.PastDefaultBehaviorCount))
	return score, deltaInPercent, totalReportedCount, deltaReportedCountPast, nil
}
