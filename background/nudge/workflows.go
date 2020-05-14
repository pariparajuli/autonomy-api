package nudge

import (
	"time"

	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/getsentry/sentry-go"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	AccountSymptomsCheckInterval = time.Hour
)

var activityOptions = workflow.ActivityOptions{
	ScheduleToStartTimeout: time.Minute,
	StartToCloseTimeout:    time.Minute,
	HeartbeatTimeout:       time.Second * 20,
}

// SymptomFollowUpNudgeWorkflow retrive the last symptom report belongs to a given account and
// validate if that report needs to be followed by sending a notification
func (n *NudgeWorker) SymptomFollowUpNudgeWorkflow(ctx workflow.Context, accountNumber string) error {

	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	logger := workflow.GetLogger(ctx)

	selector := workflow.NewSelector(ctx)

	timerCancelCtx, _ := workflow.WithCancel(ctx)
	timerFuture := workflow.NewTimer(timerCancelCtx, AccountSymptomsCheckInterval)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		logger.Info("Start periodically account symptom nedge follow up")
	})

	selector.Select(ctx)

	logger.Info("Check symptoms for following up")
	symptoms := make([]schema.Symptom, 0)
	err := workflow.ExecuteActivity(ctx, n.SymptomsNeedFollowUpActivity, accountNumber).Get(ctx, &symptoms)
	if err != nil {
		logger.Error("Fail to check symptoms for user", zap.Error(err), zap.String("accountNumber", accountNumber))
		sentry.CaptureException(err)
		return workflow.NewContinueAsNewError(ctx, n.SymptomFollowUpNudgeWorkflow, accountNumber)
	}

	if len(symptoms) > 0 {
		err := workflow.ExecuteActivity(ctx, n.NotifySymptomFollowUpActivity, accountNumber, symptoms).Get(ctx, nil)
		if err != nil {
			logger.Error("Fail to notify user", zap.Error(err))
			sentry.CaptureException(err)
			return workflow.NewContinueAsNewError(ctx, n.SymptomFollowUpNudgeWorkflow, accountNumber)
		}
	}

	return workflow.NewContinueAsNewError(ctx, n.SymptomFollowUpNudgeWorkflow, accountNumber)
}

func (n *NudgeWorker) FindNotificationReceiver(accountNumber, poiID string) ([]string, error) {
	accountNumbers := make([]string, 0)

	if poiID != "" {
		profiles, err := n.mongo.GetProfilesByPOI(poiID)
		if err != nil {
			return nil, err
		}

		for _, p := range profiles {
			accountNumbers = append(accountNumbers, p.AccountNumber)
		}

	} else if accountNumber != "" {
		accountNumbers = append(accountNumbers, accountNumber)
	}
	return accountNumbers, nil
}

func (n *NudgeWorker) NotifySymptomSpikeWorkflow(ctx workflow.Context, accountNumber string, poiID string, symptoms []schema.Symptom) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	logger := workflow.GetLogger(ctx)

	receivers := make([]string, 0)
	if err := workflow.ExecuteActivity(ctx, n.FindNotificationReceiver, accountNumber, poiID).Get(ctx, &receivers); err != nil {
		logger.Error("Fail to get notification receivers", zap.Error(err))
		return err
	}

	logger.Info("notify symptom spike", zap.Any("receivers", receivers), zap.Any("symptoms", symptoms))

	for _, accountNumber := range receivers {
		err := workflow.ExecuteActivity(ctx, n.NotifySymptomSpikeActivity, accountNumber, symptoms).Get(ctx, nil)
		if err != nil {
			logger.Error("Fail to notify user", zap.Error(err))
			return err
		}
	}

	return nil
}

func (n *NudgeWorker) NotifyBehaviorOnRiskAreaWorkflow(ctx workflow.Context, accountNumber string) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	logger := workflow.GetLogger(ctx)

	err := workflow.ExecuteActivity(ctx, n.NotifyBehaviorNudgeActivity, accountNumber).Get(ctx, nil)
	if err != nil {
		logger.Error("Fail to notify user behavior nudge on risk area", zap.Error(err))
		return err
	}

	return nil
}
