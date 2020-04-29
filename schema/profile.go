package schema

import "time"

const (
	ProfileCollection = "profile"
)

type ScoreCoefficient struct {
	Symptoms  float64   `json:"symptoms" bson:"symptoms"`
	Behaviors float64   `json:"behaviors" bson:"behaviors"`
	Confirms  float64   `json:"confirms" bson:"confirms"`
	UpdatedAt time.Time `json:"-" bson:"updated_at"`
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
