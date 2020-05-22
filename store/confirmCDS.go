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
	CdsUSA    = "United States"
	CdsTaiwan = "Taiwan"
)

var (
	ErrNoConfirmDataset      = fmt.Errorf("no data-set")
	ErrInvalidConfirmDataset = fmt.Errorf("invalid confirm data-set")
	ErrPoliticalTypeGeoInfo  = fmt.Errorf("no political type geo info")
	ErrConfirmDataFetch      = fmt.Errorf("fetch cds confirm data fail")
	ErrConfirmDecode         = fmt.Errorf("decode confirm data fail")
)

type ConfirmCDS interface {
	ReplaceCDS(result []schema.CDSData, country string) error
	CreateCDS(result []schema.CDSData, country string) error
	GetCDSConfirm(loc schema.Location) (float64, float64, float64, error)
}

var CDSCountyCollectionMatrix = map[CDSCountryType]string{
	CDSCountryType(CdsUSA):    "ConfirmUS",
	CDSCountryType(CdsTaiwan): "ConfirmTaiwan",
}

func (m *mongoDB) ReplaceCDS(result []schema.CDSData, country string) error {
	collection, ok := CDSCountyCollectionMatrix[CDSCountryType(country)]
	if !ok {
		return errors.New("no cds country availible")
	}
	if len(result) <= 0 {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix}).Debug("no record to update")
		return nil
	}

	for _, v := range result {
		filter := bson.M{"name": v.Name, "report_ts": v.ReportTime}
		replacement := bson.M{
			"name":        v.Name,
			"city":        v.City,
			"county":      v.County,
			"state":       v.State,
			"country":     v.Country,
			"level":       v.Level,
			"cases":       v.Cases,
			"deaths":      v.Deaths,
			"recovered":   v.Recovered,
			"report_ts":   v.ReportTime,
			"update_ts":   v.UpdateTime,
			"report_date": v.ReportTimeDate,
			"countryId":   v.CountryID,
			"stateId":     v.StateID,
			"countyId":    v.CountyID,
			"location":    v.Location,
			"tz":          v.Timezone,
		}
		opts := options.Replace().SetUpsert(true)
		_, err := m.client.Database(m.database).Collection(collection).ReplaceOne(context.Background(), filter, replacement, opts)
		if err != nil {
			if errs, hasErr := err.(mongo.BulkWriteException); hasErr {
				if 1 == len(errs.WriteErrors) && DuplicateKeyCode == errs.WriteErrors[0].Code {
					log.WithField("prefix", mongoLogPrefix).Warnf("cds update with error: %s", err)
				}
			}
		}
	}
	return nil
}

func (m *mongoDB) CreateCDS(result []schema.CDSData, country string) error {
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
				log.WithFields(log.Fields{"prefix": mongoLogPrefix, "err": errs}).Warn("createCDSData Insert data")
				return nil
			}
		}
	}
	if res != nil {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "records": len(res.InsertedIDs)}).Info("createCDSData Insert data")
	}
	return nil
}

func (m mongoDB) GetCDSConfirm(location schema.Location) (float64, float64, float64, error) {
	log.WithField("location", location).Info("get cds confirmation by location")
	switch location.Country { //  Currently this function support only USA data
	case CdsTaiwan:
		// use taiwan cdc data (temp solution)
		today, delta, percent, err := m.GetConfirm("tw", location.County)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("%s: %s", ErrConfirmDataFetch, err)
		}
		return today, delta, percent, nil

	case CdsUSA:
		c := m.client.Database(m.database).Collection(CDSCountyCollectionMatrix[CDSCountryType(CdsUSA)])
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		opts := options.Find()
		opts = opts.SetSort(bson.M{"report_ts": -1}).SetLimit(2)
		filter := bson.M{"county": location.County, "state": location.State}
		cur, err := c.Find(context.Background(), filter, opts)
		if nil != err {
			return 0, 0, 0, ErrConfirmDataFetch
		}
		var results []schema.CDSData

		for cur.Next(ctx) {
			var result schema.CDSData
			if errDecode := cur.Decode(&result); errDecode != nil {
				return 0, 0, 0, ErrConfirmDecode
			}
			results = append(results, result)
		}
		percent := float64(100)
		if len(results) >= 2 {
			if results[0].ReportTime > results[1].ReportTime {
				today := results[0]
				yesterday := results[1]
				delta := today.Cases - yesterday.Cases
				if yesterday.Cases > 0 {
					percent = 100 * delta / yesterday.Cases
				}
				log.WithField("prefix", mongoLogPrefix).Debugf("cds score results: today:%f, delta:%f, percent:%f", today.Cases, delta, percent)
				return today.Cases, delta, percent, nil
			} else if 1 == len(results) {
				today := results[0]
				delta := today.Cases
				log.WithField("prefix", mongoLogPrefix).Debugf("cds score results: today:%f, delta:%f, percent:%f", today.Cases, delta, percent)
				return today.Cases, delta, percent, nil
			}
			return 0, 0, 0, ErrInvalidConfirmDataset
		}
	}
	return 0, 0, 0, ErrNoConfirmDataset

}
