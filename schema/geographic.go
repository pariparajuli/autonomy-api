package schema

import (
	"encoding/json"
)

const (
	GeographicCollection = "geographic"
)

type Geographic struct {
	AccountNumber string  `bson:"account_number"`
	Location      GeoJSON `bson:"location"`
	Timestamp     int64   `bson:"ts"`
}

func (g *Geographic) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Location  Location `json:"location"`
		Timestamp int64    `json:"timestamp"`
	}{
		Location:  Location{Longitude: g.Location.Coordinates[0], Latitude: g.Location.Coordinates[1]},
		Timestamp: g.Timestamp,
	})
}
