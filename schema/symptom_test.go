package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitSymptoms(t *testing.T) {
	symptoms := []Symptom{
		{ID: "dry cough"},
		{ID: "cough"},
		{ID: "something new"},
	}
	official, nonOfficial := SplitSymptoms(symptoms)
	assert.Equal(t, 1, len(official))
	assert.Equal(t, "cough", official[0].ID)

	assert.Equal(t, 2, len(nonOfficial))
	assert.Equal(t, "dry cough", nonOfficial[0].ID)
	assert.Equal(t, "something new", nonOfficial[1].ID)
}
