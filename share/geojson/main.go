package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GeoFeature struct {
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties"`
	Geometry   schema.Geometry   `json:"geometry"`
}

type GeoJSONTW struct {
	Name     string       `json:"name"`
	Features []GeoFeature `json:"features"`
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("autonomy")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func main() {
	ctx := context.Background()
	opts := options.Client().ApplyURI(viper.GetString("mongo.conn"))
	client, err := mongo.NewClient(opts)
	if err != nil {
		panic(err)
	}
	if err := client.Connect(ctx); err != nil {
		panic(err)
	}

	var result GeoJSONTW

	file, err := os.Open("tw-boundary.json")
	if err != nil {
		panic(err)
	}

	if err := json.NewDecoder(file).Decode(&result); err != nil {
		panic(err)
	}

	var boundaries []interface{}
	for _, b := range result.Features {
		boundaries = append(boundaries, schema.Boundary{
			Country:  "Taiwan",
			State:    "",
			County:   b.Properties["COUNTYENG"],
			Geometry: b.Geometry,
		})
	}

	if _, err := client.Database(viper.GetString("mongo.database")).Collection(schema.BoundaryCollection).InsertMany(context.Background(), boundaries); err != nil {
		panic(err)
	}
}
