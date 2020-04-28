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
	GetReportedLocations(accountNumber string, earierThan, limit int64) ([]schema.Geographic, error)
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

func (m *mongoDB) GetReportedBehaviors(accountNumber string, earierThan, limit int64) ([]*schema.BehaviorReportData, error) {
	c := m.client.Database(m.database).Collection(schema.BehaviorReportCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query, options := historyQuery(accountNumber, earierThan, limit)
	cur, err := c.Find(ctx, query, options)
	if err != nil {
		return nil, err
	}

	reports := make([]*schema.BehaviorReportData, 0)
	for cur.Next(ctx) {
		var r schema.BehaviorReportData
		if err := cur.Decode(&r); err != nil {
			return nil, err
		}
		reports = append(reports, &r)
	}

	return reports, nil
}

func (m *mongoDB) GetReportedLocations(accountNumber string, earierThan, limit int64) ([]schema.Geographic, error) {
	c := m.client.Database(m.database).Collection(schema.GeographicCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query, options := historyQuery(accountNumber, earierThan, limit)
	cur, err := c.Find(ctx, query, options)
	if err != nil {
		return nil, err
	}

	result := make([]schema.Geographic, 0)
	for cur.Next(ctx) {
		var g schema.Geographic
		if err = cur.Decode(&g); err != nil {
			return nil, err
		}
		result = append(result, g)
	}

	return result, nil
}

func historyQuery(accountNumber string, earierThan, limit int64) (bson.M, *options.FindOptions) {
	query := bson.M{
		"account_number": accountNumber,
		"ts":             bson.M{"$lt": earierThan},
	}
	options := options.Find()
	options = options.SetSort(bson.M{"ts": -1}).SetLimit(limit)
	return query, options
}
