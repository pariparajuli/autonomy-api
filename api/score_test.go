package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/bitmark-inc/autonomy-api/api/mocks"
	"github.com/bitmark-inc/autonomy-api/schema"
)

func TestScore(t *testing.T) {
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
		EncPubKey:     "2",
		Profile: schema.AccountProfile{
			State: schema.ActivityState{
				LastLocation: &schema.Location{
					Latitude:  1.23,
					Longitude: 4.56,
				},
			},
		},
	}, nil).Times(1)

	ids := []string{"a", "b", "c", "d"}
	var score float64 = 77.8899

	m.EXPECT().NearestCount(gomock.Any(), gomock.Any()).Return(ids, nil).Times(1)
	m.EXPECT().Health(ids).Return(score, nil).Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(s.recognizeAccountMiddleware())
	router.GET("/", s.score)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "wrong status code")

	var jResp map[string]float64
	err := json.Unmarshal([]byte(w.Body.String()), &jResp)

	assert.Nil(t, err, "wrong json unmarshal")
	assert.Equal(t, score, jResp["score"], "wrong json response of score")
}
