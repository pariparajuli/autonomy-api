package score

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorChangeFrom0To1(t *testing.T) {
	assert.False(t, CheckScoreColorChange(0, 1))
}

func TestColorChangeFrom33To34(t *testing.T) {
	assert.True(t, CheckScoreColorChange(33, 34))
}

func TestColorChangeFrom66To67(t *testing.T) {
	assert.True(t, CheckScoreColorChange(66, 67))
}

func TestColorChangeFrom99To100(t *testing.T) {
	assert.False(t, CheckScoreColorChange(99, 100))
}
