package schema

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	POICollection = "poi"
)

type POI struct {
	ID       primitive.ObjectID `bson:"_id"`
	Location *GeoJSON           `bson:"location"`
	Score    float64            `bson:"score"`
	Metric   Metric             `bson:"metric"`
}

type ProfilePOI struct {
	ID      primitive.ObjectID `bson:"id" json:"id"`
	Alias   string             `bson:"alias" json:"alias"`
	Address string             `bson:"address" json:"address"`
	Score   float64            `bson:"score" json:"score"`
}

// This structure will not store into database, it's only for client response.
// Data comes from schema Profile.PointsOfInterest & POI
type POIDetail struct {
	ProfilePOI `bson:",inline"`
	Location   *Location `json:"location"`
}
