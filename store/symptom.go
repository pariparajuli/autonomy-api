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

var localizedSymptoms map[string][]schema.Symptom = map[string][]schema.Symptom{}
var localizedSuggestedSymptoms map[string][]schema.Symptom = map[string][]schema.Symptom{}

type Symptom interface {
	CreateSymptom(symptom schema.Symptom) (string, error)
	ListOfficialSymptoms(string) ([]schema.Symptom, error)
	ListSuggestedSymptoms(lang string) ([]schema.Symptom, error)
	ListCustomizedSymptoms() ([]schema.Symptom, error)
	SymptomReportSave(data *schema.SymptomReportData) error
	AreaCustomizedSymptomList(distInMeter int, location schema.Location) ([]schema.Symptom, error)
	IDToSymptoms(ids []schema.SymptomType) ([]schema.Symptom, []schema.Symptom, []schema.SymptomType, error)
	NearestSymptomScore(distInMeter int, location schema.Location) (schema.NearestSymptomData, schema.NearestSymptomData, error)
	NearOfficialSymptomInfo(meter int, loc schema.Location) (schema.SymptomDistribution, float64, float64, error)
	SymptomCount(meter int, loc schema.Location) (int, error)
}

func (m *mongoDB) CreateSymptom(symptom schema.Symptom) (string, error) {
	if 0 == len(symptom.Name) {
		return "", errors.New("empty symptom")
	}
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

func (m *mongoDB) ListOfficialSymptoms(lang string) ([]schema.Symptom, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	lang = strings.ReplaceAll(strings.ToLower(lang), "-", "_")

	if symptoms, ok := localizedSymptoms[lang]; ok {
		return symptoms, nil
	}

	c := m.client.Database(m.database)

	query := bson.M{"source": schema.OfficialSymptom}

	cursor, err := c.Collection(schema.SymptomCollection).Find(ctx, query, options.Find().SetSort(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}

	loc := utils.NewLocalizer(lang)

	symptoms := make([]schema.Symptom, 0)

	for cursor.Next(ctx) {
		var s schema.Symptom
		if err := cursor.Decode(&s); err != nil {
			return nil, err
		}

		if name, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: fmt.Sprintf("symptoms.%s.name", s.ID),
		}); err == nil {
			s.Name = name
		} else {
			log.WithError(err).Warnf("can not decode name")
		}

		if desc, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: fmt.Sprintf("symptoms.%s.desc", s.ID),
		}); err == nil {
			s.Desc = desc
		} else {
			log.WithError(err).Warnf("can not decode description")
		}

		symptoms = append(symptoms, s)
	}

	localizedSymptoms[lang] = symptoms

	return symptoms, nil
}

func (m *mongoDB) ListSuggestedSymptoms(lang string) ([]schema.Symptom, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	lang = strings.ReplaceAll(strings.ToLower(lang), "-", "_")

	if symptoms, ok := localizedSuggestedSymptoms[lang]; ok {
		return symptoms, nil
	}

	c := m.client.Database(m.database)

	query := bson.M{"source": schema.SuggestedSymptom}

	cursor, err := c.Collection(schema.SymptomCollection).Find(ctx, query, options.Find().SetSort(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}

	loc := utils.NewLocalizer(lang)

	symptoms := make([]schema.Symptom, 0)

	for cursor.Next(ctx) {
		var s schema.Symptom
		if err := cursor.Decode(&s); err != nil {
			return nil, err
		}

		if name, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: fmt.Sprintf("symptoms.%s.name", s.ID),
		}); err == nil {
			s.Name = name
		} else {
			log.WithError(err).Warnf("can not decode name")
		}

		symptoms = append(symptoms, s)
	}

	localizedSuggestedSymptoms[lang] = symptoms

	return symptoms, nil
}

func (m *mongoDB) ListCustomizedSymptoms() ([]schema.Symptom, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	c := m.client.Database(m.database)
	query := bson.M{"source": schema.CustomizedSymptom}
	cursor, err := c.Collection(schema.SymptomCollection).Find(ctx, query, options.Find().SetSort(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}

	symptoms := make([]schema.Symptom, 0)
	for cursor.Next(ctx) {
		var s schema.Symptom
		if err := cursor.Decode(&s); err != nil {
			return nil, err
		}
		symptoms = append(symptoms, s)
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
	log.Debug(fmt.Sprintf("AreaCustomizedSymptomList location long:%f, lat: %f ", location.Longitude, location.Latitude))
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
	cSymptoms := make([]schema.Symptom, 0)
	for _, b := range cbMap {
		cSymptoms = append(cSymptoms, b)
	}
	return cSymptoms, nil
}

// NearestGoodBehaviorScore return  the total symptom score and delta score of users within distInMeter range
func (m *mongoDB) NearestSymptomScore(distInMeter int, location schema.Location) (schema.NearestSymptomData, schema.NearestSymptomData, error) {
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
			{"official_symptoms", bson.D{
				{"$first", "$official_symptoms"},
			}},
		}}}
	officialDistribution := make(schema.SymptomDistribution)
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{geoStage, timeStageToday, sortStage, groupStage})
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("aggregate nearest symptom score")
		return schema.NearestSymptomData{}, schema.NearestSymptomData{}, err
	}
	userCount := int32(0)
	OfficialSymptom := 0
	CustomizedSymptom := 0
	var dataToday schema.NearestSymptomData
	for cursor.Next(ctx) {
		var result schema.SymptomReportData
		if err := cursor.Decode(&result); err != nil {
			log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("decode nearest symptom score")
			continue
		}
		userCount++
		OfficialSymptom = OfficialSymptom + len(result.OfficialSymptoms)
		CustomizedSymptom = CustomizedSymptom + len(result.CustomizedSymptoms)
		for _, s := range result.OfficialSymptoms {
			value, ok := officialDistribution[schema.SymptomType(s.ID)]
			if !ok {
				officialDistribution[s.ID] = 1
			} else {
				officialDistribution[s.ID] = value + 1
			}
		}
		dataToday = schema.NearestSymptomData{
			UserCount:          float64(userCount),
			OfficialCount:      float64(OfficialSymptom),
			CustomizedCount:    float64(CustomizedSymptom),
			WeightDistribution: officialDistribution,
		}
	}

	officialDistributionYesterday := make(schema.SymptomDistribution)
	cursor, err = collection.Aggregate(ctx, mongo.Pipeline{geoStage, timeStageYesterday, sortStage, groupStage})
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("aggregate nearest symptom score")
		return schema.NearestSymptomData{}, schema.NearestSymptomData{}, err
	}
	userCount = 0
	OfficialSymptom = 0
	CustomizedSymptom = 0
	var dataYesterday schema.NearestSymptomData
	for cursor.Next(ctx) {
		var result schema.SymptomReportData
		if err := cursor.Decode(&result); err != nil {
			log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("decode nearest symptom score")
			continue
		}
		userCount++
		OfficialSymptom = OfficialSymptom + len(result.OfficialSymptoms)
		CustomizedSymptom = CustomizedSymptom + len(result.CustomizedSymptoms)
		for _, s := range result.OfficialSymptoms {
			value, ok := officialDistributionYesterday[schema.SymptomType(s.ID)]
			if !ok {
				officialDistributionYesterday[s.ID] = 1
			} else {
				officialDistributionYesterday[s.ID] = value + 1
			}
		}
		dataYesterday = schema.NearestSymptomData{
			UserCount:          float64(userCount),
			OfficialCount:      float64(OfficialSymptom),
			CustomizedCount:    float64(CustomizedSymptom),
			WeightDistribution: officialDistributionYesterday,
		}
	}

	return dataToday, dataYesterday, nil

}

func (m *mongoDB) NearOfficialSymptomInfo(meter int, loc schema.Location) (schema.SymptomDistribution, float64, float64, error) {
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
		return schema.SymptomDistribution{}, 0, 0, err
	}

	type data struct {
		AccountNumber string           `bson:"account_number"`
		Symptoms      []schema.Symptom `bson:"symptoms"`
		Timestamp     int64            `bson:"ts"`
	}

	var d data
	distribution := make(schema.SymptomDistribution)
	var symptomCount float64

	for cur.Next(ctx) {
		if err = cur.Decode(&d); err != nil {
			log.WithFields(log.Fields{
				"prefix":   mongoLogPrefix,
				"distance": meter,
				"lat":      loc.Latitude,
				"lng":      loc.Longitude,
				"error":    err,
			}).Error("decode nearby official symptoms")
			return schema.SymptomDistribution{}, 0, 0, err
		}

		for _, s := range d.Symptoms {
			distribution[s.ID]++
			symptomCount++
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
		return schema.SymptomDistribution{}, 0, 0, err
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
			return distribution, 0, 0, err
		}
	}

	log.WithFields(log.Fields{
		"prefix":                        mongoLogPrefix,
		"user_count":                    userSumData.Total,
		"official_symptom_distribution": distribution,
		"official_symptom_count":        symptomCount,
	}).Debug("near symptom info")

	return distribution, symptomCount, float64(userSumData.Total), nil
}

func (m *mongoDB) SymptomCount(meter int, loc schema.Location) (int, error) {
	log.WithFields(log.Fields{
		"prefix":   mongoLogPrefix,
		"distance": meter,
		"lat":      loc.Latitude,
		"lng":      loc.Longitude,
	}).Info("get symptom count")

	c := m.client.Database(m.database).Collection(schema.SymptomReportCollection)
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

	todayBeginTime := todayStartAt()
	addedToday := bson.D{
		{"$match", bson.M{
			"ts": bson.M{
				"$gte": todayBeginTime - 86400,
			},
		}},
	}

	latestFirst := bson.D{
		{"$sort", bson.M{
			"ts": -1,
		}},
	}

	groupAll := bson.D{
		{"$group", bson.M{
			"_id": "$account_number",
			"official": bson.M{
				"$first": "$official_symptoms",
			},
			"customized": bson.M{
				"$first": "$customized_symptoms",
			},
			"ts": bson.M{
				"$first": "$ts",
			},
		}},
	}

	cur, err := c.Aggregate(ctx, mongo.Pipeline{
		geoNear,
		addedToday,
		latestFirst,
		groupAll,
	})

	if nil != err {
		log.WithFields(log.Fields{
			"prefix": mongoLogPrefix,
			"error":  err,
		}).Error("aggregate nearby symptoms")
		return 0, err
	}

	type aggregatedData struct {
		AccountNumber string           `bson:"_id"`
		Official      []schema.Symptom `bson:"official"`
		Customized    []schema.Symptom `bson:"customized"`
		Timestamp     int64            `bson:"ts"`
	}

	var data aggregatedData
	var officialCount, customizedCount int

	for cur.Next(ctx) {
		err = cur.Decode(&data)
		if nil != err {
			log.WithFields(log.Fields{
				"prefix": mongoLogPrefix,
				"error":  err,
			}).Error("decode aggregated symptoms")
			return 0, err
		}

		officialCount += len(data.Official)
		customizedCount += len(data.Customized)
	}

	log.WithFields(log.Fields{
		"prefix":           mongoLogPrefix,
		"timestamp":        todayBeginTime,
		"official_count":   officialCount,
		"customized_count": customizedCount,
	}).Debug("aggregated symptoms")

	return officialCount + customizedCount, nil
}
