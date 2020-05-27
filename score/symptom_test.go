package score

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bitmark-inc/autonomy-api/schema"
)

func TestCalculateSymptomScoreUsingDefaultWeights(t *testing.T) {
	metric := schema.Metric{
		Details: schema.Details{
			Symptoms: schema.SymptomDetail{
				TotalPeople: 10,
				TodayData: schema.NearestSymptomData{
					WeightDistribution: map[string]int{
						"cough":       3, // weight 2
						"fever":       7, // weight 3
						"new-symptom": 1, // weight 1
					}},
				YesterdayData: schema.NearestSymptomData{
					WeightDistribution: map[string]int{
						"cough":       1,
						"fever":       1,
						"new-symptom": 2,
					}},
			},
		},
	}
	updatedMetric := CalculateSymptomScore(schema.DefaultSymptomWeights, metric)
	assert.Equal(t, "76.86", fmt.Sprintf("%.2f", updatedMetric.Details.Symptoms.Score))
	assert.Equal(t, 10.0, updatedMetric.Details.Symptoms.TotalPeople)
	assert.Equal(t, metric.Details.Symptoms.TodayData, updatedMetric.Details.Symptoms.TodayData)
	assert.Equal(t, metric.Details.Symptoms.YesterdayData, updatedMetric.Details.Symptoms.YesterdayData)
	assert.Equal(t, 11.0, updatedMetric.SymptomCount)
	assert.Equal(t, 175.0, updatedMetric.SymptomDelta)

	// the function must be idempotent
	updatedMetric = CalculateSymptomScore(schema.DefaultSymptomWeights, metric)
	assert.Equal(t, "76.86", fmt.Sprintf("%.2f", updatedMetric.Details.Symptoms.Score))
	assert.Equal(t, 10.0, updatedMetric.Details.Symptoms.TotalPeople)
	assert.Equal(t, metric.Details.Symptoms.TodayData, updatedMetric.Details.Symptoms.TodayData)
	assert.Equal(t, metric.Details.Symptoms.YesterdayData, updatedMetric.Details.Symptoms.YesterdayData)
	assert.Equal(t, 11.0, updatedMetric.SymptomCount)
	assert.Equal(t, 175.0, updatedMetric.SymptomDelta)
}

func TestCalculateSymptomScoreUsingCustomizedWeights(t *testing.T) {
	weights := schema.SymptomWeights{
		"cough":            1,
		"breath":           1,
		"fever":            1,
		"chills":           1,
		"muscle_pain":      1,
		"throat":           1,
		"loss_taste_smell": 1,
	}
	metric := schema.Metric{
		Details: schema.Details{
			Symptoms: schema.SymptomDetail{
				TotalPeople: 10,
				TodayData: schema.NearestSymptomData{
					WeightDistribution: map[string]int{
						"cough":       3, // weight 1
						"fever":       7, // weight 1
						"new-symptom": 2, // weight 1
					}},
				YesterdayData: schema.NearestSymptomData{
					WeightDistribution: map[string]int{
						"cough":       1,
						"fever":       1,
						"new-symptom": 2,
					}},
			},
		},
	}
	updatedMetric := CalculateSymptomScore(weights, metric)
	assert.Equal(t, "83.33", fmt.Sprintf("%.2f", updatedMetric.Details.Symptoms.Score))
	assert.Equal(t, 10.0, updatedMetric.Details.Symptoms.TotalPeople)
	assert.Equal(t, metric.Details.Symptoms.TodayData, updatedMetric.Details.Symptoms.TodayData)
	assert.Equal(t, metric.Details.Symptoms.YesterdayData, updatedMetric.Details.Symptoms.YesterdayData)
	assert.Equal(t, 12.0, updatedMetric.SymptomCount)
	assert.Equal(t, 200.0, updatedMetric.SymptomDelta)
}

func TestCalculateSymptomScoreNoReportYesterday(t *testing.T) {
	metric := schema.Metric{
		Details: schema.Details{
			Symptoms: schema.SymptomDetail{
				TotalPeople: 5,
				TodayData: schema.NearestSymptomData{
					WeightDistribution: schema.SymptomDistribution{
						"fever":         10, // weight 3
						"new-symptom-1": 1,  // weight 1
						"new-symptom-2": 1,  // weight 1
					}},
			},
		},
	}
	updatedMetric := CalculateSymptomScore(schema.DefaultSymptomWeights, metric)
	assert.Equal(t, "48.39", fmt.Sprintf("%.2f", updatedMetric.Details.Symptoms.Score))
	assert.Equal(t, 5.0, updatedMetric.Details.Symptoms.TotalPeople)
	assert.Equal(t, metric.Details.Symptoms.TodayData, updatedMetric.Details.Symptoms.TodayData)
	assert.Equal(t, metric.Details.Symptoms.YesterdayData, updatedMetric.Details.Symptoms.YesterdayData)
	assert.Equal(t, 12.0, updatedMetric.SymptomCount)
	assert.Equal(t, 100.0, updatedMetric.SymptomDelta)
}

func TestCalculateSymptomScoreNoReportToday(t *testing.T) {
	metric := schema.Metric{
		Details: schema.Details{
			Symptoms: schema.SymptomDetail{
				TotalPeople: 5,
				YesterdayData: schema.NearestSymptomData{
					WeightDistribution: schema.SymptomDistribution{
						"nasal":         10, // weight 1
						"new-symptom-1": 1,  // weight 1
						"new-symptom-2": 1,  // weight 1
					}},
			},
		},
	}
	updatedMetric := CalculateSymptomScore(schema.DefaultSymptomWeights, metric)
	assert.Equal(t, 100.0, updatedMetric.Details.Symptoms.Score)
	assert.Equal(t, 5.0, updatedMetric.Details.Symptoms.TotalPeople)
	assert.Equal(t, metric.Details.Symptoms.TodayData, updatedMetric.Details.Symptoms.TodayData)
	assert.Equal(t, metric.Details.Symptoms.YesterdayData, updatedMetric.Details.Symptoms.YesterdayData)
	assert.Equal(t, 0.0, updatedMetric.SymptomCount)
	assert.Equal(t, -100.0, updatedMetric.SymptomDelta)
}
