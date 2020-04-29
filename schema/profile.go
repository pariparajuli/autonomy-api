package schema

import "time"

const (
	ProfileCollection = "profile"
)

// SymptomWeights is structure for customized symptom weights
type SymptomWeights struct {
	BluishFace      int64 `json:"face" bson:"face"`
	BreathShortness int64 `json:"breath" bson:"breath"`
	ChestPain       int64 `json:"chest" bson:"chest"`
	DryCough        int64 `json:"cough" bson:"cough"`
	Fatigue         int64 `json:"fatigue" bson:"fatigue"`
	Fever           int64 `json:"fever" bson:"fever"`
	NasalCongestion int64 `json:"nasal" bson:"nasal"`
	SoreThroat      int64 `json:"throat" bson:"throat"`
}

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
	ID               string            `bson:"id"`
	AccountNumber    string            `bson:"account_number"`
	Location         *GeoJSON          `bson:"location,omitempty"`
	HealthScore      float64           `bson:"health_score"`
	Metric           Metric            `bson:"metric"`
	ScoreCoefficient *ScoreCoefficient `bson:"score_coefficient"`
	PointsOfInterest []ProfilePOI      `bson:"points_of_interest,omitempty"`
}

// GeoJSON - mongo location format
type GeoJSON struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}
