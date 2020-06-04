package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFitnessPlace(t *testing.T) {
	types := []string{"establishment",
		"gym",
		"health",
		"point_of_interest",
	}

	placeType := ReadPlaceType(types)
	assert.Equal(t, FitnessPlace, placeType)
}

func TestHealthCarePlace(t *testing.T) {
	types := []string{"establishment",
		"establishment",
		"health",
		"hospital",
		"point_of_interest",
	}

	placeType := ReadPlaceType(types)
	assert.Equal(t, HealthCarePlace, placeType)
}

func TestHealthCarePlaceByDentist(t *testing.T) {
	types := []string{"establishment",
		"dentist",
		"establishment",
		"health",
		"point_of_interest",
	}

	placeType := ReadPlaceType(types)
	assert.Equal(t, HealthCarePlace, placeType)
}

func TestHealthCarePlaceOnlyHealth(t *testing.T) {
	types := []string{"establishment",
		"establishment",
		"health",
		"point_of_interest",
	}

	placeType := ReadPlaceType(types)
	assert.Equal(t, HealthCarePlace, placeType)
}

func TestFoodPlace(t *testing.T) {
	types := []string{"establishment",
		"establishment",
		"food",
		"point_of_interest",
		"restaurant",
	}

	placeType := ReadPlaceType(types)
	assert.Equal(t, FoodPlace, placeType)
}

func TestUnknownPlace(t *testing.T) {
	types := []string{"establishment",
		"establishment",
		"point_of_interest",
	}

	placeType := ReadPlaceType(types)
	assert.Equal(t, UnknownPlace, placeType)
}
