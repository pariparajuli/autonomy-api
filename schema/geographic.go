package schema

const (
	GeographicCollection = "geographic"
)

type Geographic struct {
	AccountNumber string   `bson:"account_number"`
	Location      Location `bson:"location"`
	Timestamp     int64    `bson:"ts"`
}
