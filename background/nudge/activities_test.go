package nudge

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"
	"go.uber.org/zap"

	"github.com/bitmark-inc/autonomy-api/background"
	"github.com/bitmark-inc/autonomy-api/mocks"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/utils"
)

type anyTimestamp struct {
}

func (t *anyTimestamp) Matches(x interface{}) bool {
	_, ok := x.(int64)
	return ok
}

func (t *anyTimestamp) String() string {
	return "any timestamp"
}

func AnyTimestamp() gomock.Matcher {
	return &anyTimestamp{}
}

type NudgeActivityTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env               *testsuite.TestActivityEnvironment
	worker            *NudgeWorker
	mongoMock         *mocks.MockMongoStore
	testAccountNumber string
}

func (ts *NudgeActivityTestSuite) SetupSuite() {
	ts.SetLogger(zap.NewNop())
	ts.testAccountNumber = "e5KNBJCzwBqAyQzKx1pv8CR4MacrUBBTQpWwAbmcLbYNsEg5WS"
	ctrl := gomock.NewController(ts.T())

	mongoMock = mocks.NewMockMongoStore(ctrl)
	nudgeWorker.mongo = mongoMock
	ts.mongoMock = mongoMock
	ts.worker = nudgeWorker
}

func (ts *NudgeActivityTestSuite) SetupTest() {
	ts.env = ts.NewTestActivityEnvironment()
	ts.env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: context.Background(),
	})
}

func (ts *NudgeActivityTestSuite) TestGetNotificationReceiverActivityByAccount() {
	values, err := ts.env.ExecuteActivity(ts.worker.GetNotificationReceiverActivity, ts.testAccountNumber, "")
	ts.NoError(err)
	accounts := make([]string, 0)
	err = values.Get(&accounts)
	ts.NoError(err)
	ts.Len(accounts, 1)
	ts.Equal(accounts[0], ts.testAccountNumber)
}

func (ts *NudgeActivityTestSuite) TestGetNotificationReceiverActivityByPOI() {
	mockID := "1234567890"
	ts.mongoMock.
		EXPECT().
		GetProfilesByPOI(gomock.Eq(mockID)).
		Return([]schema.Profile{
			{
				ID:            "test1",
				AccountNumber: "eU9hA8vDQncHfuVtUeeNzKwZKo793hThDYECtYHRHbpnN6Hf93",
			},
			{
				ID:            "test2",
				AccountNumber: "egDG3D7aMzyU9pVgpyzX5BQQwcscFUPbKjrLsSVSiiLwd4jxKE",
			},
		}, nil)

	values, err := ts.env.ExecuteActivity(ts.worker.GetNotificationReceiverActivity, "", mockID)
	ts.NoError(err)
	accounts := make([]string, 0)
	err = values.Get(&accounts)
	ts.NoError(err)
	ts.Len(accounts, 2)
}

func (ts *NudgeActivityTestSuite) TestGetNotificationReceiverActivityWithoutAnything() {
	values, err := ts.env.ExecuteActivity(ts.worker.GetNotificationReceiverActivity, "", "")
	ts.Error(err, background.ErrBothAccountPOIEmpty.Error())
	ts.Nil(values)
}

func (ts *NudgeActivityTestSuite) TestSymptomsNeedFollowUpActivityReportedYesterdayShouldNudgeInTheMorning() {
	now = func() time.Time {
		t := time.Now()
		return time.Date(t.Year(), t.Month(), t.Day(), 10, 0, 0, 0, utils.GetLocation("GMT+8"))
	}

	ts.mongoMock.
		EXPECT().
		GetProfile(gomock.Eq(ts.testAccountNumber)).
		Return(&schema.Profile{
			ID:            "test-account",
			AccountNumber: ts.testAccountNumber,
			Timezone:      "GMT+8",
			LastNudge: schema.NudgeTime{
				schema.NudgeSymptomFollowUp: time.Now().Add(-24 * time.Hour), // nedged yesterday
			},
		}, nil)

	ts.mongoMock.
		EXPECT().
		GetReportedSymptoms(
			gomock.Eq(ts.testAccountNumber),
			AnyTimestamp(),
			gomock.Eq(int64(1)),
			gomock.Eq(""),
		).
		Return([]*schema.SymptomReportData{
			{
				ProfileID:     "test-account",
				AccountNumber: ts.testAccountNumber,
				Timestamp:     time.Now().Add(-24 * time.Hour).Unix(), // reported yesterday
				Symptoms: []schema.Symptom{
					schema.COVID19Symptoms[0],
					schema.COVID19Symptoms[1],
				},
			},
		}, nil)

	values, err := ts.env.ExecuteActivity(ts.worker.SymptomsNeedFollowUpActivity, ts.testAccountNumber)
	ts.NoError(err)

	symptoms := make([]schema.Symptom, 0)
	err = values.Get(&symptoms)
	ts.NoError(err)
	ts.Len(symptoms, 2)
}

func (ts *NudgeActivityTestSuite) TestSymptomsNeedFollowUpActivityReportedTodayWillNotReceiveAnyNudge() {
	now = func() time.Time {
		t := time.Now()
		return time.Date(t.Year(), t.Month(), t.Day(), 10, 0, 0, 0, utils.GetLocation("GMT+8"))
	}

	ts.mongoMock.
		EXPECT().
		GetProfile(gomock.Eq(ts.testAccountNumber)).
		Return(&schema.Profile{
			ID:            "test-account",
			AccountNumber: ts.testAccountNumber,
			Timezone:      "GMT+8",
			LastNudge: schema.NudgeTime{
				schema.NudgeSymptomFollowUp: time.Now().Add(-24 * time.Hour), // nedged yesterday
			},
		}, nil)

	ts.mongoMock.
		EXPECT().
		GetReportedSymptoms(
			gomock.Eq(ts.testAccountNumber),
			AnyTimestamp(),
			gomock.Eq(int64(1)),
			gomock.Eq(""),
		).
		Return([]*schema.SymptomReportData{
			{
				ProfileID:     "test-account",
				AccountNumber: ts.testAccountNumber,
				Timestamp:     now().Unix(), // reported today
				Symptoms: []schema.Symptom{
					schema.COVID19Symptoms[0],
					schema.COVID19Symptoms[1],
				},
			},
		}, nil)

	values, err := ts.env.ExecuteActivity(ts.worker.SymptomsNeedFollowUpActivity, ts.testAccountNumber)
	ts.NoError(err)

	symptoms := make([]schema.Symptom, 0)
	err = values.Get(&symptoms)
	ts.NoError(err)
	ts.Len(symptoms, 0)
}

func (ts *NudgeActivityTestSuite) TestSymptomsNeedFollowUpActivityReportedYesterdayButNotNudgedInTheMorningShouldNotNudgeInTheAfternoon() {
	now = func() time.Time {
		t := time.Now()
		return time.Date(t.Year(), t.Month(), t.Day(), 15, 0, 0, 0, utils.GetLocation("GMT+8"))
	}

	ts.mongoMock.
		EXPECT().
		GetProfile(gomock.Eq(ts.testAccountNumber)).
		Return(&schema.Profile{
			ID:            "test-account",
			AccountNumber: ts.testAccountNumber,
			Timezone:      "GMT+8",
			LastNudge: schema.NudgeTime{
				schema.NudgeSymptomFollowUp: time.Now().Add(-24 * time.Hour), // nedged yesterday
			},
		}, nil)

	ts.mongoMock.
		EXPECT().
		GetReportedSymptoms(
			gomock.Eq(ts.testAccountNumber),
			AnyTimestamp(),
			gomock.Eq(int64(1)),
			gomock.Eq(""),
		).
		Return([]*schema.SymptomReportData{
			{
				ProfileID:     "test-account",
				AccountNumber: ts.testAccountNumber,
				Timestamp:     time.Now().Add(-24 * time.Hour).Unix(), // reported yesterday
				Symptoms: []schema.Symptom{
					schema.COVID19Symptoms[0],
					schema.COVID19Symptoms[1],
				},
			},
		}, nil)

	values, err := ts.env.ExecuteActivity(ts.worker.SymptomsNeedFollowUpActivity, ts.testAccountNumber)
	ts.NoError(err)

	symptoms := make([]schema.Symptom, 0)
	err = values.Get(&symptoms)
	ts.NoError(err)
	ts.Len(symptoms, 0)
}

func (ts *NudgeActivityTestSuite) TestSymptomsNeedFollowUpActivityReportedYesterdayAndNudgedInTheMorningShouldNudgeInTheAfternoon() {
	now = func() time.Time {
		t := time.Now()
		return time.Date(t.Year(), t.Month(), t.Day(), 15, 0, 0, 0, utils.GetLocation("GMT+8"))
	}

	ts.mongoMock.
		EXPECT().
		GetProfile(gomock.Eq(ts.testAccountNumber)).
		Return(&schema.Profile{
			ID:            "test-account",
			AccountNumber: ts.testAccountNumber,
			Timezone:      "GMT+8",
			LastNudge: schema.NudgeTime{
				schema.NudgeSymptomFollowUp: now().Add(-5 * time.Hour), // nedged yesterday
			},
		}, nil)

	ts.mongoMock.
		EXPECT().
		GetReportedSymptoms(
			gomock.Eq(ts.testAccountNumber),
			AnyTimestamp(),
			gomock.Eq(int64(1)),
			gomock.Eq(""),
		).
		Return([]*schema.SymptomReportData{
			{
				ProfileID:     "test-account",
				AccountNumber: ts.testAccountNumber,
				Timestamp:     time.Now().Add(-24 * time.Hour).Unix(), // reported yesterday
				Symptoms: []schema.Symptom{
					schema.COVID19Symptoms[0],
					schema.COVID19Symptoms[1],
				},
			},
		}, nil)

	values, err := ts.env.ExecuteActivity(ts.worker.SymptomsNeedFollowUpActivity, ts.testAccountNumber)
	ts.NoError(err)

	symptoms := make([]schema.Symptom, 0)
	err = values.Get(&symptoms)
	ts.NoError(err)
	ts.Len(symptoms, 2)
}

func (ts *NudgeActivityTestSuite) TestCheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivityAt10am() {
	now = func() time.Time {
		t := time.Now()
		return time.Date(t.Year(), t.Month(), t.Day(), 10, 0, 0, 0, utils.GetLocation("GMT+8"))
	}

	ts.mongoMock.
		EXPECT().
		GetProfile(gomock.Eq(ts.testAccountNumber)).
		Return(&schema.Profile{
			ID:            "test-account",
			AccountNumber: ts.testAccountNumber,
			Timezone:      "GMT+8",
			LastNudge: schema.NudgeTime{
				schema.NudgeBehaviorOnSelfHighRiskSymptoms: now().Add(-5 * time.Hour), // nedged yesterday
			},
		}, nil)

	ts.mongoMock.
		EXPECT().
		GetReportedSymptoms(
			gomock.Eq(ts.testAccountNumber),
			AnyTimestamp(),
			gomock.Eq(int64(1)),
			gomock.Eq(""),
		).
		Return([]*schema.SymptomReportData{
			{
				ProfileID:     "test-account",
				AccountNumber: ts.testAccountNumber,
				Timestamp:     time.Now().Add(-24 * time.Hour).Unix(), // reported yesterday
				Symptoms: []schema.Symptom{
					schema.COVID19Symptoms[0],
					schema.COVID19Symptoms[1],
				},
			},
		}, nil)

	values, err := ts.env.ExecuteActivity(ts.worker.CheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivity, ts.testAccountNumber)
	ts.NoError(err)
	var needToFollow bool
	err = values.Get(&needToFollow)
	ts.NoError(err)
	ts.True(needToFollow)
}

func (ts *NudgeActivityTestSuite) TestCheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivityAt7am() {
	now = func() time.Time {
		t := time.Now()
		return time.Date(t.Year(), t.Month(), t.Day(), 7, 0, 0, 0, utils.GetLocation("GMT+8"))
	}

	ts.mongoMock.
		EXPECT().
		GetProfile(gomock.Eq(ts.testAccountNumber)).
		Return(&schema.Profile{
			ID:            "test-account",
			AccountNumber: ts.testAccountNumber,
			Timezone:      "GMT+8",
			LastNudge: schema.NudgeTime{
				schema.NudgeBehaviorOnSelfHighRiskSymptoms: now().Add(-5 * time.Hour), // nedged yesterday
			},
		}, nil)

	values, err := ts.env.ExecuteActivity(ts.worker.CheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivity, ts.testAccountNumber)
	ts.NoError(err)
	var needToFollow bool
	err = values.Get(&needToFollow)
	ts.NoError(err)
	ts.False(needToFollow)
}

func (ts *NudgeActivityTestSuite) TestCheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivityAt3pm() {
	now = func() time.Time {
		t := time.Now()
		return time.Date(t.Year(), t.Month(), t.Day(), 15, 0, 0, 0, utils.GetLocation("GMT+8"))
	}

	ts.mongoMock.
		EXPECT().
		GetProfile(gomock.Eq(ts.testAccountNumber)).
		Return(&schema.Profile{
			ID:            "test-account",
			AccountNumber: ts.testAccountNumber,
			Timezone:      "GMT+8",
			LastNudge: schema.NudgeTime{
				schema.NudgeBehaviorOnSelfHighRiskSymptoms: now().Add(-5 * time.Hour), // nedged yesterday
			},
		}, nil)

	ts.mongoMock.
		EXPECT().
		GetReportedSymptoms(
			gomock.Eq(ts.testAccountNumber),
			AnyTimestamp(),
			gomock.Eq(int64(1)),
			gomock.Eq(""),
		).
		Return([]*schema.SymptomReportData{
			{
				ProfileID:     "test-account",
				AccountNumber: ts.testAccountNumber,
				Timestamp:     time.Now().Add(-24 * time.Hour).Unix(), // reported yesterday
				Symptoms: []schema.Symptom{
					schema.COVID19Symptoms[0],
					schema.COVID19Symptoms[1],
				},
			},
		}, nil)

	values, err := ts.env.ExecuteActivity(ts.worker.CheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivity, ts.testAccountNumber)
	ts.NoError(err)
	var needToFollow bool
	err = values.Get(&needToFollow)
	ts.NoError(err)
	ts.True(needToFollow)
}

func (ts *NudgeActivityTestSuite) TestCheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivityWithoutSymptoms() {
	now = func() time.Time {
		t := time.Now()
		return time.Date(t.Year(), t.Month(), t.Day(), 10, 0, 0, 0, utils.GetLocation("GMT+8"))
	}

	ts.mongoMock.
		EXPECT().
		GetProfile(gomock.Eq(ts.testAccountNumber)).
		Return(&schema.Profile{
			ID:            "test-account",
			AccountNumber: ts.testAccountNumber,
			Timezone:      "GMT+8",
			LastNudge: schema.NudgeTime{
				schema.NudgeBehaviorOnSelfHighRiskSymptoms: now().Add(-5 * time.Hour), // nedged yesterday
			},
		}, nil)

	ts.mongoMock.
		EXPECT().
		GetReportedSymptoms(
			gomock.Eq(ts.testAccountNumber),
			AnyTimestamp(),
			gomock.Eq(int64(1)),
			gomock.Eq(""),
		).
		Return([]*schema.SymptomReportData{}, nil)

	values, err := ts.env.ExecuteActivity(ts.worker.CheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivity, ts.testAccountNumber)
	ts.NoError(err)
	var needToFollow bool
	err = values.Get(&needToFollow)
	ts.NoError(err)
	ts.False(needToFollow)
}

func TestNudgeActivity(t *testing.T) {
	suite.Run(t, new(NudgeActivityTestSuite))
}

func TestSymptomListingMessageNormal(t *testing.T) {
	symptoms := []schema.Symptom{
		schema.COVID19Symptoms[0],
		schema.COVID19Symptoms[1],
	}

	os.Setenv("TEST_I18N_DIR", "../../i18n")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("test")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	utils.InitI18NBundle()

	headings, contents, err := SymptomListingMessage("symptom_follow_up", symptoms)
	assert.NoError(t, err)
	assert.NotEmpty(t, headings["zh-Hant"])
	assert.NotEmpty(t, headings["en"])
	assert.NotEmpty(t, contents["zh-Hant"])
	assert.NotEmpty(t, contents["en"])
}

func TestSymptomListingMessageWithoutSymptoms(t *testing.T) {
	symptoms := []schema.Symptom{}

	os.Setenv("TEST_I18N_DIR", "../../i18n")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("test")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	utils.InitI18NBundle()

	headings, contents, err := SymptomListingMessage("symptom_follow_up", symptoms)
	assert.Errorf(t, err, "no symptoms in list")
	assert.Empty(t, headings)
	assert.Empty(t, contents)
}

func TestSymptomListingMessageWithWrongMsgType(t *testing.T) {
	symptoms := []schema.Symptom{
		schema.COVID19Symptoms[0],
	}

	os.Setenv("TEST_I18N_DIR", "../../i18n")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("test")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	utils.InitI18NBundle()

	headings, contents, err := SymptomListingMessage("incorrect-type", symptoms)
	assert.Error(t, err)
	assert.IsType(t, err, &i18n.MessageNotFoundErr{})
	assert.Empty(t, headings)
	assert.Empty(t, contents)
}
