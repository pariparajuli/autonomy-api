package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bitmark-inc/autonomy-api/background"
)

var (
	ormDB       *gorm.DB
	mongoClient *mongo.Client
	manager     *background.BackgroundManager
)

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
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

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Info("Server is preparing to shutdown")

		if initialCtx != nil && cancelInitialization != nil {
			log.Info("Cancelling initialization")
			cancelInitialization()
			<-initialCtx.Done()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if ormDB != nil {
			log.Info("Shutting down orm store")
			if err := ormDB.Close(); err != nil {
				log.Error(err)
			}
		}

		if mongoClient != nil {
			log.Info("Shutting down mongo store")
			mongoClient.Disconnect(ctx)
		}
	}()

	flag.StringVar(&configFile, "c", "./config.yaml", "[optional] path of configuration file")
	flag.Parse()

	loadConfig(configFile)

	initLog()

	var err error

	ormDB, err = gorm.Open("postgres", viper.GetString("orm.conn"))
	if err != nil {
		log.Panic(err)
	}

	// initialise mongodb connections
	opts := options.Client().ApplyURI(viper.GetString("mongo.conn"))
	opts.SetMaxPoolSize(viper.GetUint64("mongo.pool"))
	mongoClient, err = mongo.NewClient(opts)
	if nil != err {
		log.Panicf("create mongo client with error: %s", err)
	}

	err = mongoClient.Connect(initialCtx)
	if nil != err {
		log.Panicf("connect mongo database with error: %s", err)
	}

	var conf = &config.Config{
		Broker:        viper.GetString("redis.conn"),
		DefaultQueue:  "autonomy_background",
		ResultBackend: viper.GetString("redis.conn"),
	}
	taskServer, err := machinery.NewServer(conf)
	if err != nil {
		log.Panic(err)
	}

	manager = background.New(ormDB, mongoClient, taskServer)
	panicIfError(manager.RegisterTask("broadcast_help", manager.BroadcastNewHelp))
	panicIfError(manager.RegisterTask("notify_help_accepted", manager.NotifyHelpAccepted))
	panicIfError(manager.RegisterTask("expire_help_requests", manager.ExpireHelpRequests))

	if err := manager.Run(); err != nil {
		log.Panic(err)
	}
}
