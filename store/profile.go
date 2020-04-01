package store

import "github.com/bitmark-inc/autonomy-api/schema"

const (
	ProfileCollectionName = "profile"
)

// Profile - user profile data
type Profile struct {
	ID            string           `bson:"id"`
	AccountNumber string           `bson:"account_number"`
	Location      *schema.Location `bson:"location,omitempty"`
	HealthScore   float64          `bson:"health_score"`
}
