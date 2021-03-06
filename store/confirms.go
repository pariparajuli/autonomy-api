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
	GetConfirm(string, string) (float64, float64, float64, error)
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

	now := time.Now().UTC().Unix()
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
		var prev schema.Confirm
		err = c.FindOne(ctx, query).Decode(&prev)

		if nil != err {
			log.WithFields(log.Fields{
				"prefix":  mongoLogPrefix,
				"county":  county,
				"profile": ConfirmCollection,
				"error":   err,
			}).Info("create new record of confirm count")

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

			continue
		}

		log.WithFields(log.Fields{
			"prefix":  mongoLogPrefix,
			"country": country,
			"county":  prev.County,
			"count":   prev.Count,
		}).Debug("prev confirm count")

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

func (m mongoDB) GetConfirm(country, county string) (float64, float64, float64, error) {
	c := m.client.Database(m.database).Collection(ConfirmCollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	country = utils.EnNameToKey(country)
	county = utils.EnNameToKey(county)

	var latest schema.Confirm
	err := c.FindOne(ctx, countyCountryQuery(county, country)).Decode(&latest)

	if nil != err {
		log.WithFields(log.Fields{"prefix": mongoLogPrefix, "country": country,
			"county": county, "error": err}).Error("get confirm count from db")

		if err == mongo.ErrNoDocuments {
			log.WithError(err).Error("no documents found")
			return 0, 0, 0, nil
		}
		log.WithError(err).Error("other mongodb errors")
		return RecordNotExistCode, RecordNotExistCode, 0, err
	}

	percent := float64(0)
	if latest.Count != 0 {
		percent = float64(latest.DiffYesterday) / float64(latest.Count)
	}

	log.WithFields(log.Fields{"prefix": mongoLogPrefix, "county": county, "country": country,
		"latest_count": latest.Count, "previous_count": latest.Count - latest.DiffYesterday, "change_percent": percent,
	}).Debug("get confirm data")

	return float64(latest.Count), float64(latest.DiffYesterday), percent, nil
}
