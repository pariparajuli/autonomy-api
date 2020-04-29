package schema

import (
	"time"
)

const (
	ProfileCollection = "profile"
)

// SymptomWeights is structure for customized symptom weights
type SymptomWeights map[SymptomType]float64

var (
	DefaultSymptomWeights = SymptomWeights{
		Fever:   3,
		Cough:   2,
		Fatigue: 1,
		Breath:  1,
		Nasal:   1,
		Throat:  1,
		Chest:   2,
		Face:    2,
	}
)

// ScoreCoefficient is structure for all customized weights for calculating personal score
type ScoreCoefficient struct {
	Symptoms  float64   `json:"symptoms" bson:"symptoms"`
	Behaviors float64   `json:"behaviors" bson:"behaviors"`
	Confirms  float64   `json:"confirms" bson:"confirms"`
	UpdatedAt time.Time `json:"-" bson:"updated_at"`

	SymptomWeights SymptomWeights `json:"symptom_weights" bson:"symptom_weights"`
}

// Profile - user profile data
type Profile struct {
	ID                  string            `bson:"id"`
	AccountNumber       string            `bson:"account_number"`
	Location            *GeoJSON          `bson:"location,omitempty"`
	HealthScore         float64           `bson:"health_score"`
	Metric              Metric            `bson:"metric"`
	ScoreCoefficient    *ScoreCoefficient `bson:"score_coefficient"`
	PointsOfInterest    []ProfilePOI      `bson:"points_of_interest,omitempty"`
	CustomizedBehaviors []Behavior        `bson:"customized_behavior"`
	CustomizedSymptoms  []Symptom         `bson:"customized_symptom"`
}

// GeoJSON - mongo location format
type GeoJSON struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}
