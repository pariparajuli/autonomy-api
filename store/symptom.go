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
	FindSymptomsByIDs(ids []string) ([]schema.Symptom, error)
	FindNearbySymptomDistribution(dist int, loc schema.Location, start, end int64) (schema.SymptomDistribution, error)
	FindNearbyReporterCount(dist int, loc schema.Location, start, end int64) (int, error)
	FindNearbyNonOfficialSymptoms(dist int, loc schema.Location) ([]schema.Symptom, error)
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

	symptom.ID = hex.EncodeToString(h.Sum(nil))
	symptom.Source = schema.CustomizedSymptom

	if _, err := c.Collection(schema.SymptomCollection).InsertOne(ctx, &symptom); err != nil {
		if we, hasErr := err.(mongo.WriteException); hasErr {
			if 1 == len(we.WriteErrors) && DuplicateKeyCode == we.WriteErrors[0].Code {
				return symptom.ID, nil
			}
		}
		return "", err
	}

	return symptom.ID, nil
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

func (m *mongoDB) FindSymptomsByIDs(ids []string) ([]schema.Symptom, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	c := m.client.Database(m.database).Collection(schema.SymptomCollection)

	query := bson.M{"_id": bson.M{"$in": ids}}

	cursor, err := c.Find(ctx, query)
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

// FindNearbySymptomDistribution returns the mapping of each reported symptom and the number of users who have reported it
// in the specified area and within the specified time rage.
//
// Duplicated reported symptoms of a user are seen as one symptom.
//
// Here's the example: within the specified time interval, assume there are following 5 reports:
//
// | user  | symptoms              |
// |-------|-----------------------|
// | userA | [cough, fever]        |
// | userA | [fever, cough, nasal] |
// | userB | [fever]               |
// | userB | [fever]               |
// | userB | [fever] 			    |
//
// symptom_distribution = {fever: 2, cough: 1, nasal: 1}
func (m *mongoDB) FindNearbySymptomDistribution(dist int, loc schema.Location, start, end int64) (schema.SymptomDistribution, error) {
	c := m.client.Database(m.database).Collection(schema.SymptomReportCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	pipeline := []bson.M{
		aggStageGeoProximity(dist, loc),
		aggStageReportedBetween(start, end),
		{
			"$project": bson.M{
				"profile_id":     1,
				"account_number": 1,
				"symptoms": bson.M{
					"$concatArrays": bson.A{
						bson.M{"$ifNull": bson.A{"$official_symptoms", bson.A{}}},
						bson.M{"$ifNull": bson.A{"$customized_symptoms", bson.A{}}},
						bson.M{"$ifNull": bson.A{"$symptoms", bson.A{}}},
					},
				},
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$symptoms",
				"preserveNullAndEmptyArrays": false,
			},
		},
		{
			"$group": bson.M{
				"_id": "$profile_id",
				"symptoms": bson.M{
					"$addToSet": "$symptoms",
				},
			},
		}, // for each user, the number of types of symptoms reported
		{
			"$unwind": bson.M{
				"path":                       "$symptoms",
				"preserveNullAndEmptyArrays": false,
			},
		},
		{
			"$group": bson.M{
				"_id": "$symptoms._id",
				"count": bson.M{
					"$sum": 1,
				},
			},
		}, // for each symptom, the number of users who have reported it
	}

	cursor, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var aggItem struct {
		SymptomID string `bson:"_id"`
		Count     int    `bson:"count"`
	}
	result := make(map[string]int)
	for cursor.Next(ctx) {
		if err := cursor.Decode(&aggItem); err != nil {
			return nil, err
		}
		result[aggItem.SymptomID] = aggItem.Count
	}

	return result, nil
}

// FindNearbyReporterCount returns the number of users who have reported symptoms
// in the specified area and within the specified time rage.
func (m *mongoDB) FindNearbyReporterCount(dist int, loc schema.Location, start, end int64) (int, error) {
	c := m.client.Database(m.database).Collection(schema.SymptomReportCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	pipeline := []bson.M{
		aggStageGeoProximity(dist, loc),
		aggStageReportedBetween(start, end),
		{
			"$group": bson.M{
				"_id": "$profile_id",
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"count": bson.M{
					"$sum": 1,
				},
			},
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

// FindNearbyNonOfficialSymptoms returns non-official symptoms reported today in the specified area.
func (m *mongoDB) FindNearbyNonOfficialSymptoms(dist int, loc schema.Location) ([]schema.Symptom, error) {
	distribution, err := m.FindNearbySymptomDistribution(dist, loc, 0, 9223372036854775807)
	if err != nil {
		return nil, err
	}

	nonOfficialSymptomIDs := make([]string, 0)
	for symptomID := range distribution {
		if !schema.OfficialSymptoms[symptomID] {
			nonOfficialSymptomIDs = append(nonOfficialSymptomIDs, symptomID)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	c := m.client.Database(m.database).Collection(schema.SymptomCollection)
	query := bson.M{"_id": bson.M{"$in": nonOfficialSymptomIDs}}
	cursor, err := c.Find(ctx, query, options.Find().SetSort(bson.M{"_id": 1}))
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
