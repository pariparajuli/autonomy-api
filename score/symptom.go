package score

import (
	"time"

	"github.com/bitmark-inc/autonomy-api/schema"
)

func CalculateSymptomScore(weights schema.SymptomWeights, metric schema.Metric) schema.Metric {
	rawData := metric.Details.Symptoms

	totalWeight := float64(0)
	for _, w := range weights {
		totalWeight += w
	}

	weightedSum := float64(0)
	officialCount := 0
	nonOfficialCount := 0
	for symptomID, cnt := range rawData.TodayData.WeightDistribution {
		var weight float64
		var ok bool
		if schema.OfficialSymptoms[symptomID] {
			if weight, ok = weights[symptomID]; !ok {
				weight = schema.DefaultSymptomWeights[symptomID]
			}
			officialCount += cnt
		} else {
			weight = 1
			nonOfficialCount += cnt
		}

		weightedSum += float64(cnt) * weight
	}

	totalCountToday := officialCount + nonOfficialCount
	totalCountYesterday := 0
	for _, cnt := range rawData.YesterdayData.WeightDistribution {
		totalCountYesterday += cnt
	}

	maxWeightedSum := float64(rawData.TotalPeople)*totalWeight + float64(nonOfficialCount)
	score := 100.0
	if maxWeightedSum > 0 {
		score = 100 * (1 - weightedSum/maxWeightedSum)
	}

	spikeList := CheckSymptomSpike(rawData.YesterdayData.WeightDistribution, rawData.TodayData.WeightDistribution)

	metric.SymptomCount = float64(totalCountToday)
	metric.SymptomDelta = ChangeRate(float64(totalCountToday), float64(totalCountYesterday))
	metric.Details.Symptoms = schema.SymptomDetail{
		Score:             score,
		SymptomTotal:      weightedSum,
		MaxScorePerPerson: totalWeight,
		TotalPeople:       float64(rawData.TotalPeople),
		CustomizedWeight:  float64(nonOfficialCount),
		TodayData:         metric.Details.Symptoms.TodayData,
		YesterdayData:     metric.Details.Symptoms.YesterdayData,
		LastSpikeList:     spikeList,
		LastSpikeUpdate:   time.Now().UTC(),
	}

	return metric
}
