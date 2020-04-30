package score_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/score"
)

func TestSymptomScore(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	users := float64(3)
	count := map[schema.SymptomType]int{
		schema.Fever:   1,
		schema.Face:    1,
		schema.Chest:   1,
		schema.Throat:  1,
		schema.Nasal:   1,
		schema.Cough:   1,
		schema.Fatigue: 1,
		schema.Breath:  1,
	}
	metric := schema.Metric{
		Details: schema.Details{
			Symptoms: schema.SymptomDetail{
				TotalPeople: users,
				Symptoms:    count,
			},
		},
	}

	score.SymptomScore(schema.DefaultSymptomWeights, &metric, nil)

	var expectedTotal float64
	var defaultMaxWeight float64
	for k, v := range count {
		expectedTotal += float64(v) * schema.DefaultSymptomWeights[k]
		defaultMaxWeight += schema.DefaultSymptomWeights[k]
	}

	assert.Equal(t, expectedTotal, metric.Details.Symptoms.SymptomTotal, "wrong symptom total")
	assert.Equal(t, users, metric.Details.Symptoms.TotalPeople, "wrong user")
	assert.Equal(t, defaultMaxWeight, metric.Details.Symptoms.MaxScorePerPerson, "wrong max per person")

	assert.Equal(t, 100-100*(expectedTotal/(users*defaultMaxWeight)), metric.Details.Symptoms.Score, "wrong score")
}
