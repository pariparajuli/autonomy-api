package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	CitizenReportCollection = "citizenReport"
)

type CitizenReport interface {
	CitizenReportSave(data *schema.CitizenReportData) error
}

// CitizenReportSave save  a record instantly in database
func (m *mongoDB) CitizenReportSave(data *schema.CitizenReportData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{{"account_number", data.AccountNumber}, {"ts", data.Timestamp}}
	update := bson.D{{"$set", bson.D{{"health_score", data.HealthScore}}}}
	c := m.client.Database(m.database)
	updateRes, err := c.Collection(CitizenReportCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if updateRes.MatchedCount == 0 {
		_, err := c.Collection(CitizenReportCollection).InsertOne(ctx, *data)
		if err != nil {
			return err
		}
	}
	return nil
}
