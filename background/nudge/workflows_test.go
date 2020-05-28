package nudge

import (
	"context"
	"testing"

	"github.com/bitmark-inc/autonomy-api/background"
	"github.com/bitmark-inc/autonomy-api/external/cadence"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"
	"go.uber.org/zap"
)

type NudgeWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env               *testsuite.TestWorkflowEnvironment
	worker            *NudgeWorker
	testAccountNumber string
}

func (ts *NudgeWorkflowTestSuite) SetupSuite() {
	ts.SetLogger(zap.NewNop())
	ts.testAccountNumber = "e5KNBJCzwBqAyQzKx1pv8CR4MacrUBBTQpWwAbmcLbYNsEg5WS"
	ts.worker = NewNudgeWorker("test", nil)
}

func (ts *NudgeWorkflowTestSuite) SetupTest() {
	ts.env = ts.NewTestWorkflowEnvironment()
	ts.env.SetWorkerOptions(worker.Options{
		DataConverter: cadence.NewMsgPackDataConverter(),
	})
}

func (ts *NudgeWorkflowTestSuite) TestAccountSelfReportedHighRiskFollowUpWorkflowNoFollowThisTime() {
	shouldFollow := false
	ts.env.OnActivity(ts.worker.CheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivity, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string) (bool, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return shouldFollow, nil
		})

	ts.env.ExecuteWorkflow(ts.worker.AccountSelfReportedHighRiskFollowUpWorkflow, ts.testAccountNumber)

	ts.True(ts.env.IsWorkflowCompleted())
	ts.Error(ts.env.GetWorkflowError(), "ContinueAsNew")
}

func (ts *NudgeWorkflowTestSuite) TestAccountSelfReportedHighRiskFollowUpWorkflowNoSymptomInPast() {
	shouldFollow := false
	ts.env.OnActivity(ts.worker.CheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivity, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string) (bool, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return shouldFollow, background.ErrStopRenewWorkflow
		})

	ts.env.ExecuteWorkflow(ts.worker.AccountSelfReportedHighRiskFollowUpWorkflow, ts.testAccountNumber)

	ts.True(ts.env.IsWorkflowCompleted())
	ts.NoError(ts.env.GetWorkflowError())
}

func (ts *NudgeWorkflowTestSuite) TestAccountSelfReportedHighRiskFollowUpWorkflowShouldFollow() {
	shouldFollow := true

	ts.env.OnActivity(ts.worker.CheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivity, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string) (bool, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return shouldFollow, nil
		})

	ts.env.OnActivity(ts.worker.NotifyBehaviorFollowUpWhenSelfIsInHighRiskActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string, nudgeType schema.NudgeType) error {
			ts.Equal(ts.testAccountNumber, accountNumber)
			ts.Equal(schema.NudgeBehaviorOnSelfHighRiskSymptoms, nudgeType)
			return nil
		})

	ts.env.ExecuteWorkflow(ts.worker.AccountSelfReportedHighRiskFollowUpWorkflow, ts.testAccountNumber)

	ts.True(ts.env.IsWorkflowCompleted())
	ts.Error(ts.env.GetWorkflowError(), "ContinueAsNew")
}

func (ts *NudgeWorkflowTestSuite) TestNotifyBehaviorFollowUpOnEnteringSymptomSpikeAreaWorkflow() {
	ts.env.OnActivity(ts.worker.NotifyBehaviorFollowUpWhenSelfIsInHighRiskActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string, nudgeType schema.NudgeType) error {
			ts.Equal(ts.testAccountNumber, accountNumber)
			ts.Equal(schema.NudgeBehaviorOnSymptomSpikeArea, nudgeType)
			return nil
		})

	ts.env.ExecuteWorkflow(ts.worker.NotifyBehaviorFollowUpOnEnteringSymptomSpikeAreaWorkflow, ts.testAccountNumber)

	ts.True(ts.env.IsWorkflowCompleted())
	ts.NoError(ts.env.GetWorkflowError())
}

func (ts *NudgeWorkflowTestSuite) TestNotifyBehaviorOnEnteringRiskAreaWorkflow() {
	ts.env.OnActivity(ts.worker.NotifyBehaviorNudgeActivity, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string) error {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return nil
		})

	ts.env.ExecuteWorkflow(ts.worker.NotifyBehaviorOnEnteringRiskAreaWorkflow, ts.testAccountNumber)
	ts.True(ts.env.IsWorkflowCompleted())
	ts.NoError(ts.env.GetWorkflowError())
}

func (ts *NudgeWorkflowTestSuite) TestNotifySymptomSpikeWorkflowNoReceiver() {
	symptoms := []schema.Symptom{}

	ts.env.OnActivity(ts.worker.GetNotificationReceiverActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber, poiID string) ([]string, error) {
			ts.Equal("fake-poi", poiID)
			// return with no receivers
			return nil, nil
		})

	ts.env.ExecuteWorkflow(ts.worker.NotifySymptomSpikeWorkflow, "", "fake-poi", symptoms)
	ts.True(ts.env.IsWorkflowCompleted())
	ts.NoError(ts.env.GetWorkflowError())
}

func (ts *NudgeWorkflowTestSuite) TestNotifySymptomSpikeWorkflowByAccountNumber() {

	symptoms := []schema.Symptom{}

	ts.env.OnActivity(ts.worker.GetNotificationReceiverActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber, poiID string) ([]string, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return []string{ts.testAccountNumber}, nil
		})

	ts.env.OnActivity(ts.worker.NotifySymptomSpikeActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string, symptoms []schema.Symptom) error {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return nil
		})

	ts.env.ExecuteWorkflow(ts.worker.NotifySymptomSpikeWorkflow, ts.testAccountNumber, "", symptoms)
	ts.True(ts.env.IsWorkflowCompleted())
	ts.NoError(ts.env.GetWorkflowError())
}

func (ts *NudgeWorkflowTestSuite) TestSymptomFollowUpNudgeWorkflowOneNudge() {
	symptoms := []schema.Symptom{
		schema.COVID19Symptoms[0],
	}

	ts.env.OnActivity(ts.worker.SymptomsNeedFollowUpActivity, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string) ([]schema.Symptom, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return symptoms, nil
		})

	ts.env.OnActivity("NotifySymptomFollowUpActivity", mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string, symptoms []schema.Symptom) error {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return nil
		})

	ts.env.ExecuteWorkflow(ts.worker.SymptomFollowUpNudgeWorkflow, ts.testAccountNumber)

	ts.env.AssertNumberOfCalls(ts.T(), "NotifySymptomFollowUpActivity", 1)
	ts.True(ts.env.IsWorkflowCompleted())
	ts.Error(ts.env.GetWorkflowError(), "ContinueAsNew")
}

func (ts *NudgeWorkflowTestSuite) TestSymptomFollowUpNudgeWorkflowNoNudge() {
	symptoms := []schema.Symptom{}

	ts.env.OnActivity(ts.worker.SymptomsNeedFollowUpActivity, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string) ([]schema.Symptom, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return symptoms, nil
		})

	ts.env.OnActivity("NotifySymptomFollowUpActivity", mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string, symptoms []schema.Symptom) error {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return nil
		})

	ts.env.ExecuteWorkflow(ts.worker.SymptomFollowUpNudgeWorkflow, ts.testAccountNumber)

	ts.env.AssertNumberOfCalls(ts.T(), "NotifySymptomFollowUpActivity", 0)
	ts.True(ts.env.IsWorkflowCompleted())
	ts.Error(ts.env.GetWorkflowError(), "ContinueAsNew")
}

func TestNudgeWorkflow(t *testing.T) {
	suite.Run(t, new(NudgeWorkflowTestSuite))
}
