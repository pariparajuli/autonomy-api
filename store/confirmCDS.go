package store

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bitmark-inc/autonomy-api/schema"
)

type CDSCountryType string

const (
	CdsUSA         = "United States"
	CdsUSALevel    = "county"
	CdsTaiwan      = "Taiwan"
	CdsTaiwanLevel = "country"
)

type ConfirmCDS interface {
	CreateCDSData(result []schema.CDSData, country string) error
	GetCDSConfirm(loc schema.Location) (float64, float64, float64, error)
}

var CDSCountyCollectionMatrix = map[CDSCountryType]string{
	CDSCountryType(CdsUSA):    "ConfirmUS",
	CDSCountryType(CdsTaiwan): "ConfirmTaiwan",
}

func (m *mongoDB) CreateCDSData(result []schema.CDSData, country string) error {
	collection, ok := CDSCountyCollectionMatrix[CDSCountryType(country)]
	if !ok {
		return errors.New("no cds country availible")
	}
	data := make([]interface{}, len(result))
	for i, v := range result {
		data[i] = v
	}
	opts := options.InsertMany().SetOrdered(false)
	res, err := m.client.Database(m.database).Collection(collection).InsertMany(context.Background(), data, opts)
	if err != nil {
		if errs, hasErr := err.(mongo.BulkWriteException); hasErr {
			if 1 == len(errs.WriteErrors) && DuplicateKeyCode == errs.WriteErrors[0].Code {
				fmt.Println(err)
				return nil
			}
		}
	}
	if res != nil {
		log.WithFields(log.Fields{
			"prefix":  mongoLogPrefix,
			"records": len(res.InsertedIDs),
		}).Info("createCDSData Insert data")
	}
	return nil
}

func (m mongoDB) GetCDSConfirm(loc schema.Location) (float64, float64, float64, error) {
	geos, err := m.geoClient.Get(loc)
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("get geo info")
		return 0, 0, 0, err
	}

	if len(geos) == 0 {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "lat": loc.Latitude, "loc": loc.Longitude}).Warn("empty geo info")
		return 0, 0, 0, ErrEmptyGeo
	}
	info := geos[0] // address_components
	log.WithFields(log.Fields{"prefix": mongoLogPrefix, "lat": loc.Latitude, "loc": loc.Longitude, "info": info}).Debug("geo info")

	var county, country string
	for _, a := range info.AddressComponents {
		if len(a.Types) > 0 && a.Types[0] == "country" {
			country = a.LongName
		} else if len(a.Types) > 0 && a.Types[0] == "administrative_area_level_2" {
			county = a.LongName
		}
	}
	log.WithFields(log.Fields{"prefix": mongoLogPrefix, "country": country, "county": county}).Debug("geo address")

	switch country { //  Currently this function support only USA data
	case CdsTaiwan:
		// use taiwan cdc data (temp solution)
		today, delta, percent, err := m.GetConfirm(loc)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("cds confirm data  error: %s", err)
		}
		return today, delta, percent, nil

	case CdsUSA:
		c := m.client.Database(m.database).Collection(CDSCountyCollectionMatrix[CDSCountryType(CdsUSA)])
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		opts := options.Find()
		opts = opts.SetSort(bson.M{"report_ts": -1}).SetLimit(2)
		filter := bson.M{"county": county}
		cur, err := c.Find(context.Background(), filter, opts)
		if nil != err {
			log.WithField("prefix", mongoLogPrefix).Errorf("CDS confirm data find  error: %s", err)
			return 0, 0, 0, fmt.Errorf("cds confirm data find  error: %s", err)
		}
		var results []schema.CDSData

		for cur.Next(ctx) {
			var result schema.CDSData
			if errDecode := cur.Decode(&result); errDecode != nil {
				log.WithField("prefix", mongoLogPrefix).Errorf("cds Decode with error: %s", errDecode)
				return 0, 0, 0, errDecode
			}
			results = append(results, result)
			log.WithField("prefix", mongoLogPrefix).Debugf("cds query name: %s date:%s", result.Name, result.ReportTimeDate)
		}

		percent := float64(100)
		if len(results) >= 2 {
			if results[0].ReportTime > results[1].ReportTime {
				today := results[0]
				yesterday := results[1]
				delta := today.Cases - yesterday.Cases
				if yesterday.Cases > 0 {
					percent = 100 * delta / float64(yesterday.Cases)
				}
				log.WithField("prefix", mongoLogPrefix).Debugf("cds score results: today:%f, delta:%f, percent:%f", today.Cases, delta, percent)
				return today.Cases, delta, percent, nil
			} else if 1 == len(results) {
				today := results[0]
				delta := today.Cases
				return today.Cases, delta, percent, nil
			}
			return 0, 0, 0, errors.New("no enough data in CdsUS dataset")
		}
	}
	return 0, 0, 0, errors.New("no supported CDS dataset")

}
