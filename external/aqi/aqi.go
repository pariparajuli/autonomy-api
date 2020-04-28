package aqi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	defaultURL    = "https://api.waqi.info/feed"
	indexNotFound = -1
	statusOK      = "ok"
)

var (
	errResponseStatus = fmt.Errorf("response status no ok")
	errEmptyToken     = fmt.Errorf("empty token")
)

type AQI interface {
	Get(lat, lng float64) (int, error)
}

type aqi struct {
	token string
	url   string
}

type responseData struct {
	Aqi int `json:"aqi"`
}

type jsonResponse struct {
	Status string       `json:"status"`
	Data   responseData `json:"data"`
}

func (a aqi) Get(lat, lng float64) (int, error) {
	if a.token == "" {
		return indexNotFound, errEmptyToken
	}

	// https://api.waqi.info/feed/geo:1.2;3.4/?token=xxxx
	query := fmt.Sprintf("%s/geo:%f;%f/?token=%s", a.url, lat, lng, a.token)
	resp, err := http.Get(query)
	if nil != err {
		return indexNotFound, err
	}

	d, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return indexNotFound, err
	}
	defer resp.Body.Close()

	var r jsonResponse
	err = json.Unmarshal(d, &r)
	if nil != err {
		return indexNotFound, err
	}

	if r.Status != statusOK {
		return indexNotFound, errResponseStatus
	}

	return r.Data.Aqi, nil
}

func New(token string, url string) AQI {
	u := defaultURL
	if url != "" {
		u = url
	}

	return &aqi{
		token: token,
		url:   u,
	}
}
