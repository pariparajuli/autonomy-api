package score

import (
	"context"
	"testing"

	"github.com/bitmark-inc/autonomy-api/background"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

var (
	twoSpikeMetric = &schema.Metric{
		Details: schema.Details{
			Symptoms: schema.SymptomDetail{
				LastSpikeList: []string{
					schema.COVID19Symptoms[0].ID,
					schema.COVID19Symptoms[1].ID,
				},
			},
		},
	}
)

var (
	fakeAccount1 = "fcqu8Deozrzv6pQ5EqSsdvAHG1SbTafHqviUjVvP1mDmbPyiBU"
	fakeAccount2 = "eEfqMcw7ExsoUhULQ7H41r5avLJxpzPWf4vVm6pGWB1o2wvyjR"
)

type ScoreWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env               *testsuite.TestWorkflowEnvironment
	worker            *ScoreUpdateWorker
	testAccountNumber string
	testPOIID         string
}

func (ts *ScoreWorkflowTestSuite) SetupSuite() {
	ts.SetLogger(zap.NewNop())

	ts.testAccountNumber = "e5KNBJCzwBqAyQzKx1pv8CR4MacrUBBTQpWwAbmcLbYNsEg5WS"
	ts.testPOIID = "5e9806ae554b311b328e2f91"
	ts.worker = NewScoreUpdateWorker("test", nil)
}

func (ts *ScoreWorkflowTestSuite) SetupTest() {
	ts.env = ts.NewTestWorkflowEnvironment()
	ts.env.SetWorkerOptions(worker.Options{
		DataConverter: background.NewMsgPackDataConverter(),
	})
}

// TestAccountStateUpdateWorkflowNormalRun tests regular run of `AccountStateUpdateWorkflow` without any notification
func (ts *ScoreWorkflowTestSuite) TestAccountStateUpdateWorkflowNormalRun() {
	ts.env.OnActivity(ts.worker.CalculateAccountStateActivity, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string) (*schema.Metric, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return twoSpikeMetric, nil
		})

	ts.env.OnActivity(ts.worker.RefreshLocationStateActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber, poiID string, metric schema.Metric) (*NotificationProfile, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			ts.Equal("", poiID)
			return &NotificationProfile{}, nil
		})

	ts.env.OnActivity(ts.worker.CheckLocationSpikeActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, symptoms []string) ([]schema.Symptom, error) {
			return []schema.Symptom{}, nil
		})

	ts.env.ExecuteWorkflow(ts.worker.AccountStateUpdateWorkflow, ts.testAccountNumber)

	ts.env.AssertNumberOfCalls(ts.T(), "CalculateAccountStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "RefreshLocationStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "CheckLocationSpikeActivity", 1)
	ts.True(ts.env.IsWorkflowCompleted())
	ts.EqualError(ts.env.GetWorkflowError(), "ContinueAsNew")
}

// TestAccountStateUpdateWorkflowStateChange validate whether
// `NotifyLocationStateActivity` is triggered when the state of a location changes
func (ts *ScoreWorkflowTestSuite) TestAccountStateUpdateWorkflowStateChange() {
	ts.env.OnActivity(ts.worker.CalculateAccountStateActivity, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string) (*schema.Metric, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return twoSpikeMetric, nil
		})

	ts.env.OnActivity(ts.worker.RefreshLocationStateActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber, poiID string, metric schema.Metric) (*NotificationProfile, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			ts.Equal("", poiID)
			return &NotificationProfile{
				StateChangedAccounts: []string{fakeAccount1, fakeAccount2},
			}, nil
		})

	ts.env.OnActivity(ts.worker.NotifyLocationStateActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string, accounts []string) error {
			ts.Equal("", id)
			ts.Equal([]string{fakeAccount1, fakeAccount2}, accounts)
			return nil
		})

	ts.env.OnActivity(ts.worker.CheckLocationSpikeActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, symptoms []string) ([]schema.Symptom, error) {
			ts.Equal(twoSpikeMetric.Details.Symptoms.LastSpikeList, symptoms)
			return []schema.Symptom{}, nil
		})

	ts.env.ExecuteWorkflow(ts.worker.AccountStateUpdateWorkflow, ts.testAccountNumber)

	ts.env.AssertNumberOfCalls(ts.T(), "CalculateAccountStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "RefreshLocationStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "NotifyLocationStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "CheckLocationSpikeActivity", 1)

	ts.True(ts.env.IsWorkflowCompleted())
	ts.EqualError(ts.env.GetWorkflowError(), "ContinueAsNew")
}

// TestAccountStateUpdateWorkflowNotifySpike validate whether
// `NotifySymptomSpikeWorkflow` is triggered when there are new symptom spikes
// found for the account
func (ts *ScoreWorkflowTestSuite) TestAccountStateUpdateWorkflowNotifySpike() {
	ts.env.OnActivity(ts.worker.CalculateAccountStateActivity, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string) (*schema.Metric, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return twoSpikeMetric, nil
		})

	ts.env.OnActivity(ts.worker.RefreshLocationStateActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber, poiID string, metric schema.Metric) (*NotificationProfile, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			ts.Equal("", poiID)
			return &NotificationProfile{
				SymptomsSpikeAccounts: []string{fakeAccount1, fakeAccount2},
			}, nil
		})

	ts.env.OnActivity(ts.worker.CheckLocationSpikeActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, symptoms []string) ([]schema.Symptom, error) {
			ts.Equal(twoSpikeMetric.Details.Symptoms.LastSpikeList, symptoms)
			return []schema.Symptom{
				schema.COVID19Symptoms[0],
				schema.COVID19Symptoms[1],
			}, nil
		})

	ts.env.OnWorkflow("NotifySymptomSpikeWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx workflow.Context, accountNumber string, poiID string, symptoms []schema.Symptom) error {
			ts.Equal([]schema.Symptom{
				schema.COVID19Symptoms[0],
				schema.COVID19Symptoms[1],
			}, symptoms)
			return nil
		})

	ts.env.ExecuteWorkflow(ts.worker.AccountStateUpdateWorkflow, ts.testAccountNumber)

	ts.env.AssertNumberOfCalls(ts.T(), "CalculateAccountStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "RefreshLocationStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "CheckLocationSpikeActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "NotifySymptomSpikeWorkflow", 2)
	ts.True(ts.env.IsWorkflowCompleted())
	ts.EqualError(ts.env.GetWorkflowError(), "ContinueAsNew")
}

// TestAccountStateUpdateWorkflowRemindGoodBehaviorOnEnteringSymptomSpikeArea validate whether
// `NotifyBehaviorFollowUpOnEnteringSymptomSpikeAreaWorkflow` is invoked if `RemindGoodBehavior` is true.
// This only happens on `AccountStateUpdateWorkflow`
func (ts *ScoreWorkflowTestSuite) TestAccountStateUpdateWorkflowRemindGoodBehaviorOnEnteringSymptomSpikeArea() {
	ts.env.OnActivity(ts.worker.CalculateAccountStateActivity, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string) (*schema.Metric, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return twoSpikeMetric, nil
		})

	ts.env.OnActivity(ts.worker.RefreshLocationStateActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber, poiID string, metric schema.Metric) (*NotificationProfile, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			ts.Equal("", poiID)
			return &NotificationProfile{
				RemindGoodBehavior: true,
			}, nil
		})

	ts.env.OnWorkflow("NotifyBehaviorFollowUpOnEnteringSymptomSpikeAreaWorkflow", mock.Anything, mock.Anything).Return(
		func(ctx workflow.Context, accountNumber string) error {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return nil
		})

	ts.env.OnActivity(ts.worker.CheckLocationSpikeActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, symptoms []string) ([]schema.Symptom, error) {
			ts.Equal(twoSpikeMetric.Details.Symptoms.LastSpikeList, symptoms)
			return []schema.Symptom{}, nil
		})

	ts.env.ExecuteWorkflow(ts.worker.AccountStateUpdateWorkflow, ts.testAccountNumber)

	ts.env.AssertNumberOfCalls(ts.T(), "CalculateAccountStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "RefreshLocationStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "NotifyBehaviorFollowUpOnEnteringSymptomSpikeAreaWorkflow", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "CheckLocationSpikeActivity", 1)
	ts.True(ts.env.IsWorkflowCompleted())
	ts.EqualError(ts.env.GetWorkflowError(), "ContinueAsNew")
}

// TestAccountStateUpdateWorkflowOnEnteringRiskArea validate whether
// `NotifyBehaviorOnEnteringRiskAreaWorkflow` is invoked if `ReportRiskArea` is true.
// This only happens on `AccountStateUpdateWorkflow`
func (ts *ScoreWorkflowTestSuite) TestAccountStateUpdateWorkflowOnEnteringRiskArea() {
	ts.env.OnActivity(ts.worker.CalculateAccountStateActivity, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber string) (*schema.Metric, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return twoSpikeMetric, nil
		})

	ts.env.OnActivity(ts.worker.RefreshLocationStateActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber, poiID string, metric schema.Metric) (*NotificationProfile, error) {
			ts.Equal(ts.testAccountNumber, accountNumber)
			ts.Equal("", poiID)
			return &NotificationProfile{
				ReportRiskArea: true,
			}, nil
		})

	ts.env.OnWorkflow("NotifyBehaviorOnEnteringRiskAreaWorkflow", mock.Anything, mock.Anything).Return(
		func(ctx workflow.Context, accountNumber string) error {
			ts.Equal(ts.testAccountNumber, accountNumber)
			return nil
		})

	ts.env.OnActivity(ts.worker.CheckLocationSpikeActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, symptoms []string) ([]schema.Symptom, error) {
			ts.Equal(twoSpikeMetric.Details.Symptoms.LastSpikeList, symptoms)
			return []schema.Symptom{}, nil
		})

	ts.env.ExecuteWorkflow(ts.worker.AccountStateUpdateWorkflow, ts.testAccountNumber)

	ts.env.AssertNumberOfCalls(ts.T(), "CalculateAccountStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "RefreshLocationStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "NotifyBehaviorOnEnteringRiskAreaWorkflow", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "CheckLocationSpikeActivity", 1)
	ts.True(ts.env.IsWorkflowCompleted())
	ts.EqualError(ts.env.GetWorkflowError(), "ContinueAsNew")
}

// TestAccountStateUpdateWorkflowNormalRun tests regular run of `POIStateUpdateWorkflow` without any notification
func (ts *ScoreWorkflowTestSuite) TestPOIStateUpdateWorkflowNormalRun() {
	ts.env.OnActivity(ts.worker.CalculatePOIStateActivity, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) (*schema.Metric, error) {
			ts.Equal(ts.testPOIID, id)
			return twoSpikeMetric, nil
		})

	ts.env.OnActivity(ts.worker.RefreshLocationStateActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber, poiID string, metric schema.Metric) (*NotificationProfile, error) {
			ts.Equal("", accountNumber)
			ts.Equal(ts.testPOIID, poiID)
			return &NotificationProfile{}, nil
		})

	ts.env.OnActivity(ts.worker.CheckLocationSpikeActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, symptoms []string) ([]schema.Symptom, error) {
			ts.Equal(twoSpikeMetric.Details.Symptoms.LastSpikeList, symptoms)
			return []schema.Symptom{
				schema.COVID19Symptoms[0],
				schema.COVID19Symptoms[1],
			}, nil
		})

	ts.env.ExecuteWorkflow(ts.worker.POIStateUpdateWorkflow, ts.testPOIID)

	ts.env.AssertNumberOfCalls(ts.T(), "CalculatePOIStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "RefreshLocationStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "CheckLocationSpikeActivity", 1)
	ts.True(ts.env.IsWorkflowCompleted())
	ts.EqualError(ts.env.GetWorkflowError(), "ContinueAsNew")
}

// TestPOIStateUpdateWorkflowStateChange tests POI state updates with
// where a state change happens
func (ts *ScoreWorkflowTestSuite) TestPOIStateUpdateWorkflowStateChange() {
	ts.env.OnActivity(ts.worker.CalculatePOIStateActivity, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) (*schema.Metric, error) {
			ts.Equal(ts.testPOIID, id)
			return twoSpikeMetric, nil
		})

	ts.env.OnActivity(ts.worker.RefreshLocationStateActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber, poiID string, metric schema.Metric) (*NotificationProfile, error) {
			ts.Equal("", accountNumber)
			ts.Equal(ts.testPOIID, poiID)
			return &NotificationProfile{
				StateChangedAccounts: []string{fakeAccount1, fakeAccount2},
			}, nil
		})

	ts.env.OnActivity(ts.worker.NotifyLocationStateActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string, accounts []string) error {
			ts.Equal(ts.testPOIID, id)
			ts.Equal([]string{fakeAccount1, fakeAccount2}, accounts)
			return nil
		})

	ts.env.OnActivity(ts.worker.CheckLocationSpikeActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, symptoms []string) ([]schema.Symptom, error) {
			ts.Equal(twoSpikeMetric.Details.Symptoms.LastSpikeList, symptoms)
			return []schema.Symptom{
				schema.COVID19Symptoms[0],
				schema.COVID19Symptoms[1],
			}, nil
		})

	ts.env.ExecuteWorkflow(ts.worker.POIStateUpdateWorkflow, ts.testPOIID)

	ts.env.AssertNumberOfCalls(ts.T(), "CalculatePOIStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "RefreshLocationStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "NotifyLocationStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "CheckLocationSpikeActivity", 1)
	ts.True(ts.env.IsWorkflowCompleted())
	ts.EqualError(ts.env.GetWorkflowError(), "ContinueAsNew")
}

// TestPOIStateUpdateWorkflowWithSymptomSpike tests POI state updates with
// a symptom spike
func (ts *ScoreWorkflowTestSuite) TestPOIStateUpdateWorkflowWithSymptomSpike() {
	ts.env.OnActivity(ts.worker.CalculatePOIStateActivity, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, id string) (*schema.Metric, error) {
			ts.Equal(ts.testPOIID, id)
			return twoSpikeMetric, nil
		})

	ts.env.OnActivity(ts.worker.RefreshLocationStateActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, accountNumber, poiID string, metric schema.Metric) (*NotificationProfile, error) {
			ts.Equal("", accountNumber)
			ts.Equal(ts.testPOIID, poiID)
			return &NotificationProfile{
				SymptomsSpikeAccounts: []string{fakeAccount1, fakeAccount2},
			}, nil
		})

	ts.env.OnActivity(ts.worker.CheckLocationSpikeActivity, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, symptoms []string) ([]schema.Symptom, error) {
			ts.Equal(twoSpikeMetric.Details.Symptoms.LastSpikeList, symptoms)
			return []schema.Symptom{
				schema.COVID19Symptoms[0],
				schema.COVID19Symptoms[1],
			}, nil
		})

	ts.env.OnWorkflow("NotifySymptomSpikeWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx workflow.Context, accountNumber string, poiID string, symptoms []schema.Symptom) error {
			ts.Equal([]schema.Symptom{
				schema.COVID19Symptoms[0],
				schema.COVID19Symptoms[1],
			}, symptoms)
			return nil
		})

	ts.env.ExecuteWorkflow(ts.worker.POIStateUpdateWorkflow, ts.testPOIID)

	ts.env.AssertNumberOfCalls(ts.T(), "CalculatePOIStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "RefreshLocationStateActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "CheckLocationSpikeActivity", 1)
	ts.env.AssertNumberOfCalls(ts.T(), "NotifySymptomSpikeWorkflow", 2)
	ts.True(ts.env.IsWorkflowCompleted())
	ts.EqualError(ts.env.GetWorkflowError(), "ContinueAsNew")
}

func TestScoreUpdateWorkflow(t *testing.T) {
	suite.Run(t, new(ScoreWorkflowTestSuite))
}
