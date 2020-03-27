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
	DeleteAccount(string) error
}

// AutonomyStore is an implementation of AutonomyCore
type AutonomyStore struct {
	ormDB *gorm.DB
}

func NewAutonomyStore(ormDB *gorm.DB) *AutonomyStore {
	return &AutonomyStore{
		ormDB: ormDB,
	}
}

// Ping is to check the storage health status
func (s *AutonomyStore) Ping() error {
	return s.ormDB.DB().Ping()
}
