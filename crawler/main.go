package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bitmark-inc/autonomy-api/external/cdc"
	"github.com/bitmark-inc/autonomy-api/store"
)

const (
	logPrefix      = "cron"
	twURL          = "https://od.cdc.gov.tw/eic/Weekly_Age_County_Gender_19CoV.json"
	cdsURL         = "https://coronadatascraper.com/data.json"
	defaultTimeout = 15 * time.Second
)

type Cron interface {
	Run()
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("autonomy")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func initLog() {
	logLevel, err := log.ParseLevel(viper.GetString("log.level"))
	if err != nil {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(logLevel)
	}

	log.SetOutput(os.Stdout)

	log.SetFormatter(&prefixed.TextFormatter{
		ForceFormatting: true,
		FullTimestamp:   true,
	})
}

func loadConfig(file string) {
	// Config from file
	viper.SetConfigType("yaml")
	if file != "" {
		viper.SetConfigFile(file)
	}

	viper.AddConfigPath("/.config/")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("No config file. Read config from env.")
		viper.AllowEmptyEnv(false)
	}

	// Config from env if possible
	viper.AutomaticEnv()
	viper.SetEnvPrefix("autonomy")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func main() {
	var configFile string

	initialCtx, cancelInitialization := context.WithCancel(context.Background())

	flag.StringVar(&configFile, "c", "./config.yaml", "[optional] path of configuration file")
	flag.Parse()

	loadConfig(configFile)

	initLog()

	var err error

	// initialise mongodb connections
	opts := options.Client().ApplyURI(viper.GetString("mongo.conn"))
	opts.SetMaxPoolSize(viper.GetUint64("mongo.pool"))
	mongoClient, err := mongo.NewClient(opts)
	if nil != err {
		log.Panicf("create mongo client with error: %s", err)
	}

	err = mongoClient.Connect(initialCtx)
	if nil != err {
		log.Panicf("connect mongo database with error: %s", err)
	}

	mStore := store.NewMongoStore(
		mongoClient,
		viper.GetString("mongo.database"),
	)

	crawlerTWCDC := newTWCrawler("tw", mStore, cdc.NewTw(twURL))
	crawlerTWCDC.Run()

	if cancelInitialization != nil {
		cancelInitialization()
	}

	crawlerTW := newCDSCrawler("Taiwan", mStore, cdc.NewCDS("Taiwan", "country", cdc.CDSDailyHTTP, nil, cdsURL))

	crawlerTW.Run()

	if cancelInitialization != nil {
		cancelInitialization()
	}

	crawlerUS := newCDSCrawler("United States", mStore, cdc.NewCDS("United States", "county", cdc.CDSDailyHTTP, nil, cdsURL))

	crawlerUS.Run()

	if cancelInitialization != nil {
		cancelInitialization()
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if mongoClient != nil {
		log.Info("Shutting down mongo store")
		_ = mongoClient.Disconnect(ctx)
	}
}
