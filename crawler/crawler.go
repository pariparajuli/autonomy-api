package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/bitmark-inc/autonomy-api/external/cdc"
	"github.com/bitmark-inc/autonomy-api/store"
)

type twCDC struct {
	mongoStore store.MongoStore
	country    string
	countryCDC cdc.CDC
}

func (c twCDC) Run() {
	confirmCounts, err := c.countryCDC.Run()
	if nil != err {
		log.WithFields(log.Fields{
			"prefix":  logPrefix,
			"country": c.country,
			"error":   err,
		}).Error("data from CDC")
	}

	log.WithFields(log.Fields{
		"prefix":  logPrefix,
		"country": c.country,
		"data":    confirmCounts,
	}).Debug("data from CDC")

	c.mongoStore.UpdateOrInsertConfirm(confirmCounts, c.country)
}

// newCrawler - new cron job for daily crawler
func newCrawler(country string, mongoStore store.MongoStore, c cdc.CDC) Cron {
	return &twCDC{
		mongoStore: mongoStore,
		country:    country,
		countryCDC: c,
	}
}
