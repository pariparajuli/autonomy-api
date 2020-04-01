package schema

const (
	ProfileCollectionName = "profile"
)

// Profile - user profile data
type Profile struct {
	ID            string   `bson:"id"`
	AccountNumber string   `bson:"account_number"`
	Location      *GeoJSON `bson:"location,omitempty"`
	HealthScore   float64  `bson:"health_score"`
}

// GeoJSON - mongo location format
type GeoJSON struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}
