package geo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"googlemaps.github.io/maps"

	"github.com/bitmark-inc/autonomy-api/schema"
)

var (
	ErrNoGeoInfoFound         = fmt.Errorf("no geo information found")
	ErrResolverNotInitialized = fmt.Errorf("location resolver is not initialized")
)

var (
	US = "United States"
)

// LocationResolver - interface for resolving location
type LocationResolver interface {
	GetPoliticalInfo(schema.Location) (schema.Location, error)
}

var defaultResolver LocationResolver

type MultipleResolverErrors struct {
	errors []error
}

func (e *MultipleResolverErrors) Error() string {
	errorStrings := make([]string, len(e.errors))
	for i, err := range e.errors {
		errorStrings[i] = fmt.Sprintf("#%d: %s", i, err.Error())
	}
	return strings.Join(errorStrings, "\n")
}

func NewMultipleResolverErrors(errors []error) *MultipleResolverErrors {
	return &MultipleResolverErrors{
		errors: errors,
	}
}

type GeocodingLocationResolver struct {
	client *maps.Client
}

func NewGeocodingLocationResolver(client *maps.Client) *GeocodingLocationResolver {
	return &GeocodingLocationResolver{
		client: client,
	}
}

func (g *GeocodingLocationResolver) GetPoliticalInfo(loc schema.Location) (schema.Location, error) {
	if loc.Country != "" {
		return loc, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	geos, err := g.client.Geocode(ctx, &maps.GeocodingRequest{
		LatLng: &maps.LatLng{
			Lat: loc.Latitude,
			Lng: loc.Longitude,
		},
		ResultType: []string{"administrative_area_level_2|administrative_area_level_1"},
		Language:   "en",
	})
	if nil != err {
		return loc, err
	}

	if len(geos) == 0 {
		return loc, ErrNoGeoInfoFound
	}

	var level1, level2 string
	for _, a := range geos[0].AddressComponents {
		if len(a.Types) > 0 {
			switch a.Types[0] {
			case "administrative_area_level_1":
				level1 = a.LongName
			case "administrative_area_level_2":
				level2 = a.LongName
			case "country":
				loc.Country = a.LongName
			}
		}
	}

	loc.Address = geos[0].FormattedAddress
	loc.County = level2

	switch loc.Country {
	case US:
		loc.State = level1
	default:
		if loc.County == "" {
			loc.County = level1
		}
	}

	return loc, nil
}

type MongodbLocationResolver struct {
	client   *mongo.Client
	database string
}

func NewMongodbLocationResolver(client *mongo.Client, database string) *MongodbLocationResolver {
	return &MongodbLocationResolver{
		client:   client,
		database: database,
	}
}

func (g *MongodbLocationResolver) GetPoliticalInfo(location schema.Location) (schema.Location, error) {
	ctx := context.Background()

	var address schema.AddressComponent

	if err := g.client.Database(g.database).Collection(schema.BoundaryCollection).FindOne(ctx, bson.M{
		"geometry": bson.M{
			"$geoIntersects": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{location.Longitude, location.Latitude},
				},
			},
		},
	}, options.FindOne().SetProjection(bson.M{
		"country": 1,
		"state":   1,
		"county":  1,
	})).Decode(&address); err != nil {
		if err == mongo.ErrNoDocuments {
			return schema.Location{}, ErrNoGeoInfoFound
		}
		return schema.Location{}, err
	}

	location.AddressComponent = address

	return location, nil
}

type MultipleLocationResolver struct {
	resolvers []LocationResolver
}

func NewMultipleLocationResolver(resolvers ...LocationResolver) *MultipleLocationResolver {
	return &MultipleLocationResolver{
		resolvers: resolvers,
	}
}

func (r *MultipleLocationResolver) GetPoliticalInfo(location schema.Location) (schema.Location, error) {
	var errors []error
	for _, resolver := range r.resolvers {
		result, err := resolver.GetPoliticalInfo(location)
		if err != nil {
			errors = append(errors, err)
		} else {
			return result, nil
		}
	}

	return schema.Location{}, NewMultipleResolverErrors(errors)
}

func SetLocationResolver(resolver LocationResolver) {
	defaultResolver = resolver
}

func PoliticalGeoInfo(loc schema.Location) (schema.Location, error) {
	if defaultResolver == nil {
		return schema.Location{}, ErrResolverNotInitialized
	}

	return defaultResolver.GetPoliticalInfo(loc)
}
