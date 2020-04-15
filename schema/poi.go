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
}

type POIDesc struct {
	ID      primitive.ObjectID `bson:"_id"`
	Alias   string             `bson:"alias"`
	Address string             `bson:"address"`
}
