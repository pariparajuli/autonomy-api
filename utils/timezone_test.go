package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocation(t *testing.T) {
	tz8 := GetLocation("GMT+8")
	assert.NotNil(t, tz8)
	assert.Equal(t, "GMT+8", tz8.String())

	tz1245 := GetLocation("GMT+12:45")
	assert.NotNil(t, tz1245)
	assert.Equal(t, "GMT+12:45", tz1245.String())

	tz945 := GetLocation("GMT+9:45")
	assert.NotNil(t, tz945)
	assert.Equal(t, "GMT+9:45", tz945.String())

	tz_8 := GetLocation("GMT-8")
	assert.NotNil(t, tz_8)
	assert.Equal(t, "GMT-8", tz_8.String())

	tz_1245 := GetLocation("GMT-12:45")
	assert.NotNil(t, tz_1245)
	assert.Equal(t, "GMT-12:45", tz_1245.String())

	tz_945 := GetLocation("GMT-9:45")
	assert.NotNil(t, tz_945)
	assert.Equal(t, "GMT-9:45", tz_945.String())
}
