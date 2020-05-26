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
	"github.com/bitmark-inc/autonomy-api/utils"
)

const (
	DuplicateKeyCode = 11000
)

var localizedBehaviors map[string][]schema.Behavior = map[string][]schema.Behavior{}

// GoodBehaviorReport save a GoodBehaviorData into Database
type GoodBehaviorReport interface {
	CreateBehavior(behavior schema.Behavior) (string, error)
	GoodBehaviorSave(data *schema.BehaviorReportData) error
	IDToBehaviors(ids []schema.GoodBehaviorType) ([]schema.Behavior, []schema.Behavior, []schema.GoodBehaviorType, error)
	AreaCustomizedBehaviorList(distInMeter int, location schema.Location) ([]schema.Behavior, error)
	FindNearbyBehaviorDistribution(dist int, loc schema.Location, start, end int64) (map[string]int, error)
	FindNearbyBehaviorReportTimes(dist int, loc schema.Location, start, end int64) (int, error)
	ListOfficialBehavior(string) ([]schema.Behavior, error)
	ListCustomizedBehaviors() ([]schema.Behavior, error)
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

func (m *mongoDB) ListCustomizedBehaviors() ([]schema.Behavior, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	c := m.client.Database(m.database)
	query := bson.M{"source": schema.CustomizedBehavior}
	cursor, err := c.Collection(schema.BehaviorCollection).Find(ctx, query, options.Find().SetSort(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}

	behaviors := make([]schema.Behavior, 0)
	for cursor.Next(ctx) {
		var s schema.Behavior
		if err := cursor.Decode(&s); err != nil {
			return nil, err
		}
		behaviors = append(behaviors, s)
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

func (m *mongoDB) FindNearbyBehaviorDistribution(dist int, loc schema.Location, start, end int64) (map[string]int, error) {
	c := m.client.Database(m.database).Collection(schema.BehaviorReportCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	pipeline := []bson.M{
		aggStageGeoProximity(dist, loc),
		aggStageReportedBetween(start, end),
		{
			"$project": bson.M{
				"profile_id":     1,
				"account_number": 1,
				"behaviors": bson.M{
					"$concatArrays": bson.A{
						bson.M{"$ifNull": bson.A{"$official_behaviors", bson.A{}}},
						bson.M{"$ifNull": bson.A{"$customized_behaviors", bson.A{}}},
						bson.M{"$ifNull": bson.A{"$behaviors", bson.A{}}},
					},
				},
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$behaviors",
				"preserveNullAndEmptyArrays": false,
			},
		},
		{
			"$group": bson.M{
				"_id": "$behaviors._id",
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
	}

	cursor, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var aggItem struct {
		BehaviorID string `bson:"_id"`
		Count      int    `bson:"count"`
	}
	result := make(map[string]int)
	for cursor.Next(ctx) {
		if err := cursor.Decode(&aggItem); err != nil {
			return nil, err
		}
		result[aggItem.BehaviorID] = aggItem.Count
	}

	return result, nil
}

func (m *mongoDB) FindNearbyBehaviorReportTimes(dist int, loc schema.Location, start, end int64) (int, error) {
	c := m.client.Database(m.database).Collection(schema.BehaviorReportCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	pipeline := []bson.M{
		aggStageGeoProximity(dist, loc),
		aggStageReportedBetween(start, end),
		{
			"$count": "count",
		},
	}
	cursor, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}

	if !cursor.Next(ctx) {
		return 0, nil
	}

	var result struct {
		Count int `bson:"count"`
	}
	if err := cursor.Decode(&result); err != nil {
		return 0, err
	}

	return result.Count, nil
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
	cBehaviors := make([]schema.Behavior, 0)
	for _, b := range cbMap {
		cBehaviors = append(cBehaviors, b)
	}
	return cBehaviors, nil
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
