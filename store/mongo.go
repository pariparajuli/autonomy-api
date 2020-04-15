package store

import (
	"context"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	mongoLogPrefix = "mongo"
)

// MongoStore - interface for mongodb operations
type MongoStore interface {
	Group
	Healthier
	MongoAccount
	CitizenReport
	POI
	Closer
	Pinger
	ConfirmOperator
}

// MongoAccount - account related operations
// mongo version create account is different from postgres
type MongoAccount interface {
	CreateAccount(*schema.Account) error
	CreateAccountWithGeoPosition(*schema.Account, float64, float64) error
	UpdateAccountGeoPosition(string, float64, float64) error
	DeleteAccount(string) error
	UpdateAccountScore(string, float64) error
	IsAccountExist(string) (bool, error)
	AppendPOIForAccount(accountNumber string, desc *schema.POIDesc) error
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
	client   *mongo.Client
	database string
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
func NewMongoStore(client *mongo.Client, database string) MongoStore {
	return &mongoDB{
		client:   client,
		database: database,
	}
}
