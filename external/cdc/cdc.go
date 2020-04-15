package cdc

import "github.com/bitmark-inc/autonomy-api/store"

// CDC - interface to crawl cdc confirmed case
type CDC interface {
	Run() (store.ConfirmCountyCount, error)
}

const (
	logPrefix = "cdc"
)
