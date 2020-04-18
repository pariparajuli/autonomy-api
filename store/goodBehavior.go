package store

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bitmark-inc/autonomy-api/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DuplicateKeyCode = 11000
)

// GoodBehaviorReport save a GoodBehaviorData into Database
type GoodBehaviorReport interface {
	GoodBehaviorSave(data *schema.GoodBehaviorData) error
	NearestGoodBehaviorScore(distInMeter int, location schema.Location) (float64, int, error)
}

// GoodBehaviorData save a GoodBehaviorData into mongoDB
func (m *mongoDB) GoodBehaviorSave(data *schema.GoodBehaviorData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c := m.client.Database(m.database)
	_, err := c.Collection(schema.GoodBehaviorCollection).InsertOne(ctx, *data)
	we, hasErr := err.(mongo.WriteException)
	if hasErr {
		if 1 == len(we.WriteErrors) && DuplicateKeyCode == we.WriteErrors[0].Code {
			return nil
		}
		return err
	}
	return nil
}

// NearestGoodBehaviorScore return  the total behavior score (weight) and number of recorders of users within distInMeter range
func (m *mongoDB) NearestGoodBehaviorScore(distInMeter int, location schema.Location) (float64, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := m.client.Database(m.database)
	collection := db.Collection(schema.GoodBehaviorCollection)
	geoStage := bson.D{{"$geoNear", bson.M{
		"near":          bson.M{"type": "Point", "coordinates": bson.A{location.Longitude, location.Latitude}},
		"distanceField": "dist",
		"spherical":     true,
		"maxDistance":   distInMeter,
	}}}
	sortStage := bson.D{{"$sort", bson.D{{"ts", -1}}}}
	groupStage := bson.D{
		{"$group", bson.D{
			{"_id", "$profile_id"},
			{"behavior_score", bson.D{
				{"$first", "$behavior_score"},
			}},
		}}}
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{geoStage, sortStage, groupStage})
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("aggregate nearest good behavior score")
		return 0, 0, err
	}
	sum := float64(0)
	count := 0
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("decode nearest good behavior score")
			continue
		}
		sum = sum + result["behavior_score"].(float64)
		count++
	}
	return sum, count, nil
}
