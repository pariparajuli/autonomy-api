package store

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	geographicUpdateInterval = 5 * time.Minute
)

var (
	invalidEarlier = fmt.Errorf("invalid earlier")
)

// Geographic - operations for account geographic
type Geographic interface {
	AddGeographic(accountNumber string, latitude float64, longitude float64) error
}

func (m *mongoDB) AddGeographic(accountNumber string, latitude float64, longitude float64) error {
	c := m.client.Database(m.database).Collection(schema.GeographicCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := bson.M{
		"account_number": bson.M{
			"$eq": accountNumber,
		},
	}
	opts := options.Find()
	opts = opts.SetSort(bson.M{"ts": -1}).SetLimit(1)

	var g schema.Geographic
	var err error

	if err = c.FindOne(ctx, query).Decode(&g); err != nil && err != mongo.ErrNoDocuments {
		log.WithFields(log.Fields{
			"prefix":         mongoLogPrefix,
			"account_number": accountNumber,
			"error":          err,
		}).Error("account latest geographic")
		return err
	}

	if err == nil {
		log.WithFields(log.Fields{
			"prefix":         mongoLogPrefix,
			"account_number": accountNumber,
			"geographic":     g,
		}).Debug("account latest geographic")
	}

	now := time.Now().UTC()

	// update too fast, ignore those
	if err == nil && now.Sub(time.Unix(g.Timestamp, 0)) < geographicUpdateInterval {
		return nil
	}

	current := schema.Geographic{
		AccountNumber: accountNumber,
		Location: schema.GeoJSON{
			Type:        "Point",
			Coordinates: []float64{longitude, latitude},
		},
		Timestamp: now.Unix(),
	}

	if _, err = c.InsertOne(ctx, current); nil != err {
		log.WithFields(log.Fields{
			"prefix":         mongoLogPrefix,
			"account_number": accountNumber,
			"geographic":     current,
		}).Error("add account latest geographic")
		return err
	}

	return nil
}
