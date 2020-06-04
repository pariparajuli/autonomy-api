package utils

const (
	FoodPlace       = "food"
	HealthCarePlace = "healthcare"
	FitnessPlace    = "fitness"
	UnknownPlace    = "unknown"
)

// ReadPlaceType returns a place type by analyzing a list of given types
func ReadPlaceType(types []string) string {
	health := false
	for _, t := range types {
		switch t {
		case "health":
			health = true
		case "gym":
			return FitnessPlace
		case "restaurant", "food", "cafe":
			return FoodPlace
		case "hospital", "doctor", "dentist", "pharmacy":
			return HealthCarePlace
		}
	}

	if health {
		return HealthCarePlace
	}
	return UnknownPlace
}
