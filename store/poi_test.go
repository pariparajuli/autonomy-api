package store

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"googlemaps.github.io/maps"

	"github.com/bitmark-inc/autonomy-api/external/mocks"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/utils"
)

var addedPOIID = primitive.NewObjectID()
var existedPOIID = primitive.NewObjectID()
var duplicatedOriginAlias = "origin POI"

type POITestSuite struct {
	suite.Suite
	connURI       string
	testDBName    string
	mongoClient   *mongo.Client
	testDatabase  *mongo.Database
	geoClientMock *mocks.MockGeoInfo
}

func NewPOITestSuite(connURI, dbName string) *POITestSuite {
	return &POITestSuite{
		connURI:    connURI,
		testDBName: dbName,
	}
}

func (s *POITestSuite) SetupSuite() {
	if s.connURI == "" || s.testDBName == "" {
		s.T().Fatal("invalid test suite configuration")
	}

	opts := options.Client().ApplyURI(s.connURI)
	mongoClient, err := mongo.NewClient(opts)
	if nil != err {
		s.T().Fatalf("create mongo client with error: %s", err)
	}
	ctrl := gomock.NewController(s.T())

	geoClientMock := mocks.NewMockGeoInfo(ctrl)
	utils.SetGeoClient(geoClientMock)

	if err = mongoClient.Connect(context.Background()); nil != err {
		s.T().Fatalf("connect mongo database with error: %s", err.Error())
	}

	s.geoClientMock = geoClientMock
	s.mongoClient = mongoClient
	s.testDatabase = mongoClient.Database(s.testDBName)

	// make sure the test suite is run with a clean environment
	if err := s.CleanMongoDB(); err != nil {
		s.T().Fatal(err)
	}
	schema.NewMongoDBIndexer(s.connURI, s.testDBName).IndexAll()
	if err := s.LoadMongoDBFixtures(); err != nil {
		s.T().Fatal(err)
	}
}

// LoadMongoDBFixtures will preload fixtures into test mongodb
func (s *POITestSuite) LoadMongoDBFixtures() error {
	ctx := context.Background()
	uid, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	if _, err := s.testDatabase.Collection(schema.ProfileCollection).InsertOne(ctx, schema.Profile{
		ID:            uid.String(),
		AccountNumber: "account-test-poi",
		PointsOfInterest: []schema.ProfilePOI{
			{
				ID:    addedPOIID,
				Alias: duplicatedOriginAlias,
			},
		},
	}); err != nil {
		return err
	}

	if _, err := s.testDatabase.Collection(schema.POICollection).InsertOne(ctx, schema.POI{
		ID: addedPOIID,
		Location: &schema.GeoJSON{
			Type:        "Point",
			Coordinates: []float64{120.123, 25.123},
		},
		Country: "Taiwan",
		State:   "",
		County:  "Yilan County",
	}); err != nil {
		return err
	}

	if _, err := s.testDatabase.Collection(schema.POICollection).InsertOne(ctx, schema.POI{
		ID: existedPOIID,
		Location: &schema.GeoJSON{
			Type:        "Point",
			Coordinates: []float64{120.12, 25.12},
		},
		Country: "Taiwan",
		State:   "",
		County:  "Yilan County",
	}); err != nil {
		return err
	}

	return nil
}

// CleanMongoDB drop the whole test mongodb
func (s *POITestSuite) CleanMongoDB() error {
	return s.testDatabase.Drop(context.Background())
}

// LoadGeocodingFixtures load prepared geocoding fixtures
func (s *POITestSuite) LoadGeocodingFixtures() ([]maps.GeocodingResult, error) {
	f, err := os.Open("fixtures/geo_result_union_square.json")
	if err != nil {
		return nil, err
	}
	var fixture struct {
		Results []maps.GeocodingResult `json:"results"`
	}

	if err := json.NewDecoder(f).Decode(&fixture); err != nil {
		return nil, err
	}
	return fixture.Results, nil
}

// TestAddPOIWithNonExistAccount tests adding a new poi by an account which is not existent
func (s *POITestSuite) TestAddPOIWithNonExistAccount() {
	store := NewMongoStore(s.mongoClient, s.testDBName)

	geocoding, err := s.LoadGeocodingFixtures()
	s.NoError(err)

	s.geoClientMock.EXPECT().
		Get(gomock.AssignableToTypeOf(schema.Location{})).
		Return(geocoding, nil)

	poi, err := store.AddPOI("account-not-found-test-poi", "test-poi", "", 120, 25)
	s.EqualError(err, "fail to update poi into profile")
	s.Nil(poi)
}

// TestAddPOI tests adding a new poi normally
func (s *POITestSuite) TestAddPOI() {
	ctx := context.Background()
	store := NewMongoStore(s.mongoClient, s.testDBName)

	geocoding, err := s.LoadGeocodingFixtures()
	s.NoError(err)

	s.geoClientMock.EXPECT().
		Get(gomock.AssignableToTypeOf(schema.Location{})).
		Return(geocoding, nil)

	poi, err := store.AddPOI("account-test-poi", "test-poi", "", 120.1, 25.1)
	s.NoError(err)
	s.Equal("United States", poi.Country)
	s.Equal("New York", poi.State)
	s.Equal("New York County", poi.County)
	s.Equal([]float64{120.1, 25.1}, poi.Location.Coordinates)

	count, err := s.testDatabase.Collection(schema.POICollection).CountDocuments(ctx, bson.M{"_id": poi.ID})
	s.NoError(err)
	s.Equal(int64(1), count)

	count, err = s.testDatabase.Collection(schema.ProfileCollection).CountDocuments(context.Background(), bson.M{
		"account_number":           "account-test-poi",
		"points_of_interest.id":    poi.ID,
		"points_of_interest.alias": "test-poi",
	})
	s.NoError(err)
	s.Equal(int64(1), count)
}

// TestAddExistentPOI tests adding a poi where its coordinates has alreday added by other accounts
// but not in the test account
func (s *POITestSuite) TestAddExistentPOI() {
	ctx := context.Background()
	store := NewMongoStore(s.mongoClient, s.testDBName)

	geocoding, err := s.LoadGeocodingFixtures()
	s.NoError(err)

	s.geoClientMock.EXPECT().
		Get(gomock.AssignableToTypeOf(schema.Location{})).
		Return(geocoding, nil)

	count, err := s.testDatabase.Collection(schema.ProfileCollection).CountDocuments(context.Background(), bson.M{
		"account_number":        "account-test-poi",
		"points_of_interest.id": existedPOIID,
	})
	s.NoError(err)
	s.Equal(int64(0), count)

	poi, err := store.AddPOI("account-test-poi", "test-existent-poi", "", 120.12, 25.12)
	s.NoError(err)
	s.Equal("Taiwan", poi.Country)
	s.Equal("", poi.State)
	s.Equal("Yilan County", poi.County)
	s.Equal([]float64{120.12, 25.12}, poi.Location.Coordinates)

	count, err = s.testDatabase.Collection(schema.POICollection).CountDocuments(ctx, bson.M{"_id": existedPOIID})
	s.NoError(err)
	s.Equal(int64(1), count)

	count, err = s.testDatabase.Collection(schema.ProfileCollection).CountDocuments(context.Background(), bson.M{
		"account_number":           "account-test-poi",
		"points_of_interest.id":    existedPOIID,
		"points_of_interest.alias": "test-existent-poi",
	})
	s.NoError(err)
	s.Equal(int64(1), count)
}

// TestAddDuplicatedPOI tests adding a duplicated poi where its coordinates has alreday existed
// for the test account
func (s *POITestSuite) TestAddDuplicatedPOI() {
	ctx := context.Background()
	store := NewMongoStore(s.mongoClient, s.testDBName)

	geocoding, err := s.LoadGeocodingFixtures()
	s.NoError(err)

	s.geoClientMock.EXPECT().
		Get(gomock.AssignableToTypeOf(schema.Location{})).
		Return(geocoding, nil)

		// poi is not in the profile at beginning
	count, err := s.testDatabase.Collection(schema.ProfileCollection).CountDocuments(context.Background(), bson.M{
		"account_number":        "account-test-poi",
		"points_of_interest.id": addedPOIID,
	})
	s.NoError(err)
	s.Equal(int64(1), count)

	// use a different name to add an added poi
	poi, err := store.AddPOI("account-test-poi", "test-duplicated-add-poi", "", 120.123, 25.123)
	s.NoError(err)
	s.Equal("Taiwan", poi.Country)
	s.Equal("", poi.State)
	s.Equal("Yilan County", poi.County)
	s.Equal([]float64{120.123, 25.123}, poi.Location.Coordinates)

	count, err = s.testDatabase.Collection(schema.POICollection).CountDocuments(ctx, bson.M{"_id": addedPOIID})
	s.NoError(err)
	s.Equal(int64(1), count)

	// the alias of the added poi will not updated
	count, err = s.testDatabase.Collection(schema.ProfileCollection).CountDocuments(context.Background(), bson.M{
		"account_number":           "account-test-poi",
		"points_of_interest.id":    addedPOIID,
		"points_of_interest.alias": duplicatedOriginAlias,
	})
	s.NoError(err)
	s.Equal(int64(1), count)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to s.Run
func TestPOITestSuite(t *testing.T) {
	suite.Run(t, NewPOITestSuite("mongodb://127.0.0.1:27017/?compressors=disabled", "test-db"))
}
