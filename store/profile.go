package store

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

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
