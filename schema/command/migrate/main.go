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
	"github.com/bitmark-inc/autonomy-api/store"
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
	ctx := context.Background()
	opts := options.Client().ApplyURI(viper.GetString("mongo.conn"))
	opts.SetMaxPoolSize(1)
	client, _ := mongo.NewClient(opts)
	_ = client.Connect(ctx)
	c := client.Database(viper.GetString("mongo.database")).Collection(schema.ProfileCollection)

	// here is reference from api/store/profile
	// if bson key of location is changed, here should also be changed
	geo := mongo.IndexModel{
		Keys: bson.M{
			"location": "2dsphere",
		},
		Options: nil,
	}

	_, err := c.Indexes().CreateOne(ctx, geo)
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

	_, err = c.Indexes().CreateOne(ctx, id)
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

	_, err = c.Indexes().CreateOne(ctx, accountNumber)
	if nil != err {
		fmt.Println("mongodb create account_number index with error: ", err)
		return err
	}

	if err := setupCollectionPOI(client); err != nil {
		fmt.Println("failed to set up collection `poi`: ", err)
		return err
	}
	if err := setupCollectionBehavior(ctx, client); err != nil {
		fmt.Println("failed to set up collection `behavior`: ", err)
		return err
	}
	if err := setupCollectionBehaviorReport(client); err != nil {
		fmt.Println("failed to set up collection `behavior`: ", err)
		return err
	}
	if err := BehaviorListNullToEmptyArray(client); err != nil {
		fmt.Println("failed to convert null to empty array in  `behavior` list : ", err)
		return err
	}

	if err := setupCollectionSymptom(ctx, client); err != nil {
		fmt.Println("failed to set up collection `symptom`: ", err)
		return err
	}

	if err := setupCollectionSymptomReport(client); err != nil {
		fmt.Println("failed to set up collection `symptom reports`: ", err)
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

func setupCollectionBehavior(ctx context.Context, client *mongo.Client) error {
	fmt.Println("initialize behavior collection")
	c := client.Database(viper.GetString("mongo.database")).Collection(schema.BehaviorCollection)

	behaviors := make([]interface{}, 0, len(schema.OfficialBehaviorMatrix))
	for _, b := range schema.OfficialBehaviors {
		behaviors = append(behaviors, b)
	}

	if _, err := c.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"source": 1,
		},
	}); err != nil {
		return err
	}

	_, err := c.InsertMany(ctx, behaviors)
	if err != nil {
		if errs, hasErr := err.(mongo.BulkWriteException); hasErr {
			if 1 == len(errs.WriteErrors) && store.DuplicateKeyCode == errs.WriteErrors[0].Code {
				return nil
			}
		}
	}

	return err
}

func setupCollectionBehaviorReport(client *mongo.Client) error {
	c := client.Database(viper.GetString("mongo.database")).Collection(schema.BehaviorReportCollection)
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

func setupCollectionSymptom(ctx context.Context, client *mongo.Client) error {
	fmt.Println("initialize symptom collection")
	c := client.Database(viper.GetString("mongo.database")).Collection(schema.SymptomCollection)

	if _, err := c.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"source": 1,
		},
	}); err != nil {
		return err
	}

	officialSymptoms := make([]interface{}, 0, len(schema.COVID19Symptoms))
	for _, s := range schema.COVID19Symptoms {
		officialSymptoms = append(officialSymptoms, s)
	}
	_, err := c.InsertMany(ctx, officialSymptoms)
	if err != nil {
		if errs, hasErr := err.(mongo.BulkWriteException); hasErr {
			if !(1 == len(errs.WriteErrors) && store.DuplicateKeyCode == errs.WriteErrors[0].Code) {
				return err
			}
		}
	}

	generalSymptoms := make([]interface{}, 0, len(schema.GeneralSymptoms))
	for _, s := range schema.GeneralSymptoms {
		generalSymptoms = append(generalSymptoms, s)
	}
	if _, err := c.InsertMany(ctx, generalSymptoms); err != nil {
		if errs, hasErr := err.(mongo.BulkWriteException); hasErr {
			if !(1 == len(errs.WriteErrors) && store.DuplicateKeyCode == errs.WriteErrors[0].Code) {
				return err
			}
		}
	}

	return nil
}

func setupCollectionSymptomReport(client *mongo.Client) error {
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

func BehaviorListNullToEmptyArray(client *mongo.Client) error {
	c := client.Database(viper.GetString("mongo.database")).Collection(schema.BehaviorReportCollection)
	filter := bson.D{{"official_behaviors", bson.M{"$type": 10}}}
	update := bson.D{{"$set", bson.D{{"official_behaviors", []schema.Behavior{}}}}}
	result, err := c.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return err
	}
	fmt.Println("Migration: replace Official Behavior List with Empty Array result:", result.MatchedCount)
	filter = bson.D{{"customized_behaviors", bson.M{"$type": 10}}}
	update = bson.D{{"$set", bson.D{{"customized_behaviors", []schema.Behavior{}}}}}
	result, err = c.UpdateMany(context.Background(), filter, update)
	if err != nil {
		return err
	}
	fmt.Println("Migration: replace Customized Behavior List with Empty Array result:", result.MatchedCount)
	return nil
}
