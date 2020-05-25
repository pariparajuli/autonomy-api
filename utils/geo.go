package utils

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/bitmark-inc/autonomy-api/external/geoinfo"
	"github.com/bitmark-inc/autonomy-api/schema"
)

var ErrGeoClientNotInit = fmt.Errorf("geo location client is not initialized")
var ErrEmptyGeo = fmt.Errorf("empty geo info")

var (
	Taiwan = "Taiwan"
	US     = "United States"
)

var geoClient geoinfo.GeoInfo

type PoliticalGeo struct {
	Country      string
	CountryShort string
	Level1       string
	Level1Short  string
	Level2       string
	Level2Short  string
	Level3       string
	Level3Short  string
}

func InitGeoInfo(apiKey string) {
	c, err := geoinfo.New(apiKey)
	if nil != err {
		log.Panicf("get geo client with error: %s", err)
	}

	geoClient = c
}

func SetGeoClient(c geoinfo.GeoInfo) {
	geoClient = c
}

func PoliticalGeoInfo(loc schema.Location) (schema.Location, error) {
	if loc.Country != "" {
		return loc, nil
	}

	if geoClient == nil {
		return loc, ErrGeoClientNotInit
	}

	ret := PoliticalGeo{}
	geos, err := geoClient.Get(loc)
	if nil != err {
		return loc, err
	}
	if len(geos) == 0 {
		return loc, ErrEmptyGeo
	}
	for _, a := range geos[0].AddressComponents {
		if len(a.Types) > 0 && a.Types[0] == "administrative_area_level_1" {
			ret.Level1 = a.LongName
			ret.Level1Short = a.ShortName
		} else if len(a.Types) > 0 && a.Types[0] == "administrative_area_level_2" {
			ret.Level2 = a.LongName
			ret.Level2Short = a.ShortName
		} else if len(a.Types) > 0 && a.Types[0] == "administrative_area_level_3" {
			ret.Level3 = a.LongName
			ret.Level3Short = a.ShortName
		} else if len(a.Types) > 0 && a.Types[0] == "country" {
			ret.Country = a.LongName
			ret.CountryShort = a.ShortName
		}
	}

	loc.Country = ret.Country
	loc.County = ret.Level2

	switch ret.Country {
	case Taiwan:
		if loc.County == "" {
			loc.County = ret.Level1
		}
	case US:
		loc.State = ret.Level1
	}

	return loc, nil
}
