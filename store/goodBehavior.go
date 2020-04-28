package store

import (
	"context"
	"errors"
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
	GoodBehaviorSave(data *schema.BehaviorReportData) error
	NearestGoodBehavior(distInMeter int, location schema.Location) (NearestGoodBehaviorData, error)
}

type NearestGoodBehaviorData struct {
	TotalRecordCount              int32
	DefaultBehaviorWeight         float64
	DefaultBehaviorCount          int32
	SelfDefinedBehaviorWeight     float64
	SelfDefinedBehaviorCount      int32
	PastTotalRecordCount          int32
	PastDefaultBehaviorWeight     float64
	PastDefaultBehaviorCount      int32
	PastSelfDefinedBehaviorWeight float64
	PastSelfDefinedBehaviorCount  int32
}

// GoodBehaviorData save a GoodBehaviorData into mongoDB
func (m *mongoDB) GoodBehaviorSave(data *schema.BehaviorReportData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c := m.client.Database(m.database)
	_, err := c.Collection(schema.BehaviorReportCollection).InsertOne(ctx, *data)
	we, hasErr := err.(mongo.WriteException)
	if hasErr {
		if 1 == len(we.WriteErrors) && DuplicateKeyCode == we.WriteErrors[0].Code {
			return nil
		}
		return err
	}
	return nil
}

// NearestGoodBehavior returns
// default behavior weight and count, self-defined-behavior weight and count, total number of records and error
func (m *mongoDB) NearestGoodBehavior(distInMeter int, location schema.Location) (NearestGoodBehaviorData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := m.client.Database(m.database)
	collection := db.Collection(schema.BehaviorReportCollection)
	todayBegin := todayStartAt()
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
	groupStage := bson.D{{"$group", bson.D{
		{"_id", "$profile_id"},
		{"account_number", bson.D{{"$first", "$account_number"}}},
		{"default_count", bson.D{{"$first", bson.D{{"$size", "$default_behaviors"}}}}},
		{"self_count", bson.D{{"$first", bson.D{{"$size", "$self_defined_behaviors"}}}}},
		{"default_weight", bson.D{{"$first", "$default_weight"}}},
		{"self_defined_weight", bson.D{{"$first", "$self_defined_weight"}}},
	}}}
	groupMergeStage := bson.D{{"$group", bson.D{
		{"_id", 1},
		{"totalDCount", bson.D{{"$sum", "$default_count"}}},
		{"totalSCount", bson.D{{"$sum", "$self_count"}}},
		{"totalDWeight", bson.D{{"$sum", "$default_weight"}}},
		{"totalSWeight", bson.D{{"$sum", "$self_defined_weight"}}},
		{"totalRecord", bson.D{{"$sum", 1}}},
	}}}

	var rawData NearestGoodBehaviorData
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{geoStage, timeStageToday, sortStage, groupStage, groupMergeStage})
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("aggregate nearest good behavior score")
		return NearestGoodBehaviorData{}, err
	}
	var results bson.M
	if !cursor.Next(ctx) {
		return rawData, errors.New("no record")
	}
	if err := cursor.Decode(&results); err != nil {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("decode nearest good behavior score")
		return rawData, err
	}
	rawData.DefaultBehaviorWeight = results["totalDWeight"].(float64)
	rawData.DefaultBehaviorCount = results["totalDCount"].(int32)
	rawData.SelfDefinedBehaviorCount = results["totalSCount"].(int32)
	rawData.SelfDefinedBehaviorWeight = results["totalSWeight"].(float64)
	rawData.TotalRecordCount = results["totalRecord"].(int32)

	// Previous day
	cursorYesterday, err := collection.Aggregate(ctx, mongo.Pipeline{geoStage, timeStageYesterday, sortStage, groupStage, groupMergeStage})
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("aggregate nearest good behavior score")
		return NearestGoodBehaviorData{}, err
	}
	var resultsYesterday bson.M
	if !cursorYesterday.Next(ctx) {
		return rawData, errors.New("no record")
	}
	if err := cursorYesterday.Decode(&resultsYesterday); err != nil {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("decode nearest good behavior score")
		return rawData, err
	}
	rawData.PastTotalRecordCount = resultsYesterday["totalRecord"].(int32)
	rawData.PastDefaultBehaviorWeight = resultsYesterday["totalDWeight"].(float64)
	rawData.PastDefaultBehaviorCount = resultsYesterday["totalDCount"].(int32)
	rawData.PastSelfDefinedBehaviorCount = resultsYesterday["totalSCount"].(int32)
	rawData.PastSelfDefinedBehaviorWeight = resultsYesterday["totalSWeight"].(float64)
	return rawData, nil
}

func todayStartAt() int64 {
	curTime := time.Now()
	start := time.Date(curTime.Year(), curTime.Month(), curTime.Day(), 0, 0, 0, 0, time.UTC)
	return start.Unix()
}
