package geojson

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitmark-inc/autonomy-api/schema"
)

type GeoFeature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   schema.Geometry        `json:"geometry"`
}

type GeoJSON struct {
	Name     string       `json:"name"`
	Features []GeoFeature `json:"features"`
}

type GeoFeatureUS struct {
	Type       string        `json:"type"`
	Properties PropertiesUS  `json:"properties"`
	Geometry   schema.Geometry      `json:"geometry"`
}

type Geometry struct {
	Type        string      `bson:"type"`
	Coordinates interface{} `bson:"coordinates"`
}

type Boundary struct {
	Country  string   `bson:"country"`
	State    string   `bson:"state"`
	County   string   `bson:"county"`
	Geometry Geometry `bson:"geometry"`
}

type PropertiesUS struct {
	Intptlat   string    `json:"intptlat"`
	GeoPoint2d []float64 `json:"geo_point_2d"`
	Stusab     string    `json:"stusab"`
	Namelsad   string    `json:"namelsad"`
	Awater     int       `json:"awater"`
}

type GeoJSONUS struct {
	Name     string          `json:"name"`
	Features []GeoFeatureUS  `json:"features"`
}

func ImportTaiwanBoundary(client *mongo.Client, dbName, geoJSONFile string) error {
	var result GeoJSON

	file, err := os.Open(geoJSONFile)
	if err != nil {
		return err
	}

	if err := json.NewDecoder(file).Decode(&result); err != nil {
		return err
	}

	var boundaries []interface{}
	for _, b := range result.Features {
		county, ok := b.Properties["COUNTYENG"].(string)
		if !ok {
			return fmt.Errorf("invalid county value, %+v", b.Properties["COUNTYENG"])
		}
		boundaries = append(boundaries, schema.Boundary{
			Country:  "Taiwan",
			Island:   "Taiwan",
			State:    "",
			County:   county,
			Geometry: b.Geometry,
		})
	}

	if _, err := client.Database(dbName).Collection(schema.BoundaryCollection).InsertMany(context.Background(), boundaries); err != nil {
		return err
	}

	return nil
}

func ImportUSBoundary(client *mongo.Client, dbName, geoJSONFile string) error {
	var result GeoJSONUS

	file, err := os.Open(geoJSONFile)
	if err != nil {
		return err
	}

	if err := json.NewDecoder(file).Decode(&result); err != nil {
		return err
	}

	var boundaries []interface{}
	for _, b := range result.Features {
		county:= b.Properties.Namelsad
		boundaries = append(boundaries, schema.Boundary{
			Country:  "United States",
			Island:   "",
			State:    b.Properties.Stusab,
			County:   county,
			Geometry: b.Geometry,
		})
	}

	if _, err := client.Database(dbName).Collection(schema.BoundaryCollection).InsertMany(context.Background(), boundaries); err != nil {
		return err
	}

	return nil

}

func ImportWorldCountryBoundary(client *mongo.Client, dbName, geoJSONFile string) error {
	var result GeoJSON

	file, err := os.Open(geoJSONFile)
	if err != nil {
		return err
	}

	if err := json.NewDecoder(file).Decode(&result); err != nil {
		return err
	}

	var boundaries []interface{}
	for _, b := range result.Features {
		country, ok := b.Properties["COUNTRYAFF"].(string)
		if !ok {
			return fmt.Errorf("invalid country value, %+v", b.Properties["COUNTRYAFF"])
		}

		island, ok := b.Properties["COUNTRY"].(string)
		if !ok {
			return fmt.Errorf("invalid island value, %+v", b.Properties["COUNTRY"])
		}

		boundaries = append(boundaries, schema.Boundary{
			Country:  country,
			Island:   island,
			State:    "",
			County:   "",
			Geometry: b.Geometry,
		})
	}

	if _, err := client.Database(dbName).Collection(schema.BoundaryCollection).InsertMany(context.Background(), boundaries); err != nil {
		return err
	}

	return nil
}
