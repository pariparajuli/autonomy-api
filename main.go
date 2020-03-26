package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/getsentry/sentry-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"

	"github.com/bitmark-inc/autonomy-api/api"
	bitmarksdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
)

var (
	server *api.Server
	ormDB  *gorm.DB
)

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

		if server != nil {
			log.Info("Shutdown mobile api server")
			if err := server.Shutdown(ctx); err != nil {
				log.Error("Server Shutdown:", err)
			}
		}

		if ormDB != nil {
			log.Info("Shuting down db store")
			if err := ormDB.Close(); err != nil {
				log.Error(err)
			}
		}

		os.Exit(1)
	}()

	flag.StringVar(&configFile, "c", "./config.yaml", "[optional] path of configuration file")
	flag.Parse()

	loadConfig(configFile)

	initLog()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Sentry
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              viper.GetString("sentry.dsn"),
		AttachStacktrace: true,
		Environment:      viper.GetString("sentry.environment"),
		Dist:             viper.GetString("sentry.dist"),
	}); err != nil {
		log.Error(err)
	}
	log.WithField("prefix", "init").Info("Initilized sentry")

	// Init Bitmark SDK
	bitmarksdk.Init(&bitmarksdk.Config{
		Network:    bitmarksdk.Network(viper.GetString("bitmarksdk.network")),
		APIToken:   viper.GetString("bitmarksdk.token"),
		HTTPClient: httpClient,
	})
	log.WithField("prefix", "init").Info("Initilized bitmark sdk")

	// Load global bitmark account
	a, err := account.FromSeed(viper.GetString("account.seed"))
	if err != nil {
		log.Panic(err)
	}
	globalAccount := a.(*account.AccountV2)
	log.WithField("prefix", "init").Info("Global account: ", globalAccount.AccountNumber())
	log.WithField("prefix", "init").Info("Global enc pub key: ", hex.EncodeToString(globalAccount.EncrKey.PublicKeyBytes()))

	// Load JWT private key
	jwtSecretByte, err := ioutil.ReadFile(viper.GetString("jwt.keyfile"))
	if err != nil {
		log.Panic(err)
	}
	jwtPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEMWithPassword(jwtSecretByte, viper.GetString("jwt.password"))
	if err != nil {
		log.Panic(err)
	}
	log.WithField("prefix", "init").Info("Loaded global jwt key")

	// Init redis
	// var cnf = &machinerycnf.Config{
	// 	Broker:        viper.GetString("redis.conn"),
	// 	DefaultQueue:  "autonomy_background",
	// 	ResultBackend: viper.GetString("redis.conn"),
	// }
	// machineryServer, err := machinery.NewServer(cnf)
	// if err != nil {
	// 	log.Panic(err)
	// }

	ormDB, err = gorm.Open("postgres", viper.GetString("orm.conn"))
	if err != nil {
		log.Panic(err)
	}

	// Init http server
	server = api.NewServer(
		ormDB,
		jwtPrivateKey,
		globalAccount)
	log.WithField("prefix", "init").Info("Initilized http server")

	// Remove initial context
	initialCtx = nil
	cancelInitialization = nil

	log.Fatal(server.Run(":" + viper.GetString("server.port")))
}
