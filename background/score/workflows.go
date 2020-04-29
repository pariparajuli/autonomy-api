package score

import (
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

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
		sentry.CaptureException(err)
		return workflow.NewContinueAsNewError(ctx, s.POIStateUpdateWorkflow, id)
	}

	updatedAccounts := make([]string, 0)
	if err := workflow.ExecuteActivity(ctx, s.RefreshLocationStateActivity, "", id, metric).Get(ctx, &updatedAccounts); err != nil {
		logger.Error("Fail to update POI state for accounts.", zap.Error(err))
		sentry.CaptureException(err)
		return workflow.NewContinueAsNewError(ctx, s.POIStateUpdateWorkflow, id)
	}

	if len(updatedAccounts) > 0 {
		err := workflow.ExecuteActivity(ctx, s.NotifyLocationStateActivity, id, updatedAccounts).Get(ctx, nil)
		if err != nil {
			logger.Error("Fail to notify users", zap.Error(err))
			sentry.CaptureException(err)
			return workflow.NewContinueAsNewError(ctx, s.POIStateUpdateWorkflow, id)
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
		sentry.CaptureException(err)
		return workflow.NewContinueAsNewError(ctx, s.AccountStateUpdateWorkflow, accountNumber)
	}

	updatedAccounts := make([]string, 0)
	if err := workflow.ExecuteActivity(ctx, s.RefreshLocationStateActivity, accountNumber, "", metric).Get(ctx, &updatedAccounts); err != nil {
		logger.Error("Fail to update POI state for accounts.", zap.Error(err))
		sentry.CaptureException(err)
		return workflow.NewContinueAsNewError(ctx, s.AccountStateUpdateWorkflow, accountNumber)
	}

	if len(updatedAccounts) > 0 {
		err := workflow.ExecuteActivity(ctx, s.NotifyLocationStateActivity, "", updatedAccounts).Get(ctx, nil)
		if err != nil {
			logger.Error("Fail to notify users", zap.Error(err))
			sentry.CaptureException(err)
			return workflow.NewContinueAsNewError(ctx, s.AccountStateUpdateWorkflow, accountNumber)
		}
	}

	return workflow.NewContinueAsNewError(ctx, s.AccountStateUpdateWorkflow, accountNumber)
}
