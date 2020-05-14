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

type PoliticalGeo struct {
	Country      string
	CountryShort string
	Level1       string
	Level1Short  string
	Level2       string
	Level2Short  string
	Level3       string
	Level3Short  string
}

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

func (m mongoDB) GetCDSConfirm(loc schema.Location) (float64, float64, float64, error) {
	pGeo, err := m.politicalGeoInfo(loc)
	if err != nil {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error(ErrPoliticalTypeGeoInfo)
		return 0, 0, 0, ErrPoliticalTypeGeoInfo
	}

	log.WithFields(log.Fields{"prefix": mongoLogPrefix, "country": pGeo.Country, "lv1": pGeo.Level1, "lv2": pGeo.Level2, "lv3": pGeo.Level3}).Debug("political geo address")

	switch pGeo.Country { //  Currently this function support only USA data
	case CdsTaiwan:
		// use taiwan cdc data (temp solution)
		today, delta, percent, err := m.GetConfirm(loc)
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
		filter := bson.M{"county": pGeo.Level2, "state": pGeo.Level1}
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

func (m mongoDB) politicalGeoInfo(loc schema.Location) (PoliticalGeo, error) {
	ret := PoliticalGeo{}
	geos, err := m.geoClient.Get(loc)
	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "error": err}).Error("get geo info")
		return PoliticalGeo{}, err
	}
	if len(geos) == 0 {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "lat": loc.Latitude, "loc": loc.Longitude}).Warn("empty geo info")
		return PoliticalGeo{}, ErrEmptyGeo
	}
	for _, a := range geos[0].AddressComponents {
		if len(a.Types) > 0 && a.Types[0] == "administrative_area_level_1" {
			ret.Level1 = a.LongName
			ret.Level1Short = a.ShortName
		} else if len(a.Types) > 0 && a.Types[0] == "administrative_area_level_2" {
			ret.Level2 = a.LongName
			ret.Level2Short = a.ShortName
		} else if len(a.Types) > 0 && a.Types[0] == "administrative_area_level_3" {
			ret.Level3 = a.LongName
			ret.Level3Short = a.ShortName
		} else if len(a.Types) > 0 && a.Types[0] == "country" {
			ret.Country = a.LongName
			ret.CountryShort = a.ShortName
		}
	}
	return ret, nil
}
