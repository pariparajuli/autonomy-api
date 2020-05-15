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
	ErrNoConfirmDataset       = fmt.Errorf("no data-set")
	ErrInvalidConfirmDataset  = fmt.Errorf("invalid confirm data-set")
	ErrPoliticalTypeGeoInfo   = fmt.Errorf("no political type geo info")
	ErrConfirmDataFetch       = fmt.Errorf("fetch cds confirm data fail")
	ErrConfirmDecode          = fmt.Errorf("decode confirm data fail")
	ErrConfirmDuplicateRecord = fmt.Errorf("confirm data duplicate")
)

type ConfirmCDS interface {
	ReplaceCDS(result []schema.CDSData, country string) error
	CreateCDS(result []schema.CDSData, country string) error
	GetCDSConfirm(loc schema.Location) (float64, float64, float64, error)
	ContinuousDataCDSConfirm(loc schema.Location, num int64, lessEqThan int64) ([]schema.CDSScoreDataSet, int, error)
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
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "records": len(res.InsertedIDs)}).Debug("CreateCDSData Insert data")
	}
	return nil
}

func (m mongoDB) GetCDSConfirm(loc schema.Location) (float64, float64, float64, error) {
	log.WithFields(log.Fields{"prefix": mongoLogPrefix, "country": loc.Country, "lv1": loc.State, "lv2": loc.County}).Debug("GetCDSConfirm geo info")
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	var results []schema.CDSData
	var c *mongo.Collection
	var cur *mongo.Cursor
	switch loc.Country { //  Currently this function support only USA data
	case CdsTaiwan:
		c = m.client.Database(m.database).Collection(CDSCountyCollectionMatrix[CDSCountryType(CdsTaiwan)])
		opts := options.Find().SetSort(bson.M{"report_ts": -1}).SetLimit(2)
		filter := bson.M{}
		curTW, err := c.Find(context.Background(), filter, opts)
		if nil != err {
			log.WithField("prefix", mongoLogPrefix).Errorf("CDS confirm data find  error: %s", err)
			return 0, 0, 0, ErrConfirmDataFetch
		}
		cur = curTW
	case CdsUSA:
		c = m.client.Database(m.database).Collection(CDSCountyCollectionMatrix[CDSCountryType(CdsUSA)])
		opts := options.Find().SetSort(bson.M{"report_ts": -1}).SetLimit(2)
		filter := bson.M{"county": loc.County, "state": loc.State}
		curUSA, err := c.Find(context.Background(), filter, opts)
		if nil != err {
			return 0, 0, 0, ErrConfirmDataFetch
		}
		cur = curUSA
	default:
		return 0, 0, 0, ErrNoConfirmDataset
	}

	for cur.Next(ctx) {
		var result schema.CDSData
		if errDecode := cur.Decode(&result); errDecode != nil {
			return 0, 0, 0, ErrConfirmDecode
		}
		log.WithField("prefix", mongoLogPrefix).Debugf("cds query name: %s date:%s", result.Name, result.ReportTimeDate)
		results = append(results, result)
	}

	var percent, delta float64
	var today, yesterday schema.CDSData
	if len(results) >= 2 {
		if results[0].ReportTime > results[1].ReportTime {
			today = results[0]
			yesterday = results[1]
		} else if results[0].ReportTime < results[1].ReportTime {
			today = results[1]
			yesterday = results[0]
			log.WithField("prefix", mongoLogPrefix).Errorf("cds query result error: name :%v  today:%v date:%v", results[0].Name, results[0].ReportTime, results[1].ReportTime)
		} else {
			return 0, 0, 0, ErrConfirmDuplicateRecord
		}
		delta = today.Cases - yesterday.Cases
		if yesterday.Cases > 0 {
			percent = 100 * delta / yesterday.Cases
		} else if 0 == yesterday.Cases {
			percent = 100
		}
		return today.Cases, delta, percent, nil
	} else if 1 == len(results) {
		today = results[0]
		delta = today.Cases
		percent = 100
		return today.Cases, delta, percent, nil
	}
	return 0, 0, 0, ErrInvalidConfirmDataset
}

func (m mongoDB) ContinuousDataCDSConfirm(loc schema.Location, num int64, lessEqThan int64) ([]schema.CDSScoreDataSet, int, error) {
	log.WithFields(log.Fields{"prefix": mongoLogPrefix, "country": loc.Country, "lv1": loc.State, "lv2": loc.County}).Debug("ContinuousDataCDSConfirm geo info")
	lessEqThan = 1589040000
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var col *mongo.Collection
	var filter bson.M
	var opts *options.FindOptions
	var cur *mongo.Cursor
	switch loc.Country { //  Currently this function support only USA data
	case CdsTaiwan:
		col = m.client.Database(m.database).Collection(CDSCountyCollectionMatrix[CDSCountryType(CdsTaiwan)])
		opts = options.Find().SetSort(bson.M{"report_ts": -1}).SetLimit(num + 1)
		filter = bson.M{}
		if lessEqThan > 0 {
			filter = bson.M{"report_ts": bson.D{{"$lte", lessEqThan}}}
		}
		curTW, err := col.Find(context.Background(), filter, opts)
		if nil != err {
			log.WithField("prefix", mongoLogPrefix).Errorf("CDS confirm data find  error: %s", err)
			return nil, 0, ErrConfirmDataFetch
		}
		cur = curTW
	case CdsUSA:
		col = m.client.Database(m.database).Collection(CDSCountyCollectionMatrix[CDSCountryType(CdsUSA)])
		opts = options.Find().SetSort(bson.M{"report_ts": -1}).SetLimit(num + 1)
		filter := bson.M{"county": loc.County, "state": loc.State}
		if lessEqThan > 0 {
			filter = bson.M{"county": loc.County, "state": loc.State, "report_ts": bson.D{{"$lte", lessEqThan}}}
		}
		curUSA, err := col.Find(context.Background(), filter, opts)
		if nil != err {
			log.WithField("prefix", mongoLogPrefix).Errorf("CDS confirm data find  error: %s", err)
			return nil, 0, ErrConfirmDataFetch
		}
		cur = curUSA
	default:
		return nil, 0, ErrNoConfirmDataset
	}

	var results []schema.CDSScoreDataSet
	cur, err := col.Find(context.Background(), filter, opts)
	if nil != err {
		log.WithField("prefix", mongoLogPrefix).Errorf("%v: %s", ErrConfirmDataFetch, err)
		return nil, 0, ErrConfirmDataFetch
	}
	now := schema.CDSScoreDataSet{}

	for cur.Next(ctx) {
		var result schema.CDSScoreDataSet
		if errDecode := cur.Decode(&result); errDecode != nil {
			log.WithField("prefix", mongoLogPrefix).Errorf("cds Decode with error: %s", errDecode)
			return nil, 0, errDecode
		}
		if len(now.Name) > 0 { // now data is valid
			head := make([]schema.CDSScoreDataSet, 1)
			head[0] = schema.CDSScoreDataSet{Name: now.Name, Cases: now.Cases - result.Cases, ReportTimeDate: now.ReportTimeDate}
			results = append(head, results...)
		}
		now = result
	}
	cur.Close(ctx)
	return results, len(results), nil
}
