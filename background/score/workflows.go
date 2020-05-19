package score

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	cadenceClient "go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/bitmark-inc/autonomy-api/background/nudge"
	"github.com/bitmark-inc/autonomy-api/schema"
)

const (
	POIStateCheckInterval     = 5 * time.Minute
	AccountStateCheckInterval = 5 * time.Minute
)

var activityOptions = workflow.ActivityOptions{
	ScheduleToStartTimeout: time.Minute,
	StartToCloseTimeout:    time.Minute,
	HeartbeatTimeout:       time.Second * 20,
}

func (s *ScoreUpdateWorker) POIStateUpdateWorkflow(ctx workflow.Context, id string) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	signalChan := workflow.GetSignalChannel(ctx, "poiCheckSignal")
	defer signalChan.Close()

	logger := workflow.GetLogger(ctx)
	selector := workflow.NewSelector(ctx)

	timerCancelCtx, cancelTimerHandler := workflow.WithCancel(ctx)
	timerFuture := workflow.NewTimer(timerCancelCtx, POIStateCheckInterval)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		logger.Info("Start periodically POI info updates")
	})

	selector.AddReceive(signalChan, func(c workflow.Channel, more bool) {
		cancelTimerHandler()
		signalChan.Receive(ctx, nil)

		logger.Info("Trigger POI info updates by signal")
	})

	selector.Select(ctx)

	var metric schema.Metric
	err := workflow.ExecuteActivity(ctx, s.CalculatePOIStateActivity, id).Get(ctx, &metric)
	if err != nil {
		logger.Error("Fail to update POI.", zap.Error(err))
		return workflow.NewContinueAsNewError(ctx, s.POIStateUpdateWorkflow, id)
	}

	var np NotificationProfile
	if err := workflow.ExecuteActivity(ctx, s.RefreshLocationStateActivity, "", id, metric).Get(ctx, &np); err != nil {
		logger.Error("Fail to update POI state for accounts.", zap.Error(err))
		sentry.CaptureException(err)
		return workflow.NewContinueAsNewError(ctx, s.POIStateUpdateWorkflow, id)
	}

	if len(np.StateChangedAccounts) > 0 {
		err := workflow.ExecuteActivity(ctx, s.NotifyLocationStateActivity, id, np.StateChangedAccounts).Get(ctx, nil)
		if err != nil {
			logger.Error("Fail to notify users for location state", zap.Error(err))
			sentry.CaptureException(err)
		}
	}

	spikeSymptoms := make([]schema.Symptom, 0)
	if err := workflow.ExecuteActivity(ctx, s.CheckLocationSpikeActivity, metric.Details.Symptoms.LastSpikeList).Get(ctx, &spikeSymptoms); err != nil {
		logger.Error("Fail to get symptom spike", zap.Error(err))
		sentry.CaptureException(err)
		return workflow.NewContinueAsNewError(ctx, s.POIStateUpdateWorkflow, id)
	}

	if len(spikeSymptoms) > 0 {
		for _, a := range np.SymptomsSpikeAccounts {
			cwo := workflow.ChildWorkflowOptions{
				// Do not specify WorkflowID if you want Cadence to generate a unique ID for the child execution.
				WorkflowID:                   fmt.Sprintf("poi-%s-nudge-symptom-spike-%s", id, a),
				TaskList:                     nudge.TaskListName,
				ExecutionStartToCloseTimeout: time.Minute,
				WorkflowIDReusePolicy:        cadenceClient.WorkflowIDReusePolicyAllowDuplicate,
			}

			if err := workflow.ExecuteChildWorkflow(workflow.WithChildOptions(ctx, cwo), "NotifySymptomSpikeWorkflow", a, "", spikeSymptoms).Get(ctx, nil); err != nil {
				logger.Error("NotifySymptomSpikeWorkflow failed.", zap.Error(err))
				sentry.CaptureException(err)
			}
		}
	}

	return workflow.NewContinueAsNewError(ctx, s.POIStateUpdateWorkflow, id)
}

func (s *ScoreUpdateWorker) AccountStateUpdateWorkflow(ctx workflow.Context, accountNumber string) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	signalChan := workflow.GetSignalChannel(ctx, "accountCheckSignal")
	defer signalChan.Close()

	logger := workflow.GetLogger(ctx)

	selector := workflow.NewSelector(ctx)

	timerCancelCtx, cancelTimerHandler := workflow.WithCancel(ctx)
	timerFuture := workflow.NewTimer(timerCancelCtx, AccountStateCheckInterval)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		logger.Info("Start periodically account info updates")
	})

	selector.AddReceive(signalChan, func(c workflow.Channel, more bool) {
		cancelTimerHandler()
		signalChan.Receive(ctx, nil)
		logger.Info("Start account info updates by signal")
	})

	selector.Select(ctx)

	logger.Info("Check if account state color changes")

	var metric schema.Metric
	err := workflow.ExecuteActivity(ctx, s.CalculateAccountStateActivity, accountNumber).Get(ctx, &metric)
	if err != nil {
		logger.Error("Fail to update account state", zap.Error(err))
		return workflow.NewContinueAsNewError(ctx, s.AccountStateUpdateWorkflow, accountNumber)
	}

	var np NotificationProfile
	if err := workflow.ExecuteActivity(ctx, s.RefreshLocationStateActivity, accountNumber, "", metric).Get(ctx, &np); err != nil {
		logger.Error("Fail to update POI state for accounts.", zap.Error(err))
		sentry.CaptureException(err)
		return workflow.NewContinueAsNewError(ctx, s.AccountStateUpdateWorkflow, accountNumber)
	}

	if len(np.StateChangedAccounts) > 0 {
		err := workflow.ExecuteActivity(ctx, s.NotifyLocationStateActivity, "", np.StateChangedAccounts).Get(ctx, nil)
		if err != nil {
			logger.Error("Fail to notify users for location state", zap.Error(err))
			sentry.CaptureException(err)
		}
	}

	if np.RemindGoodBehavior {
		cwo := workflow.ChildWorkflowOptions{
			WorkflowID:                   fmt.Sprintf("account-behavior-on-symptom-score-spike-%s", accountNumber),
			TaskList:                     nudge.TaskListName,
			ExecutionStartToCloseTimeout: time.Minute,
			WorkflowIDReusePolicy:        cadenceClient.WorkflowIDReusePolicyAllowDuplicate,
		}

		if err := workflow.ExecuteChildWorkflow(workflow.WithChildOptions(ctx, cwo), "NotifyBehaviorFollowUpOnEnteringSymptomSpikeAreaWorkflow", accountNumber).Get(ctx, nil); err != nil {
			logger.Error("NotifyBehaviorFollowUpOnEnteringSymptomSpikeAreaWorkflow failed.", zap.Error(err))
			sentry.CaptureException(err)
		}
	}

	if np.ReportRiskArea {
		cwo := workflow.ChildWorkflowOptions{
			WorkflowID:                   fmt.Sprintf("account-behavior-on-risk-area-%s", accountNumber),
			TaskList:                     nudge.TaskListName,
			ExecutionStartToCloseTimeout: time.Minute,
			WorkflowIDReusePolicy:        cadenceClient.WorkflowIDReusePolicyAllowDuplicate,
		}

		if err := workflow.ExecuteChildWorkflow(workflow.WithChildOptions(ctx, cwo), "NotifyBehaviorOnEnteringRiskAreaWorkflow", accountNumber).Get(ctx, nil); err != nil {
			logger.Error("NotifyBehaviorOnEnteringRiskAreaWorkflow failed.", zap.Error(err))
			sentry.CaptureException(err)
		}
	}

	spikeSymptoms := make([]schema.Symptom, 0)
	if err := workflow.ExecuteActivity(ctx, s.CheckLocationSpikeActivity, metric.Details.Symptoms.LastSpikeList).Get(ctx, &spikeSymptoms); err != nil {
		logger.Error("Fail to get symptom spike", zap.Error(err))
		sentry.CaptureException(err)
		return workflow.NewContinueAsNewError(ctx, s.AccountStateUpdateWorkflow, accountNumber)
	}

	if len(spikeSymptoms) > 0 {
		for _, a := range np.SymptomsSpikeAccounts {
			cwo := workflow.ChildWorkflowOptions{
				// Do not specify WorkflowID if you want Cadence to generate a unique ID for the child execution.
				WorkflowID:                   fmt.Sprintf("account-nudge-symptom-spike-%s", accountNumber),
				TaskList:                     nudge.TaskListName,
				ExecutionStartToCloseTimeout: time.Minute,
				WorkflowIDReusePolicy:        cadenceClient.WorkflowIDReusePolicyAllowDuplicate,
			}

			if err := workflow.ExecuteChildWorkflow(workflow.WithChildOptions(ctx, cwo), "NotifySymptomSpikeWorkflow", a, "", spikeSymptoms).Get(ctx, nil); err != nil {
				logger.Error("NotifySymptomSpikeWorkflow failed.", zap.Error(err))
				sentry.CaptureException(err)
			}
		}
	}

	return workflow.NewContinueAsNewError(ctx, s.AccountStateUpdateWorkflow, accountNumber)
}
