package cdc

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/bitmark-inc/autonomy-api/store"
)

// Taiwan government returns json with chinese key
type twCovid struct {
	Year           string `json:"診斷年份"`
	Week           int    `json:"診斷週別,string"`
	County         string `json:"縣市"`
	Gender         string `json:"性別"`
	Foreign        string `json:"是否為境外移入"`
	Age            string `json:"年齡層"`
	ConfirmedCount int    `json:"確定病例數,string"`
}

type TWCDC struct {
	URL    string
	Result store.ConfirmCountyCount
}

func (t *TWCDC) Run() (int, error) {
	data, err := dataFromURL(t.URL)
	if nil != err {
		return 0, err
	}

	// decode json
	var arr []twCovid
	err = json.Unmarshal(data, &arr)
	if nil != err {
		log.WithFields(log.Fields{
			"prefix":   logPrefix,
			"error":    err,
			"raw json": string(data),
		}).Error("decode json")
		return 0, err
	}

	aggregated, count := aggregateTw(arr)
	t.Result = aggregated
	return count, nil
}

func dataFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if nil != err {
		log.WithFields(log.Fields{
			"prefix": logPrefix,
			"url":    url,
			"error":  err,
		}).Error("get tw cdc daily confirm cases")
		return []byte{}, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		log.WithFields(log.Fields{
			"prefix": logPrefix,
			"error":  err,
		}).Error("read tw cdc confirm case response")
		return []byte{}, err
	}
	return data, nil
}

func aggregateTw(data []twCovid) (map[string]int, int) {
	countyMapping := make(map[string]int)
	count := 0
	for _, d := range data {
		if _, ok := countyMapping[d.County]; !ok {
			countyMapping[d.County] = d.ConfirmedCount
		} else {
			countyMapping[d.County] += d.ConfirmedCount
		}
		count++
	}
	return countyMapping, count
}

// NewTw - new tw cdc crawler
func NewTw(url string) CDC {
	return &TWCDC{
		URL: url,
	}
}
