package score

import (
	"net/http"
	"time"

	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/bitmark-inc/autonomy-api/background"
	"github.com/bitmark-inc/autonomy-api/external/onesignal"
	"github.com/bitmark-inc/autonomy-api/store"
)

const TaskListName = "autonomy-score-tasks"

type ScoreUpdateWorker struct {
	background.Background
	domain string
	mongo  store.MongoStore
}

func NewScoreUpdateWorker(domain string, mongo store.MongoStore) *ScoreUpdateWorker {
	o := onesignal.NewClient(&http.Client{
		Timeout: 15 * time.Second,
	})

	b := background.Background{o}
	return &ScoreUpdateWorker{
		Background: b,
		domain:     domain,
		mongo:      mongo,
	}
}

func (s *ScoreUpdateWorker) Register() {
	workflow.RegisterWithOptions(s.POIStateUpdateWorkflow, workflow.RegisterOptions{Name: "POIStateUpdateWorkflow"})
	workflow.RegisterWithOptions(s.AccountStateUpdateWorkflow, workflow.RegisterOptions{Name: "AccountStateUpdateWorkflow"})

	activity.RegisterWithOptions(s.CalculatePOIStateActivity, activity.RegisterOptions{Name: "CalculatePOIStateActivity"})
	activity.RegisterWithOptions(s.CalculateAccountStateActivity, activity.RegisterOptions{Name: "CalculateAccountStateActivity"})

	activity.RegisterWithOptions(s.RefreshLocationStateActivity, activity.RegisterOptions{Name: "UpdateLocationStateActivity"})
	activity.RegisterWithOptions(s.NotifyLocationStateActivity, activity.RegisterOptions{Name: "NotifyLocationStateActivity"})

	activity.RegisterWithOptions(s.CheckLocationSpikeActivity, activity.RegisterOptions{Name: "CheckLocationSpikeActivity"})
}

func (s *ScoreUpdateWorker) Start(service workflowserviceclient.Interface, logger *zap.Logger) {
	// TaskListName identifies set of client workflows, activities, and workers.
	// It could be your group or client or application name.
	workerOptions := worker.Options{
		Logger:       logger,
		MetricsScope: tally.NewTestScope(TaskListName, map[string]string{}),
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
