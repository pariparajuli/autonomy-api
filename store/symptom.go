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

	log "github.com/sirupsen/logrus"

	"github.com/bitmark-inc/autonomy-api/schema"
)

type Symptom interface {
	CreateSymptom(symptom schema.Symptom) (string, error)
	ListOfficialSymptoms() ([]schema.Symptom, error)
	SymptomReportSave(data *schema.SymptomReportData) error
	AreaCustomizedSymptomList(distInMeter int, location schema.Location) ([]schema.Symptom, error)
	IDToSymptoms(ids []schema.SymptomType) ([]schema.Symptom, []schema.Symptom, []schema.SymptomType, error)
	NearestSymptomScore(distInMeter int, location schema.Location) (float64, float64, int, int, error)
	NearOfficialSymptomInfo(meter int, loc schema.Location) (schema.SymptomDistribution, float64, error)
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

func (m *mongoDB) ListOfficialSymptoms() ([]schema.Symptom, error) {
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

// IDToSymptoms return official and customized symptoms from a list of SymptomType ID
func (m *mongoDB) IDToSymptoms(ids []schema.SymptomType) ([]schema.Symptom, []schema.Symptom, []schema.SymptomType, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	c := m.client.Database(m.database)
	var foundOfficial []schema.Symptom
	var foundCustomized []schema.Symptom
	var notFound []schema.SymptomType
	for _, id := range ids {
		query := bson.M{"_id": string(id)}
		var result schema.Symptom
		err := c.Collection(schema.SymptomCollection).FindOne(ctx, query).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				notFound = append(notFound, id)
			} else {
				return nil, nil, nil, err
			}
		}
		if result.Source == schema.OfficialSymptom {
			foundOfficial = append(foundOfficial, result)
		} else {
			foundCustomized = append(foundCustomized, result)
		}

	}
	return foundOfficial, foundCustomized, notFound, nil
}

func (m *mongoDB) AreaCustomizedSymptomList(distInMeter int, location schema.Location) ([]schema.Symptom, error) {
	nonEmptyArray := bson.D{{"$match", bson.M{"customized_symptoms": bson.M{"$exists": true, "$ne": bson.A{}}}}}
	c := m.client.Database(m.database).Collection(schema.SymptomReportCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	log.Debug(fmt.Sprintf("AreaCustomizedSymptomList location long:%d, lat: %d ", location.Longitude, location.Latitude))
	cur, err := c.Aggregate(ctx, mongo.Pipeline{geoAggregate(distInMeter, location), nonEmptyArray})
	if nil != err {
		log.WithField("prefix", mongoLogPrefix).Errorf("area  customized symptom list with error: %s", err)
		return nil, fmt.Errorf("area  customized symptom list aggregate with error: %s", err)
	}
	cbMap := make(map[schema.SymptomType]schema.Symptom, 0)
	for cur.Next(ctx) {
		var b schema.SymptomReportData
		if errDecode := cur.Decode(&b); errDecode != nil {
			log.WithField("prefix", mongoLogPrefix).Infof("area  customized symptomwith error: %s", errDecode)
			return nil, fmt.Errorf("area  customized symptom decode record with error: %s", errDecode)
		}
		for _, symptom := range b.CustomizedSymptoms {
			cbMap[symptom.ID] = symptom
		}
	}
	var cSymptoms []schema.Symptom
	for _, b := range cbMap {
		cSymptoms = append(cSymptoms, b)
	}
	return cSymptoms, nil
}

// NearestGoodBehaviorScore return  the total symptom score and delta score of users within distInMeter range
func (m *mongoDB) NearestSymptomScore(distInMeter int, location schema.Location) (float64, float64, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := m.client.Database(m.database)
	collection := db.Collection(schema.SymptomReportCollection)
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
		totalSymptom = totalSymptom + len(result.OfficialSymptoms)
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
		totalSymptomYesterday = totalSymptomYesterday + len(result.OfficialSymptoms)
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

func (m *mongoDB) NearOfficialSymptomInfo(meter int, loc schema.Location) (schema.SymptomDistribution, float64, error) {
	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	geoNear := bson.D{
		{"$geoNear", bson.M{
			"near": bson.M{
				"type":        "Point",
				"coordinates": bson.A{loc.Longitude, loc.Latitude},
			},
			"distanceField": "dist",
			"maxDistance":   meter,
			"spherical":     true,
			"includeLocs":   "location",
		}},
	}

	joinSymptoms := bson.D{
		{"$lookup", bson.M{
			"from":         schema.SymptomReportCollection,
			"localField":   "account_number",
			"foreignField": "account_number",
			"as":           "symptoms",
		}},
	}

	nonEmptyArray := bson.D{
		{"$match", bson.M{
			"symptoms": bson.M{
				"$exists": true,
				"$ne":     bson.A{},
			},
		}},
	}

	nonNil := bson.D{
		{"$match", bson.M{
			"symptoms.official_symptoms": bson.M{"$ne": nil},
		}},
	}

	latestSymptoms := bson.D{
		{"$project", bson.M{
			"account_number": 1,
			"symptoms": bson.M{
				"$arrayElemAt": bson.A{"$symptoms.official_symptoms", -1},
			},
			"ts": bson.M{
				"$arrayElemAt": bson.A{"$symptoms.ts", -1},
			},
		}},
	}

	todayBeginTime := todayStartAt()
	updateToday := bson.D{
		{"$match", bson.M{"ts": bson.M{"$gte": todayBeginTime}}},
	}

	cur, err := c.Aggregate(ctx, mongo.Pipeline{
		geoNear,
		joinSymptoms,
		nonEmptyArray,
		nonNil,
		latestSymptoms,
		updateToday,
	})

	if nil != err {
		log.WithFields(log.Fields{
			"prefix":   mongoLogPrefix,
			"distance": meter,
			"lat":      loc.Latitude,
			"lng":      loc.Longitude,
			"error":    err,
		}).Error("aggregate nearby symptoms")
		return schema.SymptomDistribution{}, 0, err
	}

	type data struct {
		AccountNumber string           `bson:"account_number"`
		Symptoms      []schema.Symptom `bson:"symptoms"`
		Timestamp     int64            `bson:"ts"`
	}

	var d data
	distribution := make(schema.SymptomDistribution)

	for cur.Next(ctx) {
		if err = cur.Decode(&d); err != nil {
			log.WithFields(log.Fields{
				"prefix":   mongoLogPrefix,
				"distance": meter,
				"lat":      loc.Latitude,
				"lng":      loc.Longitude,
				"error":    err,
			}).Error("decode nearby official symptoms")
			return schema.SymptomDistribution{}, 0, err
		}

		for _, s := range d.Symptoms {
			distribution[s.ID]++
		}
	}

	joinGeographic := bson.D{
		{"$lookup", bson.M{
			"from":         schema.GeographicCollection,
			"localField":   "account_number",
			"foreignField": "account_number",
			"as":           "geographic",
		}},
	}

	geoDataExist := bson.D{
		{"$match", bson.M{
			"geographic": bson.M{
				"$exists": true,
				"$ne":     bson.A{},
			},
		}},
	}

	latestGeo := bson.D{
		{"$project", bson.M{
			"account_number": 1,
			"ts": bson.M{
				"$arrayElemAt": bson.A{"$geographic.ts", -1},
			},
		}},
	}

	sumUser := bson.D{
		{"$group", bson.M{
			"_id": nil,
			"total": bson.M{
				"$sum": 1,
			},
		}},
	}

	cur, err = c.Aggregate(ctx, mongo.Pipeline{
		geoNear,
		joinGeographic,
		geoDataExist,
		latestGeo,
		updateToday,
		sumUser,
	})

	if nil != err {
		log.WithFields(log.Fields{
			"prefix":   mongoLogPrefix,
			"distance": meter,
			"lat":      loc.Latitude,
			"lng":      loc.Longitude,
			"error":    err,
		}).Error("sum nearby user")
		return schema.SymptomDistribution{}, 0, err
	}

	type sumData struct {
		Total int `bson:"total"`
	}

	var userSumData sumData

	if cur.Next(ctx) {
		err = cur.Decode(&userSumData)
		if nil != err {
			log.WithFields(log.Fields{
				"prefix":   mongoLogPrefix,
				"distance": meter,
				"lat":      loc.Latitude,
				"lng":      loc.Longitude,
				"error":    err,
			}).Error("decode nearby user count")
			return distribution, 0, err
		}
	}

	log.WithFields(log.Fields{
		"prefix":               mongoLogPrefix,
		"user_count":           userSumData.Total,
		"symptom_distribution": distribution,
	}).Debug("near symptom info")

	return distribution, float64(userSumData.Total), nil
}
