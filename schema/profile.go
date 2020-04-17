package schema

const (
	ProfileCollection = "profile"
)

// Profile - user profile data
type Profile struct {
	ID               string     `bson:"id"`
	AccountNumber    string     `bson:"account_number"`
	Location         *GeoJSON   `bson:"location,omitempty"`
	HealthScore      float64    `bson:"health_score"`
	Metric           Metric     `bson:"metric"`
	PointsOfInterest *[]POIDesc `bson:"points_of_interest,omitempty"`
}

// GeoJSON - mongo location format
type GeoJSON struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}
