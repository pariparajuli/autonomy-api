package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bitmark-inc/autonomy-api/schema"
)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("autonomy")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func main() {
	db, err := gorm.Open("postgres", viper.GetString("orm.conn"))
	if err != nil {
		panic(err)
	}

	if err := db.Exec(`CREATE SCHEMA IF NOT EXISTS autonomy`).Error; err != nil {
		panic(err)
	}

	if err := db.Exec("SET search_path TO autonomy").Error; err != nil {
		panic(err)
	}

	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(
		&schema.Account{},
		&schema.AccountProfile{},
		&schema.HelpRequest{},
	).Error; err != nil {
		panic(err)
	}

	if err := db.Model(schema.HelpRequest{}).Where(fmt.Sprintf("state = '%s'", "PENDING")).
		AddUniqueIndex("help_request_unique_if_not_done", "requester").Error; err != nil {
		panic(err)
	}

	err = migrateMongo()
	if nil != err {
		panic(err)
	}
}

func migrateMongo() error {
	opts := options.Client().ApplyURI(viper.GetString("mongo.conn"))
	opts.SetMaxPoolSize(1)
	client, _ := mongo.NewClient(opts)
	_ = client.Connect(context.Background())
	c := client.Database(viper.GetString("mongo.database")).Collection(schema.ProfileCollectionName)

	// here is reference from api/store/profile
	// if bson key of location is changed, here should also be changed
	geo := mongo.IndexModel{
		Keys: bson.M{
			"location": "2dsphere",
		},
		Options: nil,
	}

	_, err := c.Indexes().CreateOne(context.Background(), geo)
	if nil != err {
		fmt.Println("mongodb create geo index with error: ", err)
		return err
	}

	id := mongo.IndexModel{
		Keys: bson.M{
			"id": 1,
		},
		Options: options.Index().SetUnique(true),
	}

	_, err = c.Indexes().CreateOne(context.Background(), id)
	if nil != err {
		fmt.Println("mongodb create id index with error: ", err)
		return err
	}

	accountNumber := mongo.IndexModel{
		Keys: bson.M{
			"account_number": 1,
		},
		Options: options.Index().SetUnique(true),
	}

	_, err = c.Indexes().CreateOne(context.Background(), accountNumber)
	if nil != err {
		fmt.Println("mongodb create account_number index with error: ", err)
		return err
	}

	if err := setupCollectionPOI(client); err != nil {
		fmt.Println("failed to set up collection `poi`: ", err)
		return err
	}

	if err := setupCollectionBehavior(client); err != nil {
		fmt.Println("failed to set up collection `poi`: ", err)
		return err
	}

	return nil
}

func setupCollectionPOI(client *mongo.Client) error {
	// add indices for collection poi
	c := client.Database(viper.GetString("mongo.database")).Collection(schema.POICollection)
	locationIndex := mongo.IndexModel{
		Keys: bson.M{
			"location": "2dsphere",
		},
		Options: nil,
	}

	_, err := c.Indexes().CreateOne(context.Background(), locationIndex)
	return err
}

func setupCollectionBehavior(client *mongo.Client) error {
	c := client.Database(viper.GetString("mongo.database")).Collection(schema.GoodBehaviorCollection)
	idAndTs := mongo.IndexModel{
		Keys: bson.M{
			"profile_id": 1,
			"ts":         1,
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := c.Indexes().CreateOne(context.Background(), idAndTs)
	if nil != err {
		fmt.Println("mongodb create id & ts combined index with error: ", err)
		return err
	}

	locationIndex := mongo.IndexModel{
		Keys: bson.M{
			"location": "2dsphere",
		},
		Options: nil,
	}
	_, err = c.Indexes().CreateOne(context.Background(), locationIndex)
	if nil != err {
		fmt.Println("mongodb create locationIndex with error: ", err)
		return err
	}
	return nil
}

func setupCollectionSymptom(client *mongo.Client) error {
	c := client.Database(viper.GetString("mongo.database")).Collection(schema.SymptomReportCollection)
	idAndTs := mongo.IndexModel{
		Keys: bson.M{
			"profile_id": 1,
			"ts":         1,
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := c.Indexes().CreateOne(context.Background(), idAndTs)
	if nil != err {
		fmt.Println("mongodb create id & ts combined index with error: ", err)
		return err
	}

	locationIndex := mongo.IndexModel{
		Keys: bson.M{
			"location": "2dsphere",
		},
		Options: nil,
	}
	_, err = c.Indexes().CreateOne(context.Background(), locationIndex)
	if nil != err {
		fmt.Println("mongodb create locationIndex with error: ", err)
		return err
	}
	return nil
}
