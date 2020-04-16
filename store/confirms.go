package store

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/utils"
)

const (
	ConfirmCollection  = "confirm"
	RecordNotExistCode = -1
)

type ConfirmCountyCount map[string]int

type ConfirmUpdater interface {
	UpdateOrInsertConfirm(confirms ConfirmCountyCount, country string)
}

type ConfirmGetter interface {
	// Read confirm count, return with current count, number of difference compare to
	// yesterday, error
	GetConfirm(country, county string) (int, int, error)

	// total confirm of a country, return with latest total, previous day total, error
	TotalConfirm(country string) (int, int, error)
}

type ConfirmOperator interface {
	ConfirmUpdater
	ConfirmGetter
}

// country code should comes from https://countrycode.org/, with lower case
func (m mongoDB) UpdateOrInsertConfirm(confirms ConfirmCountyCount, country string) {
	c := m.client.Database(m.database).Collection(ConfirmCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	now := time.Now().Unix()
	for county, count := range confirms {
		county, err := utils.TwCountyKey(county)
		if err != nil {
			log.WithFields(log.Fields{
				"prefix": mongoLogPrefix,
				"county": county,
				"error":  err,
			}).Error("convert county name")
			continue
		}

		query := countyCountryQuery(county, country)
		cur, err := c.Find(ctx, query)
		if nil != err {
			log.WithFields(log.Fields{
				"prefix":  mongoLogPrefix,
				"county":  county,
				"profile": ConfirmCollection,
				"error":   err,
			}).Error("retrieve confirm count")
		}

		var prev schema.Confirm

		if cur.Next(ctx) {
			// update existing
			err = cur.Decode(&prev)
			if nil != err {
				log.WithFields(log.Fields{
					"prefix":  mongoLogPrefix,
					"county":  county,
					"country": country,
				}).Debug("convert db record")
				continue
			}

			diff := count - prev.Count

			// confirm count should be same or increased
			if diff < 0 {
				log.WithFields(log.Fields{
					"prefix":        mongoLogPrefix,
					"prev count":    prev.Count,
					"current count": count,
				}).Error("expect daily confirm case to be mono increasing")
				continue
			}

			// no new data
			if diff == 0 {
				log.WithFields(log.Fields{
					"prefix":  mongoLogPrefix,
					"county":  county,
					"count":   count,
					"country": country,
				}).Debug("same confirm count")
			}

			// should only exist one record
			_, err = c.UpdateOne(ctx, query, countUpdateCommand(count, diff, now))
			if nil != err {
				log.WithFields(log.Fields{
					"prefix": mongoLogPrefix,
					"county": county,
					"count":  count,
					"error":  err,
				}).Error("update confirm count")
			}
		} else {
			// insert new
			latest := schema.Confirm{
				County:        county,
				Count:         count,
				Country:       country,
				UpdateTime:    now,
				DiffYesterday: 0,
			}

			_, err = c.InsertOne(ctx, latest)
			if nil != err {
				log.WithFields(log.Fields{
					"prefix": mongoLogPrefix,
					"error":  err,
				}).Error("insert confirm count")
			}
		}
	}
}

func countyCountryQuery(county, country string) bson.M {
	return bson.M{
		"country": country,
		"county":  county,
	}
}

func countUpdateCommand(count, diff int, updateTime int64) bson.M {
	return bson.M{
		"$set": bson.M{
			"count":          count,
			"diff_yesterday": diff,
			"update_time":    updateTime,
		},
	}
}

func (m mongoDB) GetConfirm(country, county string) (int, int, error) {
	c := m.client.Database(m.database).Collection(ConfirmCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var latest schema.Confirm
	err := c.FindOne(ctx, countyCountryQuery(county, country)).Decode(&latest)

	if nil != err {
		log.WithFields(log.Fields{
			"prefix":  mongoLogPrefix,
			"country": country,
			"error":   err,
		}).Error("get confirm count")
		return RecordNotExistCode, RecordNotExistCode, err
	}
	return latest.Count, latest.DiffYesterday, nil
}

func (m mongoDB) TotalConfirm(country string) (int, int, error) {
	c := m.client.Database(m.database).Collection(ConfirmCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cur, err := c.Aggregate(ctx, mongo.Pipeline{bson.D{
		{"$match", bson.D{
			{"country", country},
		}},
	}, bson.D{
		{"$group", bson.D{
			{"_id", "_id"},
			{"total", bson.D{{"$sum", "$count"}}},
			{"diff", bson.D{{"$sum", "$diff_yesterday"}}},
		}},
	}})
	if nil != err {
		log.WithFields(log.Fields{
			"prefix": mongoLogPrefix,
			"error":  err,
		}).Error("aggregate total confirm")
		return RecordNotExistCode, RecordNotExistCode, err
	}

	var results []bson.M
	err = cur.All(ctx, &results)
	if nil != err {
		log.WithFields(log.Fields{
			"prefix": mongoLogPrefix,
			"error":  err,
		}).Error("aggregate total confirm")
		return RecordNotExistCode, RecordNotExistCode, err
	}
	total := results[0]["total"].(int)
	diff := results[0]["diff"].(int)

	if total == 0 {
		log.WithFields(log.Fields{
			"prefix":  mongoLogPrefix,
			"country": country,
		}).Info("empty records")

		return 0, 0, nil
	}

	prev := total - diff

	log.WithFields(log.Fields{
		"prefix":       mongoLogPrefix,
		"country":      country,
		"total":        total,
		"previous day": prev,
	}).Info("total confirm")

	return total, prev, nil
}
