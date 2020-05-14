package nudge

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

const TaskListName = "autonomy-nudge-tasks"

type NudgeWorker struct {
	background.Background
	domain string
	mongo  store.MongoStore
}

func NewNudgeWorker(domain string, mongo store.MongoStore) *NudgeWorker {
	o := onesignal.NewClient(&http.Client{
		Timeout: 15 * time.Second,
	})

	b := background.Background{o}
	return &NudgeWorker{
		Background: b,
		domain:     domain,
		mongo:      mongo,
	}
}

func (n *NudgeWorker) Register() {
	workflow.RegisterWithOptions(n.SymptomFollowUpNudgeWorkflow, workflow.RegisterOptions{Name: "SymptomFollowUpNudgeWorkflow"})
	workflow.RegisterWithOptions(n.NotifySymptomSpikeWorkflow, workflow.RegisterOptions{Name: "NotifySymptomSpikeWorkflow"})
	workflow.RegisterWithOptions(n.NotifyBehaviorOnRiskAreaWorkflow, workflow.RegisterOptions{Name: "NotifyBehaviorOnRiskAreaWorkflow"})
	workflow.RegisterWithOptions(n.HighRiskAccountFollowUpWorkflow, workflow.RegisterOptions{Name: "HighRiskAccountFollowUpWorkflow"})

	activity.RegisterWithOptions(n.SymptomsNeedFollowUpActivity, activity.RegisterOptions{Name: "SymptomsNeedFollowUpActivity"})
	activity.RegisterWithOptions(n.NotifySymptomFollowUpActivity, activity.RegisterOptions{Name: "NotifySymptomFollowUpActivity"})
	activity.RegisterWithOptions(n.NotifySymptomSpikeActivity, activity.RegisterOptions{Name: "NotifySymptomSpikeActivity"})
	activity.RegisterWithOptions(n.NotifyBehaviorNudgeActivity, activity.RegisterOptions{Name: "NotifyBehaviorNudgeActivity"})
	activity.RegisterWithOptions(n.FindNotificationReceiver, activity.RegisterOptions{Name: "FindNotificationReceiver"})
	activity.RegisterWithOptions(n.HighRiskAccountFollowUpActivity, activity.RegisterOptions{Name: "HighRiskAccountFollowUpActivity"})
	activity.RegisterWithOptions(n.NotifyBehaviorFollowUpActivity, activity.RegisterOptions{Name: "NotifyBehaviorFollowUpActivity"})
}

func (n *NudgeWorker) Start(service workflowserviceclient.Interface, logger *zap.Logger) {
	// TaskListName identifies set of client workflows, activities, and workers.
	// It could be your group or client or application name.
	workerOptions := worker.Options{
		Logger:       logger,
		MetricsScope: tally.NewTestScope(TaskListName, map[string]string{}),
	}

	worker := worker.New(
		service,
		n.domain,
		TaskListName,
		workerOptions)

	if err := worker.Start(); err != nil {
		panic("Failed to start worker")
	}

	logger.Info("Started Worker.", zap.String("worker", TaskListName))

	select {}
}
