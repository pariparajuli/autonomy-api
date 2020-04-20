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
		Confirm:       1,
		ConfirmDelta:  2,
		Symptoms:      3,
		SymptomsDelta: 4,
		Behavior:      5,
		BehaviorDelta: 6,
		Score:         7,
	}

	m.EXPECT().ProfileMetric("1").Return(&metric, nil).Times(1)

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
	assert.Equal(t, metric, jResp, "wrong data")
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
		Confirm:       9,
		ConfirmDelta:  8,
		Symptoms:      7,
		SymptomsDelta: 6,
		Behavior:      5,
		BehaviorDelta: 4,
		Score:         3,
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
