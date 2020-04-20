package store

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bitmark-inc/autonomy-api/schema"
)

// MongoAccount - account related operations
// mongo version create account is different from postgres
type MongoAccount interface {
	CreateAccount(*schema.Account) error
	CreateAccountWithGeoPosition(*schema.Account, float64, float64) error
	UpdateAccountGeoPosition(string, float64, float64) error

	DeleteAccount(string) error
	UpdateAccountScore(string, float64) error
	IsAccountExist(string) (bool, error)
	AppendPOIToAccountProfile(accountNumber string, desc *schema.POIDesc) error
	RefreshAccountState(accountNumber string) (bool, error)
	GetAccountsByPOI(id string) ([]string, error)

	UpdateProfileMetric(accountNumber string, metric schema.Metric) error
	ProfileMetric(accountNumber string) (*schema.Metric, error)
}

var (
	errAccountNotFound = fmt.Errorf("account not found")
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
	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
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
	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
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
	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
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
	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
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
	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	count, err := c.CountDocuments(ctx, bson.M{"account_number": accountNumber})
	if nil != err {
		log.WithField("prefix", mongoLogPrefix).Errorf("query account number %s with error: %v", accountNumber, err)

		return false, err
	}
	return count > 0, nil
}

// UpdateAccountScore update a score of an account
func (m *mongoDB) UpdateAccountScore(accountNumber string, score float64) error {
	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
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
		log.WithField("prefix", mongoLogPrefix).Errorf("update score of account:%s  error: %s", accountNumber, err)
		return err
	}

	log.WithField("prefix", mongoLogPrefix).Debugf("update score of an account:%s result: %v", accountNumber, result)

	return nil
}

// AppendPOIToAccountProfile appends a POI to end of the POI list of an account
func (m *mongoDB) AppendPOIToAccountProfile(accountNumber string, desc *schema.POIDesc) error {
	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if desc == nil {
		return fmt.Errorf("poi desc not available")
	}

	prefixedLog := log.WithField("prefix", mongoLogPrefix).WithField("account", accountNumber).WithField("poi_id", desc.ID)

	// don't do anything if the POI was added before
	query := bson.M{
		"account_number":        accountNumber,
		"points_of_interest.id": desc.ID,
	}
	count, err := c.CountDocuments(ctx, query)
	if err != nil {
		prefixedLog.WithError(err).Error("unable to query POI for this account")
		return err
	}
	if count > 0 {
		return nil
	}

	// add the POI if not added before
	query = bson.M{"account_number": bson.M{"$eq": accountNumber}}
	update := bson.M{"$push": bson.M{"points_of_interest": bson.M{
		"id":      desc.ID,
		"alias":   desc.Alias,
		"address": desc.Address,
	}}}
	if _, err := c.UpdateOne(ctx, query, update); nil != err {
		prefixedLog.WithError(err).Error("unable to add POI for this account")
		return err
	}

	return nil
}

func (m *mongoDB) UpdateProfileMetric(accountNumber string, metric schema.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
	query := bson.M{
		"account_number": bson.M{
			"$eq": accountNumber,
		},
	}
	update := bson.M{
		"$set": bson.M{
			"metric": metric,
		},
	}

	result, err := c.UpdateOne(ctx, query, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errAccountNotFound
	}

	return nil
}

func (m *mongoDB) GetAccountsByPOI(id string) ([]string, error) {
	log.WithField("prefix", mongoLogPrefix).Debugf("get accounts by POI id: %s", id)

	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	poiID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.WithField("prefix", mongoLogPrefix).Errorf("incorrect POI id:%s  error: %s", id, err.Error())
		return nil, err
	}

	query := bson.M{"points_of_interest.id": poiID}

	filter := bson.M{
		"account_number": 1,
	}

	cursor, err := c.Find(ctx, query, options.Find().SetProjection(filter))

	accounts := []string{}

	for cursor.Next(ctx) {
		// Declare a result BSON object
		var profile struct {
			AccountNumber string `bson:"account_number"`
		}
		if err := cursor.Decode(&profile); err != nil {
			log.WithField("prefix", mongoLogPrefix).Errorf("fail to query accounts include POI id:%s  error: %s", id, err.Error())
			return nil, err
		}

		accounts = append(accounts, profile.AccountNumber)
	}

	return accounts, nil
}

// RefreshAccountState checks current states of a specific account
// and return true if the score has changed
func (m mongoDB) RefreshAccountState(accountNumber string) (bool, error) {
	currentMetric, err := m.ProfileMetric(accountNumber)
	if err != nil {
		return false, err
	}

	// User current metric as new metric
	newMetric := currentMetric

	if err := m.UpdateProfileMetric(accountNumber, *newMetric); err != nil {
		return false, err
	}

	return true, nil
}
