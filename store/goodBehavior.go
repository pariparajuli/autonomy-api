package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	// GoodBehaviorCollection the name of the Good Behavior Collection
	GoodBehaviorCollection = "goodBehavior"
)

type GoodBehaviorReport interface {
	GoodBehaviorSave(data *schema.GoodBehaviorData) error
}

// CitizenReportSave save  a record instantly in database
func (m *mongoDB) GoodBehaviorSave(data *schema.GoodBehaviorData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"account_number", data.AccountNumber}, {"ts", data.Timestamp}}
	update := bson.D{{"$set", bson.D{{"behavior_score", data.BehaviorScore}}}}
	c := m.client.Database(m.database)
	updateRes, err := c.Collection(GoodBehaviorCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if updateRes.MatchedCount == 0 {
		_, err := c.Collection(GoodBehaviorCollection).InsertOne(ctx, *data)
		if err != nil {
			return err
		}
	}
	return nil
}
