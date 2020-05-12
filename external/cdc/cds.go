package cdc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bitmark-inc/autonomy-api/schema"
	log "github.com/sirupsen/logrus"
)

type CovidSource string

const (
	CDSDaily                  CovidSource = "dailyFile"
	CDSDailyHTTP              CovidSource = "dailyHttp"
	CDSTimeseriesLocationFile CovidSource = "timeSeriesLocationFile"
	CDSTimeseriesByDateFile   CovidSource = "timeSeriesByDateFile"
)

type CDSData struct {
	Name           string         `json:"name" bson:"name"`
	City           string         `json:"city" bson:"city"`
	County         string         `json:"county" bson:"county"`
	State          string         `json:"state" bson:"state"`
	Country        string         `json:"country" bson:"country"`
	Level          string         `json:"level" bson:"level"`
	CountryID      string         `json:"countryId" bson:"countryId"`
	StateID        string         `json:"stateId" bson:"stateId"`
	CountyID       string         `json:"countyId" bson:"countyId"`
	Location       schema.GeoJSON `json:"location" bson:"location"`
	Timezone       []string       `json:"tz" bson:"tz"`
	Cases          float64        `json:"cases" bson:"cases"`
	Deaths         float64        `json:"deaths" bson:"deaths"`
	Recovered      float64        `json:"recovered" bson:"recovered"`
	ReportTime     int64          `json:"report_ts" bson:"report_ts"`
	UpdateTime     int64          `json:"update" , bson:"update"`
	ReportTimeDate string         `json:"report_date" bson:"report_date"`
}

type CDS struct {
	Country     string
	Level       string
	CDSDataType CovidSource
	DataFile    *os.File
	URL         string
	Result      []CDSData
}

func (c *CDS) Run() (int, error) {
	data, err := dataFromURL(c.URL)
	if nil != err {
		return 0, err
	}
	count := 0
	updateRecords := []CDSData{}
	sourceData := make([]interface{}, 0)

	err = json.Unmarshal(data, &sourceData)

	if err != nil {
		fmt.Println("ParseDailyOnline error:", err)
		return 0, err
	}

	for _, value := range sourceData {
		record := CDSData{}
		object := value.(map[string]interface{})
		name, ok := object["name"].(string)
		if ok && len(name) > 0 && strings.Contains(name, c.Country) { // Country
			record.Name = name
		} else {
			continue
		}
		record.City, _ = object["city"].(string)
		record.County, _ = object["county"].(string)
		record.State, _ = object["state"].(string)
		record.CountryID, _ = object["countryId"].(string)
		record.StateID, _ = object["stateId"].(string)
		record.CountyID, _ = object["countyId"].(string)
		record.Level, ok = object["level"].(string)

		if ok && "" == record.Level {
			switch c.Level {
			case "country":
				if "" != record.Country && "" == record.State {
					record.Level = "country"
				}
			case "state":
				if "" != record.State && "" == record.County {
					record.Level = "state"
				}
			case "county":
				if "" != record.County && "" == record.City {
					record.Level = "county"
				}
			case "city":
				record.Level = "city"
			default:
				log.WithFields(log.Fields{"prefix": logPrefix, "name": record.Name}).Warn("data from CDS")
				continue
			}
			log.WithFields(log.Fields{"prefix": logPrefix, "name": record.Name, "level": record.Level}).Warn("empty level set")
		}

		if record.Level != c.Level {
			continue
		}

		coorRaw, ok := object["coordinates"].([]interface{})
		if ok && len(coorRaw) > 0 {
			coortemp := []float64{}
			for _, coorV := range coorRaw {
				coortemp = append(coortemp, coorV.(float64))
			}
			record.Location = schema.GeoJSON{Type: "Point", Coordinates: coortemp}
		} else {
			record.Location = schema.GeoJSON{Type: "Point", Coordinates: []float64{}}
		}

		tzRaw, ok := object["tz"].([]interface{})
		if ok && len(tzRaw) > 0 {
			tztemp := []string{}
			for _, tzV := range tzRaw {
				tztemp = append(tztemp, tzV.(string))
			}
			record.Timezone = tztemp
		} else {
			record.Timezone = []string{}
		}
		record.Cases, ok = object["cases"].(float64)
		if !ok {
			log.WithFields(log.Fields{"prefix": logPrefix, "name": record.Name}).Warn("cast cases fail")
			continue
		}
		record.Deaths, _ = object["deaths"].(float64)
		record.Recovered, _ = object["Recovered"].(float64)
		year, month, day := time.Now().Date()
		dateString := fmt.Sprintf("%s-%.2s-%.2s", year, int(month), day)
		if len(record.Timezone) > 0 {
			location, err := time.LoadLocation(record.Timezone[0])
			if err == nil {
				year, month, day := time.Now().In(location).Date()
				dateString = fmt.Sprintf("%d-%.2d-%.2d", year, int(month), day)
			}
		}
		record.UpdateTime = time.Now().UTC().Unix()
		record.ReportTime = time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix()
		record.ReportTimeDate = dateString //In local time
		count++
		updateRecords = append(updateRecords, record)
	}
	c.Result = updateRecords

	return count, nil
}

func getCDSJSON(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if nil != err {
		log.WithFields(log.Fields{"prefix": logPrefix, "url": url, "error": err}).Error("get cds daily json")
		return []byte{}, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		log.WithFields(log.Fields{"prefix": logPrefix, "error": err}).Error("read cds daily json  response")
		return []byte{}, err
	}
	return data, nil
}

// NewCDS - new tw cdc crawler
func NewCDS(country string, level string, dataType CovidSource, f *os.File, url string) CDC {
	return &CDS{
		Country:     country,
		Level:       level,
		CDSDataType: dataType,
		DataFile:    f,
		URL:         url,
	}
}
