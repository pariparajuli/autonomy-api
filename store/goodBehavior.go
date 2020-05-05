package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/score"
	"github.com/bitmark-inc/autonomy-api/utils"
)

const (
	DuplicateKeyCode = 11000
)

type BehaviorSource string

const (
	OfficialBehavior   BehaviorSource = "official"
	CustomizedBehavior BehaviorSource = "customized"
)

var localizedBehaviors map[string][]schema.Behavior = map[string][]schema.Behavior{}

// GoodBehaviorReport save a GoodBehaviorData into Database
type GoodBehaviorReport interface {
	CreateBehavior(behavior schema.Behavior) (string, error)
	GoodBehaviorSave(data *schema.BehaviorReportData) error
	NearestGoodBehavior(distInMeter int, location schema.Location) (score.NearestGoodBehaviorData, score.NearestGoodBehaviorData, error)
	IDToBehaviors(ids []schema.GoodBehaviorType) ([]schema.Behavior, []schema.Behavior, []schema.GoodBehaviorType, error)
	AreaCustomizedBehaviorList(distInMeter int, location schema.Location) ([]schema.Behavior, error)
	ListOfficialBehavior(string) ([]schema.Behavior, error)
}

func (m *mongoDB) ListOfficialBehavior(lang string) ([]schema.Behavior, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	lang = strings.ReplaceAll(strings.ToLower(lang), "-", "_")

	if behaviors, ok := localizedBehaviors[lang]; ok {
		return behaviors, nil
	}

	c := m.client.Database(m.database)

	query := bson.M{"source": schema.OfficialBehavior}

	cursor, err := c.Collection(schema.BehaviorCollection).Find(ctx, query, options.Find().SetSort(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}

	loc := utils.NewLocalizer(lang)

	behaviors := make([]schema.Behavior, 0)

	for cursor.Next(ctx) {
		var b schema.Behavior
		if err := cursor.Decode(&b); err != nil {
			return nil, err
		}

		if name, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: fmt.Sprintf("behaviors.%s.name", b.ID),
		}); err == nil {
			b.Name = name
		} else {
			log.WithError(err).Warnf("can not decode name")
		}

		if desc, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: fmt.Sprintf("behaviors.%s.desc", b.ID),
		}); err == nil {
			b.Desc = desc
		} else {
			log.WithError(err).Warnf("can not decode description")
		}

		behaviors = append(behaviors, b)
	}

	return behaviors, nil
}

func (m *mongoDB) CreateBehavior(behavior schema.Behavior) (string, error) {
	if 0 == len(behavior.Name) {
		return "", errors.New("empty behavior")
	}
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
	if 0 == len(data.OfficialBehaviors) {
		data.OfficialBehaviors = []schema.Behavior{}
	}
	if 0 == len(data.CustomizedBehaviors) {
		data.CustomizedBehaviors = []schema.Behavior{}
	}
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
			} else {
				return nil, nil, nil, err
			}
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
func (m *mongoDB) NearestGoodBehavior(distInMeter int, location schema.Location) (score.NearestGoodBehaviorData, score.NearestGoodBehaviorData, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := m.client.Database(m.database)
	collection := db.Collection(schema.BehaviorReportCollection)
	todayBegin := todayStartAt()
	log.Debugf("time period today > %v, yesterday %v~ %v ", todayBegin, todayBegin-86400, todayBegin)
	timeStageToday := bson.D{{"$match", bson.M{"ts": bson.M{"$gte": todayBegin}}}}

	sortStage := bson.D{{"$sort", bson.D{{"ts", -1}}}}
	timeStageYesterday := bson.D{{"$match", bson.M{"ts": bson.M{"$gte": todayBegin - 86400, "$lt": todayBegin}}}}
	groupStage := bson.D{{"$group", bson.D{
		{"_id", "$profile_id"},
		{"account_number", bson.D{{"$first", "$account_number"}}},
		{"official_count", bson.D{{"$first", bson.D{{"$size", "$official_behaviors"}}}}},
		{"customized_count", bson.D{{"$first", bson.D{{"$size", "$customized_behaviors"}}}}},
		{"official_weight", bson.D{{"$first", "$official_weight"}}},
		{"customized_weight", bson.D{{"$first", "$customized_weight"}}},
	}}}
	groupMergeStage := bson.D{{"$group", bson.D{
		{"_id", 1},
		{"officialCount", bson.D{{"$sum", "$official_count"}}},
		{"customizedCount", bson.D{{"$sum", "$customized_count"}}},
		{"officialWeight", bson.D{{"$sum", "$official_weight"}}},
		{"customizedWeight", bson.D{{"$sum", "$customized_weight"}}},
		{"totalCount", bson.D{{"$sum", 1}}},
	}}}

	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{geoAggregate(distInMeter, location), timeStageToday, sortStage, groupStage, groupMergeStage})
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("aggregate nearest good behavior score")
		return score.NearestGoodBehaviorData{}, score.NearestGoodBehaviorData{}, err
	}

	var resultToday score.NearestGoodBehaviorData
	if cursor.Next(ctx) {
		if err := cursor.Decode(&resultToday); err != nil {
			log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("decode nearest good behavior score")
			return score.NearestGoodBehaviorData{}, score.NearestGoodBehaviorData{}, err
		}
	}

	// Previous day
	cursorYesterday, err := collection.Aggregate(ctx, mongo.Pipeline{geoAggregate(distInMeter, location), timeStageYesterday, sortStage, groupStage, groupMergeStage})
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("aggregate nearest good behavior score")
		return score.NearestGoodBehaviorData{}, score.NearestGoodBehaviorData{}, err
	}
	var resultYesterday score.NearestGoodBehaviorData
	if cursorYesterday.Next(ctx) {
		if err := cursorYesterday.Decode(&resultYesterday); err != nil {
			log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("decode nearest good behavior score")
			return score.NearestGoodBehaviorData{}, score.NearestGoodBehaviorData{}, err
		}
	}
	log.Debug(fmt.Sprintf("NearestGoodBehavior TotalCount:%v OfficialWeight:%v,OfficialCount:%v, CustimizedWeight:%v, CustimizedCount:%v, TotalCountYesterday:%v, OfficialWeightYesterday:%v, OfficialCountYesterday:%v, CustimizedWeightYesterday:%v, CustimizedCountYesterday:%v",
		resultToday.TotalCount, resultToday.OfficialWeight, resultToday.OfficialCount, resultToday.CustomizedWeight, resultToday.CustomizedCount,
		resultYesterday.TotalCount, resultYesterday.OfficialWeight, resultYesterday.OfficialCount, resultYesterday.CustomizedWeight, resultYesterday.CustomizedCount))
	return resultToday, resultYesterday, nil
}

func todayStartAt() int64 {
	curTime := time.Now().UTC()
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
