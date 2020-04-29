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
		updatedBehavior := behaviors
		for _, existBehavior := range p.CustomerizedBehaviors {
			for _, newBehavior := range behaviors {
				if newBehavior.ID != existBehavior.ID {
					updatedBehavior = append(updatedBehavior, existBehavior)
				}
			}
		}
		opts := options.Update().SetUpsert(false)
		filter := bson.D{{"account_number", p.AccountNumber}}
		update := bson.D{{"$set", bson.D{{"customerized_behavior", updatedBehavior}}}}

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
		updatedSymptom := symptoms
		for _, existSymptom := range p.CustomerizedSymptom {
			for _, newBehavior := range symptoms {
				if newBehavior.ID != existSymptom.ID {
					updatedSymptom = append(updatedSymptom, existSymptom)
				}
			}
		}
		opts := options.Update().SetUpsert(false)
		filter := bson.D{{"account_number", p.AccountNumber}}
		update := bson.D{{"$set", bson.D{{"customerized_symptom", updatedSymptom}}}}

		result, err := c.UpdateOne(context.TODO(), filter, update, opts)
		if result.MatchedCount == 0 || err != nil {
			return err
		}
		profiles = append(profiles, p)
	}

	return nil
}
