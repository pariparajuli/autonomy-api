package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	DuplicateKeyCode = 11000
)

type BehaviorSource string

const (
	OfficialBehavior     BehaviorSource = "official"
	CustomerizedBehavior BehaviorSource = "customerized"
)

// GoodBehaviorReport save a GoodBehaviorData into Database
type GoodBehaviorReport interface {
	CreateBehavior(behavior schema.Behavior) (string, error)
	GoodBehaviorSave(data *schema.BehaviorReportData) error
	NearestGoodBehavior(distInMeter int, location schema.Location) (NearestGoodBehaviorData, error)
	QueryBehaviors(ids []schema.GoodBehaviorType) ([]schema.Behavior, []schema.Behavior, []schema.GoodBehaviorType, error)
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

func (m *mongoDB) CreateBehavior(behavior schema.Behavior) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c := m.client.Database(m.database)

	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s=:=%s", behavior.Name, behavior.Desc)))
	behavior.ID = schema.GoodBehaviorType(hex.EncodeToString(h.Sum(nil)))
	behavior.Source = schema.CustomizedBehavior
	if _, err := c.Collection(schema.BehaviorCollection).InsertOne(ctx, &behavior); err != nil {
		if we, hasErr := err.(mongo.WriteException); hasErr {
			if 1 == len(we.WriteErrors) && DuplicateKeyCode == we.WriteErrors[0].Code {
				return string(behavior.ID), nil
			}
		}
		return "", err
	}
	return string(behavior.ID), nil
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

func (m *mongoDB) QueryBehaviors(ids []schema.GoodBehaviorType) ([]schema.Behavior, []schema.Behavior, []schema.GoodBehaviorType, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	c := m.client.Database(m.database)
	var foundOfficial []schema.Behavior
	var foundCustomerized []schema.Behavior
	var notFound []schema.GoodBehaviorType
	for _, id := range ids {
		query := bson.M{"_id": string(id)}
		var result schema.Behavior
		err := c.Collection(schema.BehaviorCollection).FindOne(ctx, query).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				notFound = append(notFound, id)
			}
			return nil, nil, nil, err
		}
		if result.Source == schema.OfficialBehavior {
			foundOfficial = append(foundOfficial, result)
		} else {
			foundCustomerized = append(foundCustomerized, result)
		}

	}
	return foundOfficial, foundCustomerized, notFound, nil
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
	if cursor.Next(ctx) {
		if err := cursor.Decode(&results); nil == err {
			rawData.DefaultBehaviorWeight = results["totalDWeight"].(float64)
			rawData.DefaultBehaviorCount = results["totalDCount"].(int32)
			rawData.SelfDefinedBehaviorCount = results["totalSCount"].(int32)
			rawData.SelfDefinedBehaviorWeight = results["totalSWeight"].(float64)
			rawData.TotalRecordCount = results["totalRecord"].(int32)
		} else {
			log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("decode nearest good behavior score")
			return rawData, err
		}
	}
	// Previous day
	cursorYesterday, err := collection.Aggregate(ctx, mongo.Pipeline{geoStage, timeStageYesterday, sortStage, groupStage, groupMergeStage})
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("aggregate nearest good behavior score")
		return NearestGoodBehaviorData{}, err
	}
	var resultsYesterday bson.M
	if cursorYesterday.Next(ctx) {
		if err := cursorYesterday.Decode(&resultsYesterday); nil == err {
			rawData.PastTotalRecordCount = resultsYesterday["totalRecord"].(int32)
			rawData.PastDefaultBehaviorWeight = resultsYesterday["totalDWeight"].(float64)
			rawData.PastDefaultBehaviorCount = resultsYesterday["totalDCount"].(int32)
			rawData.PastSelfDefinedBehaviorCount = resultsYesterday["totalSCount"].(int32)
			rawData.PastSelfDefinedBehaviorWeight = resultsYesterday["totalSWeight"].(float64)
		} else {
			log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("decode nearest good behavior score")
			return rawData, err
		}
	}
	return rawData, nil
}

func todayStartAt() int64 {
	curTime := time.Now()
	start := time.Date(curTime.Year(), curTime.Month(), curTime.Day(), 0, 0, 0, 0, time.UTC)
	return start.Unix()
}
