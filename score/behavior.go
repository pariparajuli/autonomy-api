package score

import (
	"math"

	"github.com/bitmark-inc/autonomy-api/schema"
)

func UpdateBehaviorMetrics(metric *schema.Metric) {
	todayTotal := 0
	yesterdayTotal := 0
	officialWeightedSum := float64(0)
	nonOfficialWeightedSum := float64(0)

	for behaviorID, cnt := range metric.Details.Behaviors.TodayDistribution {
		w, ok := schema.DefaultBehaviorWeightMatrix[schema.GoodBehaviorType(behaviorID)]
		if ok {
			officialWeightedSum += w.Weight * float64(cnt)
		} else {
			nonOfficialWeightedSum += float64(cnt)
		}

		todayTotal += cnt
	}
	for _, cnt := range metric.Details.Behaviors.YesterdayDistribution {
		yesterdayTotal += cnt
	}

	maxWeightedSum := float64(metric.Details.Behaviors.ReportTimes)*schema.TotalOfficialBehaviorWeight + nonOfficialWeightedSum
	// cap weighted sum of non-official behaviors
	nonOfficialWeightedSum = math.Min(nonOfficialWeightedSum, maxWeightedSum/2)
	weightedSum := officialWeightedSum + nonOfficialWeightedSum
	if maxWeightedSum > 0 {
		metric.Details.Behaviors.Score = 100 * weightedSum / maxWeightedSum
	}

	metric.BehaviorCount = float64(todayTotal)
	metric.BehaviorDelta = ChangeRate(float64(todayTotal), float64(yesterdayTotal))
}
