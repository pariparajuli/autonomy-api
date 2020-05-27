package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bitmark-inc/autonomy-api/consts"
	"github.com/bitmark-inc/autonomy-api/schema"
)

var (
	locationNangangTrainStation = schema.GeoJSON{
		Type:        "Point",
		Coordinates: []float64{121.605387, 25.052616},
	}
	locationSinica = schema.GeoJSON{
		Type:        "Point",
		Coordinates: []float64{121.616002, 25.042959},
	}
	locationBitmark = schema.GeoJSON{
		Type:        "Point",
		Coordinates: []float64{121.611905, 25.061037},
	}
	locationTaipeiTrainStation = schema.GeoJSON{
		Type:        "Point",
		Coordinates: []float64{121.517384, 25.047950},
	}

	tsMay25Morning = time.Date(2020, 5, 25, 9, 0, 0, 0, time.UTC).UTC().Unix()
	tsMay26Morning = time.Date(2020, 5, 26, 9, 0, 0, 0, time.UTC).UTC().Unix()
	tsMay26Evening = time.Date(2020, 5, 26, 17, 0, 0, 0, time.UTC).UTC().Unix()

	// only behavior report #2, #3, and #4 should be taken into account
	// because they are near Bitmark Taipei office and are reported during 2020 May 25
	behaviorReport1 = schema.BehaviorReportData{
		ProfileID: "userA",
		OfficialBehaviors: []schema.Behavior{
			{ID: "clean_hand"},
			{ID: "social_distancing"},
		},
		Location:  locationNangangTrainStation,
		Timestamp: tsMay25Morning,
	}
	behaviorReport2 = schema.BehaviorReportData{
		ProfileID: "userA",
		OfficialBehaviors: []schema.Behavior{
			{ID: "clean_hand"},
			{ID: "social_distancing"},
		},
		Location:  locationNangangTrainStation,
		Timestamp: tsMay26Morning,
	}
	behaviorReport3 = schema.BehaviorReportData{
		ProfileID: "userA",
		OfficialBehaviors: []schema.Behavior{
			{ID: "clean_hand"},
			{ID: "social_distancing"},
		},
		Location:  locationSinica,
		Timestamp: tsMay26Evening,
	}
	behaviorReport4 = schema.BehaviorReportData{
		ProfileID: "userB",
		OfficialBehaviors: []schema.Behavior{
			{ID: "touch_face"},
		},
		CustomizedBehaviors: []schema.Behavior{
			{ID: "new_behavior"},
		},
		Location:  locationBitmark,
		Timestamp: tsMay26Morning,
	}
	behaviorReport5 = schema.BehaviorReportData{
		ProfileID: "userB",
		CustomizedBehaviors: []schema.Behavior{
			{ID: "new_behavior"},
		},
		Location:  locationTaipeiTrainStation,
		Timestamp: tsMay26Evening,
	}
)

type BehaviorTestSuite struct {
	suite.Suite
	connURI      string
	testDBName   string
	mongoClient  *mongo.Client
	testDatabase *mongo.Database
}

func NewBehaviorTestSuite(connURI, dbName string) *BehaviorTestSuite {
	return &BehaviorTestSuite{
		connURI:    connURI,
		testDBName: dbName,
	}
}

func (s *BehaviorTestSuite) SetupSuite() {
	if s.connURI == "" || s.testDBName == "" {
		s.T().Fatal("invalid test suite configuration")
	}

	opts := options.Client().ApplyURI(s.connURI)
	mongoClient, err := mongo.NewClient(opts)
	if nil != err {
		s.T().Fatalf("create mongo client with error: %s", err)
	}

	if err = mongoClient.Connect(context.Background()); nil != err {
		s.T().Fatalf("connect mongo database with error: %s", err.Error())
	}

	s.mongoClient = mongoClient
	s.testDatabase = mongoClient.Database(s.testDBName)

	// make sure the test suite is run with a clean environment
	if err := s.CleanMongoDB(); err != nil {
		s.T().Fatal(err)
	}

	schema.NewMongoDBIndexer(s.connURI, s.testDBName).IndexAll()

	ctx := context.Background()
	if _, err := s.testDatabase.Collection(schema.BehaviorReportCollection).InsertMany(ctx, []interface{}{
		behaviorReport1,
		behaviorReport2,
		behaviorReport3,
		behaviorReport4,
		behaviorReport5,
	}); err != nil {
		s.T().Fatal(err)
	}
}

// CleanMongoDB drop the whole test mongodb
func (s *BehaviorTestSuite) CleanMongoDB() error {
	return s.testDatabase.Drop(context.Background())
}

func (s *BehaviorTestSuite) TestFindNearbyBehaviorDistribution() {
	store := NewMongoStore(s.mongoClient, s.testDBName)

	start := time.Date(2020, 5, 26, 0, 0, 0, 0, time.UTC).UTC().Unix()
	end := time.Date(2020, 5, 26, 24, 0, 0, 0, time.UTC).UTC().Unix()
	distribution, err := store.FindNearbyBehaviorDistribution(
		consts.CORHORT_DISTANCE_RANGE,
		schema.Location{
			Longitude: locationBitmark.Coordinates[0],
			Latitude:  locationBitmark.Coordinates[1],
		}, start, end)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), map[string]int{
		"clean_hand":        2,
		"social_distancing": 2,
		"touch_face":        1,
		"new_behavior":      1,
	}, distribution)
}

func (s *BehaviorTestSuite) TestFindNearbyBehaviorReportTimes() {
	store := NewMongoStore(s.mongoClient, s.testDBName)

	start := time.Date(2020, 5, 26, 0, 0, 0, 0, time.UTC).UTC().Unix()
	end := time.Date(2020, 5, 26, 24, 0, 0, 0, time.UTC).UTC().Unix()
	count, err := store.FindNearbyBehaviorReportTimes(
		consts.CORHORT_DISTANCE_RANGE,
		schema.Location{
			Longitude: locationBitmark.Coordinates[0],
			Latitude:  locationBitmark.Coordinates[1],
		}, start, end)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 3, count)
}

func TestBehaviorTestSuite(t *testing.T) {
	suite.Run(t, NewBehaviorTestSuite("mongodb://127.0.0.1:27017/?compressors=disabled", "test-db"))
}
