package store

const (
	ProfileCollectionName = "profile"
)

// Profile - user profile data
type Profile struct {
	ID            string         `bson:"id"`
	AccountNumber string         `bson:"account_number"`
	Location      *MongoLocation `bson:"location,omitempty"`
	HealthScore   float64        `bson:"health_score"`
}

// mongo location format
type MongoLocation struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}
