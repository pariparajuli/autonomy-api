package background

import (
	"errors"
	"net/http"
	"time"

	"github.com/RichardKnop/machinery/v1"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitmark-inc/autonomy-api/external/onesignal"
	"github.com/bitmark-inc/autonomy-api/store"
)

// BackgroundManager is a struct for autonomy background manager
type BackgroundManager struct {
	store store.AutonomyCore

	onesignal *onesignal.OneSignalClient

	taskServer *machinery.Server

	worker *machinery.Worker
}

func New(ormDB *gorm.DB, mongoClient *mongo.Client, taskServer *machinery.Server) *BackgroundManager {
	autonomyCore := store.NewAutonomyStore(ormDB, store.NewMongoStore(
		mongoClient,
		viper.GetString("mongo.database"),
	))

	o := onesignal.NewClient(&http.Client{
		Timeout: 15 * time.Second,
	})

	return &BackgroundManager{
		store:      autonomyCore,
		onesignal:  o,
		taskServer: taskServer,
	}
}

func (m *BackgroundManager) RegisterTask(name string, taskFunc interface{}) error {
	return m.taskServer.RegisterTask(name, taskFunc)
}

// Run spawn workers to execute background jobs
func (m *BackgroundManager) Run() error {
	if m.worker != nil {
		return errors.New("background worker has started")
	}
	m.worker = m.taskServer.NewWorker("autonomy-worker", 5)
	return m.worker.Launch()
}
