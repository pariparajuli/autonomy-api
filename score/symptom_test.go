package score

import (
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
	assert.Equal(t, 80.0, updatedMetric.Details.Symptoms.Score)
	assert.Equal(t, 28.0, updatedMetric.Details.Symptoms.SymptomTotal)
	assert.Equal(t, 10.0, updatedMetric.Details.Symptoms.TotalPeople)
	assert.Equal(t, 13.0, updatedMetric.Details.Symptoms.MaxScorePerPerson)
	assert.Equal(t, 1.0, updatedMetric.Details.Symptoms.CustomizedWeight)
	assert.Equal(t, metric.Details.Symptoms.TodayData, updatedMetric.Details.Symptoms.TodayData)
	assert.Equal(t, metric.Details.Symptoms.YesterdayData, updatedMetric.Details.Symptoms.YesterdayData)
	assert.Equal(t, 11.0, updatedMetric.SymptomCount)
	assert.Equal(t, 175.0, updatedMetric.SymptomDelta)

	// the function must be idempotent
	updatedMetric = CalculateSymptomScore(schema.DefaultSymptomWeights, metric)
	assert.Equal(t, 80.0, updatedMetric.Details.Symptoms.Score)
	assert.Equal(t, 28.0, updatedMetric.Details.Symptoms.SymptomTotal)
	assert.Equal(t, 10.0, updatedMetric.Details.Symptoms.TotalPeople)
	assert.Equal(t, 13.0, updatedMetric.Details.Symptoms.MaxScorePerPerson)
	assert.Equal(t, 1.0, updatedMetric.Details.Symptoms.CustomizedWeight)
	assert.Equal(t, metric.Details.Symptoms.TodayData, updatedMetric.Details.Symptoms.TodayData)
	assert.Equal(t, metric.Details.Symptoms.YesterdayData, updatedMetric.Details.Symptoms.YesterdayData)
	assert.Equal(t, 11.0, updatedMetric.SymptomCount)
	assert.Equal(t, 175.0, updatedMetric.SymptomDelta)
}

func TestCalculateSymptomScoreUsingCustomizedWeights(t *testing.T) {
	weights := schema.SymptomWeights{
		schema.Fever:   1,
		schema.Cough:   1,
		schema.Fatigue: 1,
		schema.Breath:  1,
		schema.Nasal:   1,
		schema.Throat:  1,
		schema.Chest:   1,
		schema.Face:    1,
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
	assert.Equal(t, 88.0, updatedMetric.Details.Symptoms.Score)
	assert.Equal(t, 12.0, updatedMetric.Details.Symptoms.SymptomTotal)
	assert.Equal(t, 10.0, updatedMetric.Details.Symptoms.TotalPeople)
	assert.Equal(t, 8.0, updatedMetric.Details.Symptoms.MaxScorePerPerson)
	assert.Equal(t, 2.0, updatedMetric.Details.Symptoms.CustomizedWeight)
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
						"nasal":         10, // weight 1
						"new-symptom-1": 1,  // weight 1
						"new-symptom-2": 1,  // weight 1
					}},
			},
		},
	}
	updatedMetric := CalculateSymptomScore(schema.DefaultSymptomWeights, metric)
	assert.Equal(t, 84.0, updatedMetric.Details.Symptoms.Score)
	assert.Equal(t, 12.0, updatedMetric.Details.Symptoms.SymptomTotal)
	assert.Equal(t, 5.0, updatedMetric.Details.Symptoms.TotalPeople)
	assert.Equal(t, 13.0, updatedMetric.Details.Symptoms.MaxScorePerPerson)
	assert.Equal(t, 2.0, updatedMetric.Details.Symptoms.CustomizedWeight)
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
	assert.Equal(t, 0.0, updatedMetric.Details.Symptoms.SymptomTotal)
	assert.Equal(t, 5.0, updatedMetric.Details.Symptoms.TotalPeople)
	assert.Equal(t, 13.0, updatedMetric.Details.Symptoms.MaxScorePerPerson)
	assert.Equal(t, 0.0, updatedMetric.Details.Symptoms.CustomizedWeight)
	assert.Equal(t, metric.Details.Symptoms.TodayData, updatedMetric.Details.Symptoms.TodayData)
	assert.Equal(t, metric.Details.Symptoms.YesterdayData, updatedMetric.Details.Symptoms.YesterdayData)
	assert.Equal(t, 0.0, updatedMetric.SymptomCount)
	assert.Equal(t, -100.0, updatedMetric.SymptomDelta)
}
