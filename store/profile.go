package store

import (
	"context"
	"fmt"

	"github.com/bitmark-inc/autonomy-api/consts"

	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bitmark-inc/autonomy-api/schema"
)

func (m *mongoDB) ProfileMetric(accountNumber string) (*schema.Metric, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
	query := bson.M{
		"account_number": bson.M{
			"$eq": accountNumber,
		},
	}

	var p schema.Profile
	err := c.FindOne(ctx, query).Decode(&p)
	if err != nil {
		return nil, err
	}

	// TODO: update account metric when account created
	return &p.Metric, nil
}

func (m *mongoDB) GetProfile(accountNumber string) (*schema.Profile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
	query := bson.M{
		"account_number": bson.M{
			"$eq": accountNumber,
		},
	}

	var p schema.Profile
	err := c.FindOne(ctx, query).Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (m *mongoDB) UpdateAreaProfileBehavior(behaviors []schema.Behavior, location schema.Location) error {
	if 0 == len(behaviors) {
		return fmt.Errorf("no behavior")
	}
	query := distanceQuery(consts.NEARBY_DISTANCE_RANGE, location)
	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cur, err := c.Find(ctx, query)
	if nil != err {
		log.WithField("prefix", mongoLogPrefix).Errorf("query nearest distance with error: %s", err)
		return fmt.Errorf("nearest distance query with error: %s", err)
	}

	var profiles []schema.Profile
	for cur.Next(ctx) {
		var p schema.Profile

		if errDecode := cur.Decode(&p); errDecode != nil {
			log.WithField("prefix", mongoLogPrefix).Infof("query nearest distance with error: %s", errDecode)
			return fmt.Errorf("nearest distance query decode record with error: %s", errDecode)
		}
		temp := make(map[schema.GoodBehaviorType]schema.Behavior, 0)
		for _, b := range behaviors {
			temp[b.ID] = b
		}
		for _, b := range p.CustomizedBehaviors {
			temp[b.ID] = b
		}
		var updatedBehavior []schema.Behavior
		for _, b := range temp {
			updatedBehavior = append(updatedBehavior, b)
		}

		opts := options.Update().SetUpsert(false)
		filter := bson.D{{"account_number", p.AccountNumber}}
		update := bson.D{{"$set", bson.D{{"customized_behavior", updatedBehavior}}}}

		result, err := c.UpdateOne(context.TODO(), filter, update, opts)
		if result.MatchedCount == 0 || err != nil {
			return err
		}
		profiles = append(profiles, p)
	}

	return nil
}

func (m *mongoDB) UpdateAreaProfileSymptom(symptoms []schema.Symptom, location schema.Location) error {
	if 0 == len(symptoms) {
		return fmt.Errorf("no symptom")
	}
	query := distanceQuery(consts.NEARBY_DISTANCE_RANGE, location)
	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cur, err := c.Find(ctx, query)
	if nil != err {
		log.WithField("prefix", mongoLogPrefix).Errorf("query nearest distance with error: %s", err)
		return fmt.Errorf("nearest distance query with error: %s", err)
	}

	var profiles []schema.Profile
	for cur.Next(ctx) {
		var p schema.Profile

		if errDecode := cur.Decode(&p); errDecode != nil {
			log.WithField("prefix", mongoLogPrefix).Infof("query nearest distance with error: %s", errDecode)
			return fmt.Errorf("nearest distance query decode record with error: %s", errDecode)
		}
		temp := make(map[schema.SymptomType]schema.Symptom, 0)
		for _, b := range symptoms {
			temp[b.ID] = b
		}
		for _, b := range p.CustomizedSymptom {
			temp[b.ID] = b
		}
		var updatedSymptom []schema.Symptom
		for _, b := range temp {
			updatedSymptom = append(updatedSymptom, b)
		}

		opts := options.Update().SetUpsert(false)
		filter := bson.D{{"account_number", p.AccountNumber}}
		update := bson.D{{"$set", bson.D{{"customized_symptom", updatedSymptom}}}}

		result, err := c.UpdateOne(context.TODO(), filter, update, opts)
		if result.MatchedCount == 0 || err != nil {
			return err
		}
		profiles = append(profiles, p)
	}

	return nil
}
