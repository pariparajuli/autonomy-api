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

	err := s.ormDB.Save(&a.Profile).Error
	if nil != err {
		return err
	}

	// if mongo db has no record, create new account with geolocation data
	exist, err := s.mongo.IsAccountExist(a.AccountNumber)
	if nil != err {
		return err
	}

	if exist {
		err = s.mongo.UpdateAccountGeoPosition(a.AccountNumber, latitude, longitude)
		return err
	}

	return s.mongo.CreateAccountWithGeoPosition(&a, latitude, longitude)
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
		HealthScore:   100,
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

func (m *mongoDB) CreateAccountWithGeoPosition(a *schema.Account, latitude, longitude float64) error {
	c := m.client.Database(m.database).Collection(schema.ProfileCollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	p := schema.Profile{
		ID:            a.ProfileID.String(),
		AccountNumber: a.AccountNumber,
		HealthScore:   100,
		Location: &schema.GeoJSON{
			Type:        "Point",
			Coordinates: []float64{longitude, latitude},
		},
	}

	log.WithField("prefix", mongoLogPrefix).Debugf("create account profile: %v", p)

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

func (m *mongoDB) DeleteAccount(accountNumber string) error {
	c := m.client.Database(m.database).Collection(schema.ProfileCollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	result, err := c.DeleteOne(ctx, bson.M{"account_number": accountNumber})
	if nil != err {
		log.WithField("prefix", mongoLogPrefix).Errorf("delete mongo account %s with error: %s", accountNumber, err)
		return err
	}

	log.WithField("prefix", mongoLogPrefix).Infof("delete account number %s success with result: %v", accountNumber, result)

	return nil
}

// IsAccountExist check if account number exist in mongo db
func (m *mongoDB) IsAccountExist(accountNumber string) (bool, error) {
	c := m.client.Database(m.database).Collection(schema.ProfileCollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	count, err := c.CountDocuments(ctx, bson.M{"account_number": accountNumber})
	if nil != err {
		log.WithField("prefix", mongoLogPrefix).Errorf("query account number %s with error: ", accountNumber, err)
		return false, err
	}
	return count > 0, nil
}

// UpdateAccountScore update a score of an account
func (m *mongoDB) UpdateAccountScore(accountNumber string, score float64) error {
	c := m.client.Database(m.database).Collection(schema.ProfileCollectionName)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	log.WithField("prefix", mongoLogPrefix).Debugf("update account %s with new score %f", accountNumber, score)

	query := bson.M{
		"account_number": bson.M{
			"$eq": accountNumber,
		},
	}
	update := bson.M{"$set": bson.M{"health_score": score}}
	result, err := c.UpdateMany(ctx, query, update)
	if nil != err {
		log.WithField("prefix", mongoLogPrefix).Errorf("update score of acount:%s  error: %s", accountNumber, err)
		return err
	}

	log.WithField("prefix", mongoLogPrefix).Debugf("update score of an acount:%s result: %v", accountNumber, result)

	return nil
}
