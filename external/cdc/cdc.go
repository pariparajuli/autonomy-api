package cdc

// CDC - interface to crawl cdc confirmed case
type CDC interface {
	Run() (int, error)
}

const (
	logPrefix = "cdc"
)
