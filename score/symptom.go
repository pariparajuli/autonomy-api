package score

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/bitmark-inc/autonomy-api/schema"
)

func CalculateSymptomScore(weights schema.SymptomWeights, metric schema.Metric) schema.Metric {
	today := metric.Details.Symptoms.TodayData
	yesterday := metric.Details.Symptoms.YesterdayData

	symptomScore, sTotalweight, sMaxScorePerPerson, sDeltaInPercent, sOfficialCount, sCustomizedCount :=
		SymptomScore(weights, today, yesterday)

	metric.SymptomDelta = sDeltaInPercent
	metric.SymptomCount = sOfficialCount + sCustomizedCount
	metric.Details.Symptoms = schema.SymptomDetail{
		SymptomTotal:       sTotalweight,
		TotalPeople:        metric.Details.Symptoms.TodayData.UserCount,
		Symptoms:           metric.Details.Symptoms.TodayData.WeightDistribution,
		MaxScorePerPerson:  sMaxScorePerPerson,
		CustomizedWeight:   sCustomizedCount,
		CustomSymptomCount: sCustomizedCount,
		Score:              symptomScore,
		TodayData:          metric.Details.Symptoms.TodayData,
	}

	return metric
}

func SymptomScore(weights schema.SymptomWeights, today, yesterday schema.NearestSymptomData) (float64, float64, float64, float64, float64, float64) {
	countYesterday := yesterday.OfficialCount + yesterday.CustomizedCount
	countToday := today.OfficialCount + today.CustomizedCount

	// Today
	maxScorePerPerson := float64(0) // Max Score,
	for _, v := range weights {
		maxScorePerPerson = maxScorePerPerson + v
	}

	if countYesterday <= 0 && countToday <= 0 {
		return 100, 0, maxScorePerPerson, 0, 0, 0
	}

	totalOfficialWeight := float64(0)
	var w float64
	var ok bool
	for k, v := range today.WeightDistribution {
		if w, ok = weights[k]; !ok {
			w = schema.DefaultSymptomWeights[k]
		}
		totalOfficialWeight += w * float64(v)
		log.Info(fmt.Sprintf("SymptomScore : k :%v w :%v , v:%v, symptomTotal:%v", k, w, v, totalOfficialWeight))
	}
	totalWeight := totalOfficialWeight + today.CustomizedCount*1
	score := float64(100)
	if today.OfficialCount*maxScorePerPerson > 0 {
		de := today.UserCount*maxScorePerPerson + today.CustomizedCount
		score = 100 * (1 - totalWeight/de)
	} else if today.CustomizedCount > 0 {
		score = 100 * (1 - totalWeight/today.CustomizedCount)
	}

	// deltaCount in percent
	deltaInPercent := float64(100)

	if countYesterday > 0 {
		deltaInPercent = (countToday - countYesterday) / countYesterday * 100
	}
	return score, totalWeight, maxScorePerPerson, deltaInPercent, today.OfficialCount, today.CustomizedCount * 1
}
