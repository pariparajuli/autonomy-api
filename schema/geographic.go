package schema

const (
	GeographicCollection = "geographic"
)

type Geographic struct {
	AccountNumber string   `json:"-" bson:"account_number"`
	Location      Location `json:"location" bson:"location"`
	Timestamp     int64    `json:"timestamp" bson:"ts"`
}
