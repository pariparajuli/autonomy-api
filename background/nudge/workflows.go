package nudge

import (
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/bitmark-inc/autonomy-api/background"
	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	AccountSymptomsCheckInterval = time.Hour
	AccountHighRiskCheckInterval = 30 * time.Minute
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

// NotifySymptomSpikeWorkflow is a workflow that deliver symptoms spike notifications to related
// accounts base on given account number or poi ID
func (n *NudgeWorker) NotifySymptomSpikeWorkflow(ctx workflow.Context, accountNumber string, poiID string, symptoms []schema.Symptom) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	logger := workflow.GetLogger(ctx)

	receivers := make([]string, 0)
	if err := workflow.ExecuteActivity(ctx, n.GetNotificationReceiverActivity, accountNumber, poiID).Get(ctx, &receivers); err != nil {
		logger.Error("Fail to get notification receivers", zap.Error(err))
		return err
	}

	logger.Info("notify symptom spike", zap.Any("receivers", receivers), zap.Any("symptoms", symptoms))

	for _, accountNumber := range receivers {
		err := workflow.ExecuteActivity(ctx, n.NotifySymptomSpikeActivity, accountNumber, symptoms).Get(ctx, nil)
		if err != nil {
			logger.Error("Fail to notify user", zap.Error(err))
			sentry.CaptureException(err)
			return err
		}
	}

	return nil
}

// NotifyBehaviorFollowUpOnEnteringSymptomSpikeAreaWorkflow is a workflow to send behavior follow up to
// accounts that enters a symptom spike area. [NB_3-1]
func (n *NudgeWorker) NotifyBehaviorFollowUpOnEnteringSymptomSpikeAreaWorkflow(ctx workflow.Context, accountNumber string) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	logger := workflow.GetLogger(ctx)

	err := workflow.ExecuteActivity(ctx, n.NotifyBehaviorFollowUpWhenSelfIsInHighRiskActivity, accountNumber, schema.NudgeBehaviorOnSymptomSpikeArea).Get(ctx, nil)
	if err != nil {
		logger.Error("Fail to notify user behavior nudge on risk area (symptom score spike)", zap.Error(err))
		return err
	}

	return nil
}

// NotifyBehaviorOnEnteringRiskAreaWorkflow is a workflow to send behavior follow up to
// accounts that enters a risk area (score is lower than 67). [NB_1]
func (n *NudgeWorker) NotifyBehaviorOnEnteringRiskAreaWorkflow(ctx workflow.Context, accountNumber string) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	logger := workflow.GetLogger(ctx)

	err := workflow.ExecuteActivity(ctx, n.NotifyBehaviorNudgeActivity, accountNumber).Get(ctx, nil)
	if err != nil {
		logger.Error("Fail to notify user behavior nudge on risk area (score is lower than 67)", zap.Error(err))
		return err
	}

	return nil
}

// AccountSelfReportedHighRiskFollowUpWorkflow is a workflow to follow up an account if it in risk by
// it self reported symptoms [NB_3-2]
// There are two activities involved.
// 1. `CheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivity` checks if an account has reported a symptoms in the past three days.
// 2. If it has, `NotifyBehaviorFollowUpWhenSelfIsInHighRiskActivity` will send notification to the account.
func (n *NudgeWorker) AccountSelfReportedHighRiskFollowUpWorkflow(ctx workflow.Context, accountNumber string) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	logger := workflow.GetLogger(ctx)

	selector := workflow.NewSelector(ctx)

	timerCancelCtx, _ := workflow.WithCancel(ctx)
	timerFuture := workflow.NewTimer(timerCancelCtx, AccountHighRiskCheckInterval)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		logger.Info("Start periodically account self high risk nudge follow up")
	})

	selector.Select(ctx)

	var shouldFollowUp bool
	if err := workflow.ExecuteActivity(ctx, n.CheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivity, accountNumber).Get(ctx, &shouldFollowUp); err != nil {
		if err.Error() == background.ErrStopRenewWorkflow.Error() {
			logger.Info("Stop following high risk for account (no symptoms in the past)", zap.Any("accountNumber", accountNumber))
			return nil
		}
		logger.Error("Fail to check if an account needs to follow up", zap.Error(err))
		return workflow.NewContinueAsNewError(ctx, n.AccountSelfReportedHighRiskFollowUpWorkflow, accountNumber)
	}

	if shouldFollowUp {
		if err := workflow.ExecuteActivity(ctx, n.NotifyBehaviorFollowUpWhenSelfIsInHighRiskActivity, accountNumber, schema.NudgeBehaviorOnSelfHighRiskSymptoms).Get(ctx, nil); err != nil {
			logger.Error("Fail to send notification user", zap.Error(err))
		}
	}

	return workflow.NewContinueAsNewError(ctx, n.AccountSelfReportedHighRiskFollowUpWorkflow, accountNumber)
}
