package store

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/bitmark-inc/autonomy-api/schema"
)

// autonomy main datastore
type Autonomy interface {
	Ping() error
	CreateAccount(string, string, map[string]interface{}) (*schema.Account, error)
	GetAccount(string) (*schema.Account, error)
	DeleteAccount(string) error
}

type ORMStore struct {
	ormDB *gorm.DB
}

func NewORMStore(ormDB *gorm.DB) *ORMStore {
	return &ORMStore{
		ormDB: ormDB,
	}
}

func (s *ORMStore) Ping() error {
	return s.ormDB.DB().Ping()
}

func (s *ORMStore) CreateAccount(accountNumber, encPubKey string, metadata map[string]interface{}) (*schema.Account, error) {
	a := schema.Account{
		AccountNumber: accountNumber,
		EncPubKey:     encPubKey,
		Profile: schema.AccountProfile{
			AccountNumber: accountNumber,
			State: schema.ActivityState{
				LastActiveTime: time.Now(),
			},
			Metadata: schema.AccountMetadata(metadata),
		},
	}

	if err := s.ormDB.Create(&a).Error; err != nil {
		return nil, err
	}

	return &a, nil
}

func (s *ORMStore) GetAccount(accountNumber string) (*schema.Account, error) {
	var a schema.Account
	if err := s.ormDB.Preload("Profile").Where("account_number = ?", accountNumber).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *ORMStore) DeleteAccount(accountNumber string) error {
	if err := s.ormDB.Delete(schema.Account{}, "account_number = ?", accountNumber).Error; err != nil {
		return err
	}

	if err := s.ormDB.Delete(schema.AccountProfile{}, "account_number = ?", accountNumber).Error; err != nil {
		return err
	}

	return nil
}
