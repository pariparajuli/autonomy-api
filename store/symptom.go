package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bitmark-inc/autonomy-api/schema"
	log "github.com/sirupsen/logrus"
)

type SymptomList interface {
	CreateSymptom(symptom schema.Symptom) (string, error)
	ListSymptoms() ([]schema.Symptom, error)
}

type SymptomReport interface {
	SymptomReportSave(data *schema.SymptomReportData) error
	NearestSymptomScore(distInMeter int, location schema.Location) (float64, float64, int, int, error)
}

func (m *mongoDB) CreateSymptom(symptom schema.Symptom) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c := m.client.Database(m.database)

	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s=:=%s", symptom.Name, symptom.Desc)))

	symptom.ID = schema.SymptomType(hex.EncodeToString(h.Sum(nil)))
	symptom.Source = schema.CustomizedSymptom

	if _, err := c.Collection(schema.SymptomCollection).InsertOne(ctx, &symptom); err != nil {
		if we, hasErr := err.(mongo.WriteException); hasErr {
			if 1 == len(we.WriteErrors) && DuplicateKeyCode == we.WriteErrors[0].Code {
				return string(symptom.ID), nil
			}
		}
		return "", err
	}

	return string(symptom.ID), nil
}

func (m *mongoDB) ListSymptoms() ([]schema.Symptom, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	c := m.client.Database(m.database)

	query := bson.M{"source": schema.OfficialSymptom}

	cursor, err := c.Collection(schema.SymptomCollection).Find(ctx, query, options.Find().SetSort(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}

	symptoms := make([]schema.Symptom, 0)
	if err := cursor.All(ctx, &symptoms); err != nil {
		return nil, err
	}

	return symptoms, nil
}

// SymptomReportSave save  a record instantly in database
func (m *mongoDB) SymptomReportSave(data *schema.SymptomReportData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c := m.client.Database(m.database)
	_, err := c.Collection(schema.SymptomReportCollection).InsertOne(ctx, *data)
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
func (m *mongoDB) NearestSymptomScore(distInMeter int, location schema.Location) (float64, float64, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := m.client.Database(m.database)
	collection := db.Collection(schema.SymptomReportCollection)
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
			{"symptom_score", bson.D{
				{"$first", "$symptom_score"},
			}},
			{"account_number", bson.D{
				{"$first", "$account_number"},
			}},
			{"symptoms", bson.D{
				{"$first", "$symptoms"},
			}},
		}}}

	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{geoStage, timeStageToday, sortStage, groupStage})
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("aggregate nearest symptom score")
		return 0, 0, 0, 0, err
	}
	sum := float64(0)
	count := 0
	totalSymptom := 0
	for cursor.Next(ctx) {
		var result schema.SymptomReportData
		if err := cursor.Decode(&result); err != nil {
			log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("decode nearest symptom score")
			continue
		}
		sum = sum + result.SymptomScore
		count++
		totalSymptom = totalSymptom + len(result.Symptoms)
	}
	score := float64(100)
	if count > 0 {
		score = 100 - 100*(sum/(schema.TotalSymptomWeight*2))
		if score < 0 {
			score = 0
		}
	}

	// Previous day
	cursorYesterday, err := collection.Aggregate(ctx, mongo.Pipeline{geoStage, timeStageYesterday, sortStage, groupStage})
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("aggregate nearest symptom score")
		return 0, 0, 0, 0, err
	}
	sumYesterday := float64(0)
	countYesterday := 0
	totalSymptomYesterday := 0
	for cursorYesterday.Next(ctx) {
		var result schema.SymptomReportData
		if err := cursor.Decode(&result); err != nil {
			log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("decode nearest symptom score")
			continue
		}
		sumYesterday = sumYesterday + result.SymptomScore
		countYesterday++
		totalSymptomYesterday = totalSymptomYesterday + len(result.Symptoms)
	}
	scoreYesterday := float64(100)
	if countYesterday > 0 {
		scoreYesterday = 100 - 100*(sumYesterday/(schema.TotalSymptomWeight*2))
		if scoreYesterday < 0 {
			scoreYesterday = 0
		}
	}
	scoreDelta := score - scoreYesterday
	symptomDelta := totalSymptom - totalSymptomYesterday
	return score, scoreDelta, totalSymptom, symptomDelta, nil
}
