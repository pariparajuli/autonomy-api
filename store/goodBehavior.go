package store

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	DuplicateKeyCode = 11000
)

// GoodBehaviorReport save a GoodBehaviorData into Database
type GoodBehaviorReport interface {
	GoodBehaviorSave(data *schema.GoodBehaviorData) error
	NearestGoodBehaviorScore(distInMeter int, location schema.Location) (float64, float64, error)
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

// NearestGoodBehaviorScore return  the total behavior score and delta score of users within distInMeter range
func (m *mongoDB) NearestGoodBehaviorScore(distInMeter int, location schema.Location) (float64, float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := m.client.Database(m.database)
	collection := db.Collection(schema.GoodBehaviorCollection)
	todayBegin := todayInterval()
	log.Debugf("time period today > %v, yesterday %v~ %v ", todayBegin, todayBegin-86400, todayBegin)
	geoStage := bson.D{{"$geoNear", bson.M{
		"near":          bson.M{"type": "Point", "coordinates": bson.A{location.Longitude, location.Latitude}},
		"distanceField": "dist",
		"spherical":     true,
		"maxDistance":   distInMeter,
	}}}

	timeStageToday := bson.D{{"$match", bson.M{"ts": bson.M{"$gte": todayBegin}}}}
	timeStageYesterday := bson.D{{"$match", bson.M{"ts": bson.M{"$gte": todayBegin - 86400, "$lt": todayBegin}}}}

	sortStage := bson.D{{"$sort", bson.D{{"ts", -1}}}}

	groupStage := bson.D{
		{"$group", bson.D{
			{"_id", "$profile_id"},
			{"behavior_score", bson.D{
				{"$first", "$behavior_score"},
			}},
			{"account_number", bson.D{
				{"$first", "$profile_id"},
			}},
		}}}

	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{geoStage, timeStageToday, sortStage, groupStage})
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("aggregate nearest good behavior score")
		return 0, 0, err
	}
	sum := float64(0)
	count := 0
	for cursor.Next(ctx) {
		var result schema.GoodBehaviorData
		if err := cursor.Decode(&result); err != nil {
			log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("decode nearest good behavior score")
			continue
		}
		sum = sum + result.BehaviorScore
		count++
	}
	score := 100 * (sum / float64(count*schema.TotalGoodBehaviorWeight))

	// Previous day
	cursorYesterday, err := collection.Aggregate(ctx, mongo.Pipeline{geoStage, timeStageYesterday, sortStage, groupStage})
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("aggregate nearest good behavior score")
		return 0, 0, err
	}
	sumYesterday := float64(0)
	countYesterday := 0
	for cursorYesterday.Next(ctx) {
		var result schema.GoodBehaviorData
		if err := cursorYesterday.Decode(&result); err != nil {
			log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("decode yesterday nearest good behavior score")
			continue
		}
		sumYesterday = sumYesterday + result.BehaviorScore
		countYesterday++
	}
	scoreYesterday := float64(0)
	if countYesterday > 0 {
		scoreYesterday = 100 * (sumYesterday / float64(countYesterday*schema.TotalGoodBehaviorWeight))
	} else {
		scoreYesterday = 100
	}
	delta := score - scoreYesterday
	return score, delta, nil
}

func todayInterval() int64 {
	curTime := time.Now()
	start := time.Date(curTime.Year(), curTime.Month(), curTime.Day(), 0, 0, 0, 0, time.UTC)
	return start.Unix()
}
