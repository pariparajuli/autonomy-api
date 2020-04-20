package store

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitmark-inc/autonomy-api/external/geoinfo"
)

const (
	mongoLogPrefix = "mongo"
	defaultTimeout = 5 * time.Second
)

// MongoStore - interface for mongodb operations
type MongoStore interface {
	Group
	Healthier
	MongoAccount
	SymptomReport
	POI
	GoodBehaviorReport
	Closer
	Pinger
	ConfirmOperator
}

// Closer - close db connection
type Closer interface {
	Close()
}

// Pinger - ping database
type Pinger interface {
	Ping() error
}

type mongoDB struct {
	client    *mongo.Client
	database  string
	geoClient geoinfo.GeoInfo
}

// Ping - ping mongo db
func (m mongoDB) Ping() error {
	return m.client.Ping(context.Background(), nil)
}

// Close - close mongo db connections
func (m mongoDB) Close() {
	log.WithField("prefix", mongoLogPrefix).Info("closing mongo db connections")
	_ = m.client.Disconnect(context.Background())
}

// NewMongoStore - return mongo db operations
func NewMongoStore(client *mongo.Client, database string, geoClient geoinfo.GeoInfo) MongoStore {
	return &mongoDB{
		client:    client,
		database:  database,
		geoClient: geoClient,
	}
}
