package score

import (
	"net/http"
	"time"

	"github.com/spf13/viper"
	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/bitmark-inc/autonomy-api/background"
	"github.com/bitmark-inc/autonomy-api/external/cadence"
	"github.com/bitmark-inc/autonomy-api/external/onesignal"
	"github.com/bitmark-inc/autonomy-api/store"
)

const TaskListName = "autonomy-score-tasks"

type ScoreUpdateWorker struct {
	domain             string
	mongo              store.MongoStore
	notificationCenter background.NotificationCenter
}

func NewScoreUpdateWorker(domain string, mongo store.MongoStore) *ScoreUpdateWorker {
	o := onesignal.NewClient(&http.Client{
		Timeout: 15 * time.Second,
	})

	return &ScoreUpdateWorker{
		domain:             domain,
		mongo:              mongo,
		notificationCenter: background.NewOnesignalNotificationCenter(viper.GetString("onesignal.appid"), o),
	}
}

func (s *ScoreUpdateWorker) Register() {
	workflow.RegisterWithOptions(s.POIStateUpdateWorkflow, workflow.RegisterOptions{Name: "POIStateUpdateWorkflow"})
	workflow.RegisterWithOptions(s.AccountStateUpdateWorkflow, workflow.RegisterOptions{Name: "AccountStateUpdateWorkflow"})

	activity.RegisterWithOptions(s.CalculatePOIStateActivity, activity.RegisterOptions{Name: "CalculatePOIStateActivity"})
	activity.RegisterWithOptions(s.CalculateAccountStateActivity, activity.RegisterOptions{Name: "CalculateAccountStateActivity"})

	activity.RegisterWithOptions(s.RefreshLocationStateActivity, activity.RegisterOptions{Name: "RefreshLocationStateActivity"})
	activity.RegisterWithOptions(s.NotifyLocationStateActivity, activity.RegisterOptions{Name: "NotifyLocationStateActivity"})

	activity.RegisterWithOptions(s.CheckLocationSpikeActivity, activity.RegisterOptions{Name: "CheckLocationSpikeActivity"})
}

func (s *ScoreUpdateWorker) Start(service workflowserviceclient.Interface, logger *zap.Logger) {
	// TaskListName identifies set of client workflows, activities, and workers.
	// It could be your group or client or application name.
	workerOptions := worker.Options{
		Logger:        logger,
		MetricsScope:  tally.NewTestScope(TaskListName, map[string]string{}),
		DataConverter: cadence.NewMsgPackDataConverter(),
	}

	worker := worker.New(
		service,
		s.domain,
		TaskListName,
		workerOptions)

	if err := worker.Start(); err != nil {
		panic("Failed to start worker")
	}

	logger.Info("Started Worker.", zap.String("worker", TaskListName))

	select {}
}
