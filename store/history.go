package store

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bitmark-inc/autonomy-api/schema"
)

type History interface {
	GetReportedSymptoms(accountNumber string, earierThan, limit int64) ([]*schema.SymptomReportData, error)
	GetReportedBehaviors(accountNumber string, earierThan, limit int64) ([]*schema.GoodBehaviorData, error)
}

func (m *mongoDB) GetReportedSymptoms(accountNumber string, earierThan, limit int64) ([]*schema.SymptomReportData, error) {
	c := m.client.Database(m.database).Collection(schema.SymptomReportCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query, options := historyQuery(accountNumber, earierThan, limit)
	cur, err := c.Find(ctx, query, options)
	if err != nil {
		return nil, err
	}

	reports := make([]*schema.SymptomReportData, 0)
	for cur.Next(ctx) {
		var r schema.SymptomReportData
		if err := cur.Decode(&r); err != nil {
			return nil, err
		}
		reports = append(reports, &r)
	}

	return reports, nil
}

func (m *mongoDB) GetReportedBehaviors(accountNumber string, earierThan, limit int64) ([]*schema.GoodBehaviorData, error) {
	c := m.client.Database(m.database).Collection(schema.GoodBehaviorCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query, options := historyQuery(accountNumber, earierThan, limit)
	cur, err := c.Find(ctx, query, options)
	if err != nil {
		return nil, err
	}

	reports := make([]*schema.GoodBehaviorData, 0)
	for cur.Next(ctx) {
		var r schema.GoodBehaviorData
		if err := cur.Decode(&r); err != nil {
			return nil, err
		}
		reports = append(reports, &r)
	}

	return reports, nil
}

func historyQuery(accountNumber string, earierThan, limit int64) (bson.M, *options.FindOptions) {
	query := bson.M{"account_number": accountNumber}
	if earierThan > 0 {
		query["ts"] = bson.M{"$lt": earierThan}
	}
	options := options.Find()
	options = options.SetSort(bson.M{"ts": -1}).SetLimit(limit)
	return query, options
}
