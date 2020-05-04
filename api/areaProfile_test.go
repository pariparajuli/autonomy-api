package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bitmark-inc/autonomy-api/api/mocks"
	"github.com/bitmark-inc/autonomy-api/consts"
	"github.com/bitmark-inc/autonomy-api/schema"
)

func TestCurrentAreaProfile(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	a := mocks.NewMockAutonomyCore(ctl)
	m := mocks.NewMockMongoStore(ctl)

	s := Server{
		store:      a,
		mongoStore: m,
	}

	a.EXPECT().GetAccount(gomock.Any()).Return(&schema.Account{
		AccountNumber: "1",
		Profile:       schema.AccountProfile{},
	}, nil).Times(1)

	metric := schema.Metric{
		ConfirmedCount: 1,
		ConfirmedDelta: 2,
		SymptomCount:   3,
		SymptomDelta:   4,
		BehaviorCount:  5,
		BehaviorDelta:  6,
		Score:          55.66,
	}

	m.EXPECT().GetProfile("1").Return(&schema.Profile{
		Metric: metric,
		Location: &schema.GeoJSON{
			Type:        "Point",
			Coordinates: []float64{1.1, 2.2},
		},
	}, nil).Times(1)
	m.EXPECT().NearestGoodBehavior(consts.CORHORT_DISTANCE_RANGE, gomock.Any()).Return(metric.Score, 2.2, int(metric.BehaviorCount), int(metric.BehaviorDelta), nil).Times(1)
	m.EXPECT().NearestSymptomScore(consts.CORHORT_DISTANCE_RANGE, gomock.Any()).Return(metric.Score, 2.2, int(metric.SymptomCount), int(metric.SymptomDelta), nil).Times(1)
	m.EXPECT().GetConfirm(gomock.Any()).Return(int(metric.ConfirmedCount), int(metric.ConfirmedDelta), nil).Times(1)
	m.EXPECT().UpdateProfileMetric("1", gomock.Any()).Return(nil).Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(s.recognizeAccountMiddleware())
	router.GET("/", s.currentAreaProfile)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "wrong status code")

	var jResp schema.Metric
	err := json.Unmarshal([]byte(w.Body.String()), &jResp)

	assert.Nil(t, err, "wrong json unmarshal")
	assert.Equal(t, metric.ConfirmedCount, jResp.ConfirmedCount, "wrong confirm")
	assert.Equal(t, metric.ConfirmedDelta, jResp.ConfirmedDelta, "wrong confirm delta")
	assert.Equal(t, metric.SymptomCount, jResp.SymptomCount, "wrong symptoms")
	assert.Equal(t, metric.SymptomDelta, jResp.SymptomDelta, "wrong symptoms delta")
	assert.Equal(t, metric.BehaviorCount, jResp.BehaviorCount, "wrong behavior")
	assert.Equal(t, metric.BehaviorDelta, jResp.BehaviorDelta, "wrong behavior delta")
	assert.Equal(t, metric.Score, jResp.Score, "wrong score")
}

func TestSingleAreaProfile(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	a := mocks.NewMockAutonomyCore(ctl)
	m := mocks.NewMockMongoStore(ctl)

	s := Server{
		store:      a,
		mongoStore: m,
	}

	a.EXPECT().GetAccount(gomock.Any()).Return(&schema.Account{
		AccountNumber: "1",
		Profile:       schema.AccountProfile{},
	}, nil).Times(1)

	poiID, _ := primitive.ObjectIDFromHex("5e8bf47a0ff4f2d27df71bb5")

	metric := schema.Metric{
		ConfirmedCount: 9,
		ConfirmedDelta: 8,
		SymptomCount:   7,
		SymptomDelta:   6,
		BehaviorCount:  5,
		BehaviorDelta:  4,
		Score:          3,
	}

	m.EXPECT().GetPOIMetrics(poiID).Return(&metric, nil).Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(s.recognizeAccountMiddleware())
	router.GET("/:poiID", s.singleAreaProfile)

	req := httptest.NewRequest("GET", "/5e8bf47a0ff4f2d27df71bb5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "wrong status code")

	var jResp schema.Metric
	err := json.Unmarshal([]byte(w.Body.String()), &jResp)

	assert.Nil(t, err, "wrong json unmarshal")
	assert.Equal(t, metric, jResp, "wrong data")
}
