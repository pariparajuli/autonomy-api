package geoinfo

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"

	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	logPrefix      = "geoinfo"
	defaultTimeout = 5 * time.Second
)

// GeoInfo - interface to operate google maps
type GeoInfo interface {
	Get(schema.Location) ([]maps.GeocodingResult, error)
}

type geoInfo struct {
	client *maps.Client
}

// latLng - a string representation of a Lat,Lng pair, e.g. 1.23,4.56
func (g geoInfo) Get(loc schema.Location) ([]maps.GeocodingResult, error) {
	log.WithFields(log.Fields{
		"prefix": logPrefix,
		"lat":    loc.Latitude,
		"lng":    loc.Longitude,
	}).Info("query geo info")

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return g.client.Geocode(ctx, &maps.GeocodingRequest{LatLng: &maps.LatLng{
		Lat: loc.Latitude,
		Lng: loc.Longitude,
	}})
}

// New - new GeoInfo interface
func New(apiKey string) (GeoInfo, error) {
	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		log.WithFields(log.Fields{
			"prefix": logPrefix,
			"error":  err,
		}).Error("new map client")

		return nil, err
	}

	return &geoInfo{
		client: client,
	}, nil
}
