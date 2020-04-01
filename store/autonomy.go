package store

import (
	"github.com/jinzhu/gorm"

	"github.com/bitmark-inc/autonomy-api/schema"
)

// autonomy main datastore
type AutonomyCore interface {
	Ping() error

	// Account
	CreateAccount(string, string, map[string]interface{}) (*schema.Account, error)
	GetAccount(string) (*schema.Account, error)
	UpdateAccountMetadata(string, map[string]interface{}) error
	UpdateAccountGeoPosition(accountNumber string, latitude, longitude float64) error
	DeleteAccount(string) error

	// Help
	RequestHelp(accountNumber, subject, needs, meetingPlace, contactInfo string) (*schema.HelpRequest, error)
	GetHelp(helpID string) (*schema.HelpRequest, error)
	ListHelps(accountNumber string, latitude, longitude float64, count int64) ([]schema.HelpRequest, error)
	AnswerHelp(accountNumber string, helpID string) error
}

// AutonomyStore is an implementation of AutonomyCore
type AutonomyStore struct {
	ormDB *gorm.DB
	mongo MongoStore
}

func NewAutonomyStore(ormDB *gorm.DB, mongo MongoStore) *AutonomyStore {
	return &AutonomyStore{
		ormDB: ormDB,
		mongo: mongo,
	}
}

// Ping is to check the storage health status
func (s *AutonomyStore) Ping() error {
	return s.ormDB.DB().Ping()
}
