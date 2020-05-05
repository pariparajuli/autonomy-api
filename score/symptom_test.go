package score_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/score"
)

func TestSymptomScoreWhenDoTwice(t *testing.T) {
	yesterday := schema.NearestSymptomData{
		UserCount:       1,
		OfficialCount:   2,
		CustomizedCount: 3,
		WeightDistribution: map[schema.SymptomType]int{
			schema.Fever:   1,
			schema.Face:    1,
			schema.Chest:   1,
			schema.Throat:  1,
			schema.Nasal:   1,
			schema.Cough:   1,
			schema.Fatigue: 1,
			schema.Breath:  1,
		},
	}

	today := schema.NearestSymptomData{
		UserCount:       2,
		OfficialCount:   3,
		CustomizedCount: 4,
		WeightDistribution: map[schema.SymptomType]int{
			schema.Fever:   2,
			schema.Face:    2,
			schema.Chest:   2,
			schema.Throat:  3,
			schema.Nasal:   3,
			schema.Cough:   3,
			schema.Fatigue: 4,
			schema.Breath:  5,
		},
	}

	y1, y2, y3, y4, y5, y6 := score.SymptomScore(schema.DefaultSymptomWeights, today, yesterday)
	t1, t2, t3, t4, t5, t6 := score.SymptomScore(schema.DefaultSymptomWeights, today, yesterday)

	assert.Equal(t, t1, y1, "wrong symptom score")
	assert.Equal(t, t2, y2, "wrong total weight")
	assert.Equal(t, t3, y3, "wrong max score per person")
	assert.Equal(t, t4, y4, "wrong delta in percent")
	assert.Equal(t, t5, y5, "wrong official count")
	assert.Equal(t, t6, y6, "wrong customized count")
}
