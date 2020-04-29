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
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	DuplicateKeyCode = 11000
)

type BehaviorSource string

const (
	OfficialBehavior   BehaviorSource = "official"
	CustomizedBehavior BehaviorSource = "customized"
)

// GoodBehaviorReport save a GoodBehaviorData into Database
type GoodBehaviorReport interface {
	CreateBehavior(behavior schema.Behavior) (string, error)
	GoodBehaviorSave(data *schema.BehaviorReportData) error
	NearestGoodBehavior(distInMeter int, location schema.Location) (NearestGoodBehaviorData, error)
	IDToBehaviors(ids []schema.GoodBehaviorType) ([]schema.Behavior, []schema.Behavior, []schema.GoodBehaviorType, error)
	AreaCustomizedBehaviorList(distInMeter int, location schema.Location) ([]schema.Behavior, error)
	ListOfficialBehavior() ([]schema.Behavior, error)
}

type NearestGoodBehaviorData struct {
	TotalRecordCount             int32
	OfficialBehaviorWeight       float64
	OfficialBehaviorCount        int32
	CustomizedBehaviorWeight     float64
	CustomizedBehaviorCount      int32
	PastTotalRecordCount         int32
	PastOfficialBehaviorWeight   float64
	PastOfficialBehaviorCount    int32
	PastCustomizedBehaviorWeight float64
	PastCustomizedBehaviorCount  int32
}

func (m *mongoDB) ListOfficialBehavior() ([]schema.Behavior, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	c := m.client.Database(m.database)

	query := bson.M{"source": schema.OfficialBehavior}

	cursor, err := c.Collection(schema.BehaviorCollection).Find(ctx, query, options.Find().SetSort(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}

	behaviors := make([]schema.Behavior, 0)
	if err := cursor.All(ctx, &behaviors); err != nil {
		return nil, err
	}

	return behaviors, nil
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

// IDToBehaviors return official and customized behavuiors from a list of GoodBehaviorType ID
func (m *mongoDB) IDToBehaviors(ids []schema.GoodBehaviorType) ([]schema.Behavior, []schema.Behavior, []schema.GoodBehaviorType, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	c := m.client.Database(m.database)
	var foundOfficial []schema.Behavior
	var foundCustomized []schema.Behavior
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
			foundCustomized = append(foundCustomized, result)
		}

	}
	return foundOfficial, foundCustomized, notFound, nil
}

// AreaCustomizedBehaviorList return a list  of customized behaviors within distInMeter range
func (m *mongoDB) AreaCustomizedBehaviorList(distInMeter int, location schema.Location) ([]schema.Behavior, error) {
	filterStage := bson.D{{"$match", bson.M{"customized_weight": bson.M{"$gt": 0}}}}
	c := m.client.Database(m.database).Collection(schema.BehaviorReportCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	log.Info(fmt.Sprintf("NearestCustomizedBehaviorList location %v", location))
	cur, err := c.Aggregate(ctx, mongo.Pipeline{geoAggregate(distInMeter, location), filterStage})
	if nil != err {
		log.WithField("prefix", mongoLogPrefix).Errorf("nearest  distance  customized behavio with error: %s", err)
		return nil, fmt.Errorf("nearest  distance  customized behavior list query with error: %s", err)
	}
	cbMap := make(map[schema.GoodBehaviorType]schema.Behavior, 0)
	for cur.Next(ctx) {
		var b schema.BehaviorReportData
		if errDecode := cur.Decode(&b); errDecode != nil {
			log.WithField("prefix", mongoLogPrefix).Infof("query nearest distance with error: %s", errDecode)
			return nil, fmt.Errorf("nearest distance query decode record with error: %s", errDecode)
		}
		for _, behavior := range b.CustomizedBehaviors {
			cbMap[behavior.ID] = behavior
		}
	}
	var cBehaviors []schema.Behavior
	for _, b := range cbMap {
		cBehaviors = append(cBehaviors, b)
	}
	return cBehaviors, nil
}

// NearestGoodBehavior returns NearestGoodBehaviorData which caculates from location within distInMeter distance
func (m *mongoDB) NearestGoodBehavior(distInMeter int, location schema.Location) (NearestGoodBehaviorData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := m.client.Database(m.database)
	collection := db.Collection(schema.BehaviorReportCollection)
	todayBegin := todayStartAt()
	log.Debugf("time period today > %v, yesterday %v~ %v ", todayBegin, todayBegin-86400, todayBegin)
	timeStageToday := bson.D{{"$match", bson.M{"ts": bson.M{"$gte": todayBegin}}}}
	timeStageYesterday := bson.D{{"$match", bson.M{"ts": bson.M{"$gte": todayBegin - 86400, "$lt": todayBegin}}}}
	sortStage := bson.D{{"$sort", bson.D{{"ts", -1}}}}
	groupStage := bson.D{{"$group", bson.D{
		{"_id", "$profile_id"},
		{"account_number", bson.D{{"$first", "$account_number"}}},
		{"default_count", bson.D{{"$first", bson.D{{"$size", "$official_behaviors"}}}}},
		{"self_count", bson.D{{"$first", bson.D{{"$size", "$customized_behaviors"}}}}},
		{"default_weight", bson.D{{"$first", "$official_weight"}}},
		{"self_defined_weight", bson.D{{"$first", "$customized_weight"}}},
	}}}
	groupMergeStage := bson.D{{"$group", bson.D{
		{"_id", 1},
		{"totalDCount", bson.D{{"$sum", "$default_count"}}},
		{"totalSCount", bson.D{{"$sum", "$self_count"}}},
		{"totalDWeight", bson.D{{"$sum", "$official_weight"}}},
		{"totalSWeight", bson.D{{"$sum", "$customized_weight"}}},
		{"totalRecord", bson.D{{"$sum", 1}}},
	}}}

	var rawData NearestGoodBehaviorData
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{geoAggregate(distInMeter, location), timeStageToday, sortStage, groupStage, groupMergeStage})
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("aggregate nearest good behavior score")
		return NearestGoodBehaviorData{}, err
	}
	var results bson.M
	if cursor.Next(ctx) {
		if err := cursor.Decode(&results); nil == err {
			rawData.OfficialBehaviorWeight = results["totalDWeight"].(float64)
			rawData.OfficialBehaviorCount = results["totalDCount"].(int32)
			rawData.CustomizedBehaviorCount = results["totalSCount"].(int32)
			rawData.CustomizedBehaviorWeight = results["totalSWeight"].(float64)
			rawData.TotalRecordCount = results["totalRecord"].(int32)
		} else {
			log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("decode nearest good behavior score")
			return rawData, err
		}
	}
	// Previous day
	cursorYesterday, err := collection.Aggregate(ctx, mongo.Pipeline{geoAggregate(distInMeter, location), timeStageYesterday, sortStage, groupStage, groupMergeStage})
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("aggregate nearest good behavior score")
		return NearestGoodBehaviorData{}, err
	}
	var resultsYesterday bson.M
	if cursorYesterday.Next(ctx) {
		if err := cursorYesterday.Decode(&resultsYesterday); nil == err {
			rawData.PastTotalRecordCount = resultsYesterday["totalRecord"].(int32)
			rawData.PastOfficialBehaviorWeight = resultsYesterday["totalDWeight"].(float64)
			rawData.PastOfficialBehaviorCount = resultsYesterday["totalDCount"].(int32)
			rawData.PastCustomizedBehaviorCount = resultsYesterday["totalSCount"].(int32)
			rawData.PastCustomizedBehaviorWeight = resultsYesterday["totalSWeight"].(float64)
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

func geoAggregate(maxDist int, loc schema.Location) bson.D {
	return bson.D{{"$geoNear", bson.M{
		"near":          bson.M{"type": "Point", "coordinates": bson.A{loc.Longitude, loc.Latitude}},
		"distanceField": "dist",
		"spherical":     true,
		"maxDistance":   maxDist,
	}}}
}
