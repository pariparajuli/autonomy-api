package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/bitmark-inc/autonomy-api/external/cdc"
	"github.com/bitmark-inc/autonomy-api/store"
)

type twCrawler struct {
	mongoStore store.MongoStore
	country    string
	countryCDC cdc.CDC
}

func (c twCrawler) Run() {
	count, err := c.countryCDC.Run()
	if nil != err {
		log.WithFields(log.Fields{"prefix": logPrefix, "country": c.country, "error": err}).Error("data from CDC")
	}

	cdc, ok := c.countryCDC.(*cdc.TWCDC)
	if ok {
		c.mongoStore.UpdateOrInsertConfirm(cdc.Result, c.country)
		log.WithFields(log.Fields{"prefix": logPrefix, "country": c.country, "count": count}).Debug("data from CDC")
	} else {
		log.WithFields(log.Fields{"prefix": logPrefix, "country": c.country, "count": count}).Error("get TW data  from CDC failed!")
	}
}

// newTWCrawler - new cron job for daily crawler
func newTWCrawler(country string, mongoStore store.MongoStore, c cdc.CDC) Cron {
	return &twCrawler{
		mongoStore: mongoStore,
		country:    country,
		countryCDC: c,
	}
}
