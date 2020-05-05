package aqi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bitmark-inc/autonomy-api/external/aqi"
)

func TestGet(t *testing.T) {
	index := 23
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		type data struct {
			Aqi int `json:"aqi"`
		}

		type resp struct {
			Status string `json:"status"`
			Data   data   `json:"data"`
		}

		r := resp{
			Status: "ok",
			Data: data{
				Aqi: index,
			},
		}

		b, _ := json.Marshal(r)
		_, _ = w.Write(b)
	}))
	defer ts.Close()

	a := aqi.New("test", ts.URL)
	actual, err := a.Get(1.2, 3.4)
	assert.Nil(t, err, "wrong Get")
	assert.Equal(t, index, actual, "wrong aqi index")
}
