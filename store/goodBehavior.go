package store

import (
	"context"
	"time"

	"github.com/bitmark-inc/autonomy-api/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DuplicateKeyCode = 11000
)

// GoodBehaviorReport save a GoodBehaviorData into Database
type GoodBehaviorReport interface {
	GoodBehaviorSave(data *schema.GoodBehaviorData) error
}

// GoodBehaviorData save a GoodBehaviorData into mongoDB
func (m *mongoDB) GoodBehaviorSave(data *schema.GoodBehaviorData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c := m.client.Database(m.database)
	_, err := c.Collection(schema.GoodBehaviorCollection).InsertOne(ctx, *data)
	we, hasErr := err.(mongo.WriteException)
	if hasErr {
		if 1 == len(we.WriteErrors) && DuplicateKeyCode == we.WriteErrors[0].Code {
			return nil
		}
		return err
	}
	return nil
}
