package store

import (
	"context"
	"github.com/bitmark-inc/autonomy-api/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

const (
	// GoodBehaviorCollection the name of the Good Behavior Collection
	GoodBehaviorCollection = "goodBehavior"
	DuplicateKeyCode       = 11000
)

type GoodBehaviorReport interface {
	GoodBehaviorSave(data *schema.GoodBehaviorData) error
}

func (m *mongoDB) GoodBehaviorSave(data *schema.GoodBehaviorData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c := m.client.Database(m.database)
	_, err := c.Collection(GoodBehaviorCollection).InsertOne(ctx, *data)
	we, hasErr := err.(mongo.WriteException)
	if hasErr {
		if 1 == len(we.WriteErrors) && DuplicateKeyCode == we.WriteErrors[0].Code {
			return nil
		}
		return err
	}
	return nil
}
