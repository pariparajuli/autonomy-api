package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitBehaviors(t *testing.T) {
	behaviors := []Behavior{
		{ID: "touch_face"},
		{ID: "something new"},
		{ID: "social_distancing"},
		{ID: "something brand new"},
	}
	official, nonOfficial := SplitBehaviors(behaviors)
	assert.Equal(t, 2, len(official))
	assert.Equal(t, "touch_face", string(official[0].ID))
	assert.Equal(t, "social_distancing", string(official[1].ID))

	assert.Equal(t, 2, len(nonOfficial))
	assert.Equal(t, "something new", string(nonOfficial[0].ID))
	assert.Equal(t, "something brand new", string(nonOfficial[1].ID))
}
