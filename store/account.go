package store

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/bitmark-inc/autonomy-api/schema"
)

// CreateAccount is to register an account into autonomy system
func (s *AutonomyStore) CreateAccount(accountNumber, encPubKey string, metadata map[string]interface{}) (*schema.Account, error) {
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

// GetAccount returns an account instance of a given account number
func (s *AutonomyStore) GetAccount(accountNumber string) (*schema.Account, error) {
	var a schema.Account
	if err := s.ormDB.Preload("Profile").Where("account_number = ?", accountNumber).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// UpdateAccountMetadata is to update metadata for a specific account
func (s *AutonomyStore) UpdateAccountMetadata(accountNumber string, metadata map[string]interface{}) error {
	var a schema.Account
	if err := s.ormDB.Preload("Profile").Where("account_number = ?", accountNumber).First(&a).Error; err != nil {
		return err
	}

	for k, v := range metadata {
		a.Profile.Metadata[k] = v
	}

	return s.ormDB.Save(&a.Profile).Error
}

// UpdateAccountMetadata is to update metadata for a specific account
func (s *AutonomyStore) UpdateAccountGeoPosition(accountNumber string, latitude, longitude float64) error {
	var a schema.Account
	if err := s.ormDB.Preload("Profile").Where("account_number = ?", accountNumber).First(&a).Error; err != nil {
		return err
	}

	a.Profile.State.LastLocation = &schema.Location{
		Latitude:  latitude,
		Longitude: longitude,
	}

	return s.ormDB.Save(&a.Profile).Error
}

// DeleteAccount removes an account from our system permanently
func (s *AutonomyStore) DeleteAccount(accountNumber string) error {
	if err := s.ormDB.Delete(schema.Account{}, "account_number = ?", accountNumber).Error; err != nil {
		return err
	}

	if err := s.ormDB.Delete(schema.AccountProfile{}, "account_number = ?", accountNumber).Error; err != nil {
		return err
	}

	return nil
}

func (m *mongoDB) CreateAccount(a *schema.Account) error {
	c := m.client.Database(m.database).Collection(schema.ProfileCollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	p := schema.Profile{
		ID:            a.ProfileID.String(),
		AccountNumber: a.AccountNumber,
		HealthScore:   0,
	}

	log.WithField("prefix", mongoLogPrefix).Debug("account profile")

	result, err := c.InsertOne(ctx, p)
	if nil != err {
		log.WithField("prefix", mongoLogPrefix).Errorf("create mongo account error: %s", err)
		return err
	}

	log.WithField("prefix", mongoLogPrefix).Infof("create mongo account result: %v", result)
	return nil
}

func (m *mongoDB) UpdateAccountGeoPosition(accountNumber string, latitude, longitude float64) error {
	c := m.client.Database(m.database).Collection(schema.ProfileCollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	log.WithField("prefix", mongoLogPrefix).Debugf("update account %s to new geolocation: long %f, lat %f", accountNumber, longitude, latitude)

	query := bson.M{
		"account_number": bson.M{
			"$eq": accountNumber,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"location": bson.M{
				"type": "Point",
				"coordinates": bson.A{
					longitude, latitude,
				},
			},
		},
	}

	result, err := c.UpdateMany(ctx, query, update)
	if nil != err {
		log.WithField("prefix", mongoLogPrefix).Errorf("update account %s geolocation with error: %s", accountNumber, err)
		return err
	}

	log.WithField("prefix", mongoLogPrefix).Debugf("update mongo account geolocation result: %v", result)

	return nil
}
