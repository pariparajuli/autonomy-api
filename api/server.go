package api

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"encoding/hex"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/RichardKnop/machinery/v1"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/bitmark-inc/bitmark-sdk-go/account"

	"github.com/bitmark-inc/autonomy-api/external/geoinfo"
	"github.com/bitmark-inc/autonomy-api/external/onesignal"
	"github.com/bitmark-inc/autonomy-api/logmodule"
	"github.com/bitmark-inc/autonomy-api/store"
)

var log *logrus.Entry

func init() {
	log = logrus.WithField("prefix", "gin")
}

// Server to run a http server instance
type Server struct {
	// Server instance
	server *http.Server

	// Stores
	store      store.AutonomyCore
	mongoStore store.MongoStore

	// JWT private key
	jwtPrivateKey *rsa.PrivateKey

	// External services
	oneSignalClient *onesignal.OneSignalClient
	geoClient       geoinfo.GeoInfo

	// account
	bitmarkAccount *account.AccountV2

	// http client for calling external services
	httpClient *http.Client

	// job pool enqueuer
	background *machinery.Server
}

// NewServer new instance of server
func NewServer(
	ormDB *gorm.DB,
	mongoClient *mongo.Client,
	machineryServer *machinery.Server,
	jwtKey *rsa.PrivateKey,
	bitmarkAccount *account.AccountV2,
	geoClient geoinfo.GeoInfo) *Server {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	httpClient := &http.Client{
		Timeout:   5 * time.Minute,
		Transport: tr,
	}

	mongoStore := store.NewMongoStore(
		mongoClient,
		viper.GetString("mongo.database"),
	)

	return &Server{
		store:           store.NewAutonomyStore(ormDB, mongoStore),
		mongoStore:      mongoStore,
		background:      machineryServer,
		jwtPrivateKey:   jwtKey,
		httpClient:      httpClient,
		bitmarkAccount:  bitmarkAccount,
		oneSignalClient: onesignal.NewClient(httpClient),
		geoClient:       geoClient,
	}
}

// Run to run the server
func (s *Server) Run(addr string) error {
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.setupRouter(),
	}

	return s.server.ListenAndServe()
}

func (s *Server) setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(sentrygin.New(sentrygin.Options{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         10 * time.Second,
	}))

	webhookRoute := r.Group("/webhook")
	webhookRoute.Use(logmodule.Ginrus("Webhook"))
	{
	}

	apiRoute := r.Group("/api")
	apiRoute.Use(logmodule.Ginrus("API"))
	apiRoute.GET("/information", s.information)

	// api route other than `/information` will apply the following middleware
	apiRoute.Use(s.clientVersionGateway())

	apiRoute.POST("/auth", s.requestJWT)

	// api route other than `/auth` will apply the following middleware
	apiRoute.Use(s.authMiddleware())
	apiRoute.Use(s.updateGeoPositionMiddleware)

	accountRoute := apiRoute.Group("/accounts")
	{
		accountRoute.POST("", s.accountRegister)
	}

	accountRoute.Use(s.recognizeAccountMiddleware())
	{
		accountRoute.GET("/me", s.accountDetail)

		accountRoute.PATCH("/me", s.accountUpdateMetadata)
		accountRoute.DELETE("/me", s.accountDelete)

		// accountRoute.POST("/me/export", s.accountPrepareExport)
		// accountRoute.GET("/me/export", s.accountExportStatus)
		// accountRoute.GET("/me/export/download", s.accountDownloadExport)
	}

	helpRoute := apiRoute.Group("/helps")
	helpRoute.Use(s.recognizeAccountMiddleware())
	{
		helpRoute.POST("", s.askForHelp)
		helpRoute.GET("", s.queryHelps)
		helpRoute.GET("/:helpID", s.queryHelps)
		helpRoute.PATCH("/:helpID", s.answerHelp)
	}

	secretRoute := r.Group("/secret")
	secretRoute.Use(logmodule.Ginrus("Secret"))
	secretRoute.Use(s.apikeyAuthentication(viper.GetString("server.apikey.admin")))
	{
		// secretRoute.POST("/delete-accounts", s.adminAccountDelete)
		secretRoute.POST("/expire-reqeusts", s.adminExpireRequests)

	}

	symptomRoute := apiRoute.Group("/symptoms")
	{
		symptomRoute.GET("", s.getSymptoms)

	}

	metricRoute := r.Group("/metrics")
	metricRoute.Use(logmodule.Ginrus("Metric"))
	metricRoute.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowAllOrigins:  true,
		MaxAge:           12 * time.Hour,
	}))
	metricRoute.Use(s.apikeyAuthentication(viper.GetString("server.apikey.metric")))
	{
		// What kind of metrics do we need?
		// metricRoute.GET("/total-users", s.metricAccountCreation)
	}

	// health score
	scoreRoute := apiRoute.Group("/score")
	scoreRoute.Use(s.recognizeAccountMiddleware())
	{
		scoreRoute.GET("", s.score)
	}

	// points of interest
	poiRoute := apiRoute.Group("/points-of-interest")
	poiRoute.Use(s.recognizeAccountMiddleware())
	{
		poiRoute.POST("", s.addPOI)
		poiRoute.GET("", s.getPOI)
		poiRoute.PATCH("/:poiID", s.updatePOIAlias)
		poiRoute.DELETE("/:poiID", s.deletePOI)
	}

	r.GET("/healthz", s.healthz)

	reportRoute := apiRoute.Group("/report")
	reportRoute.Use(s.recognizeAccountMiddleware())
	{
		reportRoute.POST("", s.report)
	}

	behaviorRoute := apiRoute.Group("/behaviors")
	behaviorRoute.Use(s.recognizeAccountMiddleware())
	{
		behaviorRoute.GET("", s.goodBehaviors)
		behaviorRoute.POST("", s.reportBehaviors)
	}

	return r
}

// Shutdown to shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.mongoStore.Close()
	return s.server.Shutdown(ctx)
}

// shouldInterupt sends error message and determine if it should interupt the current flow
func shouldInterupt(err error, c *gin.Context) bool {
	if err == nil {
		return false
	}

	log.Error(err)
	abortWithEncoding(c, http.StatusInternalServerError, errorInternalServer)
	return true
}

func (s *Server) healthz(c *gin.Context) {
	// Ping db
	err := s.store.Ping()
	if shouldInterupt(err, c) {
		return
	}

	err = s.mongoStore.Ping()
	if shouldInterupt(err, c) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"version": viper.GetString("server.version"),
	})
}

func (s *Server) information(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"information": map[string]interface{}{
			"server": map[string]interface{}{
				"version":                viper.GetString("server.version"),
				"enc_pub_key":            hex.EncodeToString(s.bitmarkAccount.EncrKey.PublicKeyBytes()),
				"bitmark_account_number": s.bitmarkAccount.AccountNumber(),
			},
			"android":        viper.GetStringMap("clients.android"),
			"ios":            viper.GetStringMap("clients.ios"),
			"system_version": "Autonomy 0.1",
			"docs":           viper.GetStringMap("docs"),
		},
	})
}

func responseWithEncoding(c *gin.Context, code int, obj gin.H) {
	acceptEncoding := c.GetHeader("Accept-Encoding")
	switch acceptEncoding {
	default:
		c.JSON(code, obj)
	}
}

func abortWithEncoding(c *gin.Context, code int, obj ErrorResponse, errors ...error) {
	for _, err := range errors {
		c.Error(err)
	}
	responseWithEncoding(c, code, gin.H{
		"error": obj,
	})
	c.Abort()
}
