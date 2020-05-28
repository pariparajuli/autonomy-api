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

	schema.NewMongoDBIndexer(viper.GetString("mongo.conn"), viper.GetString("mongo.database")).IndexAll()

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

	if err := setupCollectionBehavior(ctx, client); err != nil {
		fmt.Println("failed to set up collection `behavior`: ", err)
		return err
	}

	if err := setupCollectionSymptom(ctx, client); err != nil {
		fmt.Println("failed to set up collection `symptom`: ", err)
		return err
	}

	return nil
}

func setupCollectionBehavior(ctx context.Context, client *mongo.Client) error {
	fmt.Println("initialize behavior collection")
	c := client.Database(viper.GetString("mongo.database")).Collection(schema.BehaviorCollection)

	behaviors := make([]interface{}, 0, len(schema.OfficialBehaviorMatrix))
	for _, b := range schema.OfficialBehaviors {
		behaviors = append(behaviors, b)
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

func setupCollectionSymptom(ctx context.Context, client *mongo.Client) error {
	fmt.Println("initialize symptom collection")
	c := client.Database(viper.GetString("mongo.database")).Collection(schema.SymptomCollection)

	if _, err := c.DeleteMany(ctx, bson.M{"source": schema.OfficialSymptom}); err != nil {
		return err
	}

	officialSymptoms := make([]interface{}, 0, len(schema.COVID19Symptoms))
	for _, s := range schema.COVID19Symptoms {
		officialSymptoms = append(officialSymptoms, s)
	}
	if _, err := c.InsertMany(ctx, officialSymptoms); err != nil {
		return err
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
