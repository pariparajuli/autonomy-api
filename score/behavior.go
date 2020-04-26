package score

import (
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/store"
)

func behaviorScore(mongo *store.MongoStore, radiusMeter int, location schema.Location) (float64, float64, int, int, error) {
	dWeight, sWeight, count, dWeightPast, sWeightPast, countPast, err := (*mongo).NearestGoodBehaviorScore(radiusMeter, location)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	countDelta := count - countPast
	// Score Rule:  Self defined weight can not exceed more than 1/2 of weight, if it exceeds 1/2 of weight, it counts as 1/2 of weight
	if sWeight > (float64(count) * schema.TotalGoodBehaviorWeight / 2) {
		sWeight = float64(count) * schema.TotalGoodBehaviorWeight / 2
	}
	score := float64(0)
	if score := 100 * (dWeight + sWeight) / (float64(count) * schema.TotalGoodBehaviorWeight); score > 100 {
		score = 100
	}
	scorePast := float64(0)
	if scorePast := 100 * (dWeightPast + sWeightPast) / (float64(countPast) * schema.TotalGoodBehaviorWeight); scorePast > 100 {
		scorePast = 100
	}
	return score, scorePast, count, countDelta, nil
}
