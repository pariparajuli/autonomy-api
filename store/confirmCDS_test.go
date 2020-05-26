package store

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bitmark-inc/autonomy-api/external/mocks"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/utils"
)

type ConfirmCDSTestSuite struct {
	suite.Suite
	connURI       string
	testDBName    string
	mongoClient   *mongo.Client
	testDatabase  *mongo.Database
	geoClientMock *mocks.MockGeoInfo
}

func NewConfirmTestSuite(connURI, dbName string) *ConfirmCDSTestSuite {
	return &ConfirmCDSTestSuite{
		connURI:    connURI,
		testDBName: dbName,
	}
}

func (s *ConfirmCDSTestSuite) SetupSuite() {
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

// CleanMongoDB drop the whole test mongodb
func (s *ConfirmCDSTestSuite) CleanMongoDB() error {
	return s.testDatabase.Drop(context.Background())
}

// LoadMongoDBFixtures will preload fixtures into test mongodb
func (s *ConfirmCDSTestSuite) LoadMongoDBFixtures() error {
	return nil
}

func (s *ConfirmCDSTestSuite) LoadConfirmFixtures() ([]schema.CDSData, error) {
	f, err := os.Open("fixtures/confirm_taiwan.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var testData []schema.CDSData
	if err := json.NewDecoder(f).Decode(&testData); err != nil {
		return nil, err
	}
	return testData, nil
}

func (s *ConfirmCDSTestSuite) TestCreateCDS() {
	data, err := s.LoadConfirmFixtures()
	// validate data set
	s.NoError(err)
	s.Equal(29, len(data))
	s.Equal("country", data[0].Level)
	s.Equal(CdsTaiwan, data[0].Name)
	store := NewMongoStore(s.mongoClient, s.testDBName)
	err = store.CreateCDS(data, CdsTaiwan)
	s.NoError(err)
	collection, ok := CDSCountyCollectionMatrix[CDSCountryType(CdsTaiwan)]
	s.True(ok)
	count, err := s.testDatabase.Collection(collection).CountDocuments(context.Background(), bson.M{})
	s.NoError(err)
	s.Equal(int64(29), count)
}



func TestConfirmTestSuite(t *testing.T) {
	suite.Run(t, NewConfirmTestSuite("mongodb://127.0.0.1:27017/?compressors=disabled", "test-db"))
}
