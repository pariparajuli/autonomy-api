package nudge

import (
	"context"
	"fmt"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"

	"github.com/bitmark-inc/autonomy-api/background"
	"github.com/bitmark-inc/autonomy-api/external/onesignal"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/utils"
)

const SymptomFollowUpExpiry = 24 * time.Hour
const HighRiskSymptomExpiry = 3 * 24 * time.Hour

// now is alias of `time.Now`. `time.Now` is wildy used for checking notification intervals
// and is hard to mock it up. After adding this alias, we can easily mock the time.Now function.
// Another approach would be create a Clock interface instead.
// The alias approach is easier, but it might create some race condition during testing. Please
// make sure not run test cases in parallel
var now = time.Now

func (n *NudgeWorker) getLastSymptomReport(accountNumber string) *schema.SymptomReportData {
	symptoms, err := n.mongo.GetReportedSymptoms(accountNumber, now().Unix(), 1, "")
	if err != nil {
		return nil
	}

	if len(symptoms) > 0 {
		return symptoms[0]
	}

	return nil
}

// GetNotificationReceiverActivity returns a list of notification receiver based on
// either an account number or a poi ID
func (n *NudgeWorker) GetNotificationReceiverActivity(ctx context.Context, accountNumber, poiID string) ([]string, error) {
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
	} else {
		return nil, background.ErrBothAccountPOIEmpty
	}

	return accountNumbers, nil
}

// SymptomsNeedFollowUpActivity is an activity that determine if an account contains symptoms to follow up
func (n *NudgeWorker) SymptomsNeedFollowUpActivity(ctx context.Context, accountNumber string) ([]schema.Symptom, error) {
	logger := activity.GetLogger(ctx)

	p, err := n.mongo.GetProfile(accountNumber)
	if err != nil {
		return nil, err
	}

	// get account timezone
	accountLocation := utils.GetLocation(p.Timezone)
	if accountLocation == nil {
		accountLocation = utils.GetLocation("GMT+8")
	}

	accountNow := now().In(accountLocation)
	accountCurrentHour := accountNow.Hour()
	accountToday := time.Date(accountNow.Year(), accountNow.Month(), accountNow.Day(), 0, 0, 0, 0, accountLocation)

	report := n.getLastSymptomReport(accountNumber)

	if report == nil {
		return nil, nil
	}

	lastSymptomReportTime := time.Unix(report.Timestamp, 0)
	lastSymptomReportDuration := accountToday.UTC().Sub(lastSymptomReportTime)

	// check if the last symptom is reported yesterday.
	if lastSymptomReportDuration > 0 && lastSymptomReportDuration < SymptomFollowUpExpiry {
		lastNudgeSinceToday := p.LastNudge[schema.NudgeSymptomFollowUp].Sub(accountToday.UTC())
		logger.Info("Last nudge sent since today", zap.Any("lastNudgeSinceToday", lastNudgeSinceToday))

		if lastNudgeSinceToday < 8*time.Hour { // last notified time is before this morning
			if accountCurrentHour >= 8 && accountCurrentHour < 12 {
				logger.Info("trigger morning symptom follow up nudge", zap.Any("accountCurrentHour", accountCurrentHour))
				return report.Symptoms, nil
			}
		} else if lastNudgeSinceToday >= 8*time.Hour && lastNudgeSinceToday < 12*time.Hour { // last notified time is in this morning
			if accountCurrentHour >= 13 && accountCurrentHour < 17 {
				logger.Info("trigger afternoon symptom follow up nudge", zap.Any("accountCurrentHour", accountCurrentHour))
				return report.Symptoms, nil
			}
		}
	}

	logger.Info("No symptoms for following",
		zap.String("account_number", accountNumber),
		zap.Any("lastReportSinceToday", -lastSymptomReportDuration))
	return nil, nil
}

// SymptomListingMessage returns headings and contents in a map where its keys are languages
func SymptomListingMessage(msgType string, sourceSymptoms []schema.Symptom) (map[string]string, map[string]string, error) {
	headings := map[string]string{}
	contents := map[string]string{}

	if len(sourceSymptoms) == 0 {
		return nil, nil, fmt.Errorf("no symptoms in list")
	}

	for key, lang := range background.OneSignalLanguageCode {
		loc := utils.NewLocalizer(lang)

		// translate heading
		heading, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: fmt.Sprintf("notification.%s.heading", msgType),
		})
		if err != nil {
			return nil, nil, err
		}

		headings[key] = heading

		// translate content
		content, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: fmt.Sprintf("notification.%s.content", msgType),
			TemplateData: map[string]interface{}{
				"Symptoms": background.CommaSeparatedSymptoms(lang, sourceSymptoms),
			},
		})
		if err != nil {
			return nil, nil, err
		}

		contents[key] = content
	}

	return headings, contents, nil
}

// NotifySymptomFollowUpActivity send notifications to accounts those have symptoms to be followed [NSy_2]
func (n *NudgeWorker) NotifySymptomFollowUpActivity(ctx context.Context, accountNumber string, symptoms []schema.Symptom) error {
	logger := activity.GetLogger(ctx)

	logger.Info("Prepare the message context for following up symptoms", zap.Any("symptoms", symptoms))

	var symptomsIDs = make([]string, 0)
	for _, s := range symptoms {
		symptomsIDs = append(symptomsIDs, s.ID)
	}

	headings, contents, err := SymptomListingMessage("symptom_follow_up", symptoms)
	if err != nil {
		logger.Error("can not generate symptoms follow-up message", zap.Error(err))
		return err
	}

	if err := n.Background.NotifyAccountByText(accountNumber,
		headings, contents,
		map[string]interface{}{
			"notification_type": "ACCOUNT_SYMPTOM_FOLLOW_UP",
			"symptoms":          symptomsIDs,
		},
	); err != nil {
		if !onesignal.IsErrAllPlayersNotSubscribed(err) {
			return err
		} else {
			logger.Warn("account is not subscribed in onesignal", zap.String("accountNumber", accountNumber))
		}
	}

	return n.mongo.UpdateAccountNudge(accountNumber, schema.NudgeSymptomFollowUp)
}

// NotifySymptomSpikeActivity send notifications to accounts who have symptoms spiked around [NSy_1]
func (n *NudgeWorker) NotifySymptomSpikeActivity(ctx context.Context, accountNumber string, symptoms []schema.Symptom) error {
	logger := activity.GetLogger(ctx)

	logger.Info("Prepare the message context for following up symptoms", zap.Any("symptoms", symptoms))

	var symptomsIDs = make([]string, 0)
	for _, s := range symptoms {
		symptomsIDs = append(symptomsIDs, s.ID)
	}

	headings, contents, err := SymptomListingMessage("symptom_spike", symptoms)
	if err != nil {
		logger.Error("can not generate symptoms follow-up message", zap.Error(err))
		return err
	}

	if err := n.Background.NotifyAccountByText(accountNumber,
		headings, contents,
		map[string]interface{}{
			"notification_type": "ACCOUNT_SYMPTOM_SPIKE",
			"symptoms":          symptomsIDs,
		},
	); err != nil {
		if !onesignal.IsErrAllPlayersNotSubscribed(err) {
			return err
		} else {
			logger.Warn("account is not subscribed in onesignal", zap.String("accountNumber", accountNumber))
		}
	}

	return nil
}

// NotifySymptomSpikeActivity send notifications to accounts those have symptoms spiked around
func (n *NudgeWorker) NotifyBehaviorNudgeActivity(ctx context.Context, accountNumber string) error {
	logger := activity.GetLogger(ctx)

	headings := map[string]string{}
	contents := map[string]string{}

	for key, lang := range background.OneSignalLanguageCode {
		loc := utils.NewLocalizer(lang)

		// translate heading
		heading, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: "notification.behavior_suggestion.heading",
		})
		if err != nil {
			return err
		}

		headings[key] = heading

		// translate content
		content, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: "notification.behavior_suggestion.content",
		})
		if err != nil {
			return err
		}

		contents[key] = content
	}

	if err := n.Background.NotifyAccountByText(accountNumber,
		headings, contents,
		map[string]interface{}{
			"notification_type": "BEHAVIOR_REPORT_ON_RISK_AREA",
			"behaviors":         []schema.GoodBehaviorType{schema.CleanHand, schema.SocialDistancing, schema.WearMask},
		},
	); err != nil {
		if !onesignal.IsErrAllPlayersNotSubscribed(err) {
			return err
		} else {
			logger.Warn("account is not subscribed in onesignal", zap.String("accountNumber", accountNumber))
		}
	}
	return nil
}

// CheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivity is an activity that determine if an account contains
// high risk symptoms and need to be follow up. [NB_3-2]
// An account will be notified twice a day. Once in the morning and once in the afternoon.
// The function will return a boolean to deteremine whether to deliver a notification
func (n *NudgeWorker) CheckSelfHasHighRiskSymptomsAndNeedToFollowUpActivity(ctx context.Context, accountNumber string) (bool, error) {
	logger := activity.GetLogger(ctx)
	var shouldSendBehaviorNudge bool

	p, err := n.mongo.GetProfile(accountNumber)
	if err != nil {
		return shouldSendBehaviorNudge, err
	}

	// get account timezone
	accountLocation := utils.GetLocation(p.Timezone)
	if accountLocation == nil {
		accountLocation = utils.GetLocation("GMT+8")
	}

	accountNow := now().In(accountLocation)
	accountCurrentHour := accountNow.Hour()
	accountToday := time.Date(accountNow.Year(), accountNow.Month(), accountNow.Day(), 0, 0, 0, 0, accountLocation)

	if accountCurrentHour < 8 || accountCurrentHour >= 17 {
		return shouldSendBehaviorNudge, nil
	}

	report := n.getLastSymptomReport(accountNumber)

	if report == nil {
		return shouldSendBehaviorNudge, nil
	}

	lastSymptomReportTime := time.Unix(report.Timestamp, 0)
	lastHighRiskMoment := accountToday.Add(-1 * HighRiskSymptomExpiry)

	// check if the last symptom is reported in the past.
	officialSympoms, _ := schema.SplitSymptoms(report.Symptoms)
	if lastSymptomReportTime.Sub(lastHighRiskMoment) > 0 && len(officialSympoms) > 0 {
		lastNudgeSinceToday := p.LastNudge[schema.NudgeBehaviorOnSelfHighRiskSymptoms].Sub(accountToday.UTC())
		logger.Info("Risky symptoms found in the past",
			zap.Any("lastNudgeSinceToday", lastNudgeSinceToday),
			zap.Any("accountCurrentHour", accountCurrentHour))

		if lastNudgeSinceToday < 8*time.Hour { // not report yet in the morning
			if accountCurrentHour >= 8 && accountCurrentHour < 12 {
				logger.Info("trigger morning behavior nudge")
				shouldSendBehaviorNudge = true
			}
		} else if lastNudgeSinceToday < 12*time.Hour { // not report yet in the afternoon
			if accountCurrentHour >= 13 && accountCurrentHour < 17 {
				logger.Info("trigger afternoon behavior nudge")
				shouldSendBehaviorNudge = true
			}
		}
	} else {
		return false, background.ErrStopRenewWorkflow
	}

	return shouldSendBehaviorNudge, nil
}

// NotifyBehaviorFollowUpWhenSelfIsInHighRiskActivity is an activity to prepare and
// send a follow-up message if a user is under risk.
// The notification will be triggered either an account reported high risk symptoms or
// it enters an area where has a symptom spike detected.
// The nudge type is reflected to either kind of nudge. (BehaviorOnHighRiskNudge, BehaviorOnSymptomSpikeNudge)
func (n *NudgeWorker) NotifyBehaviorFollowUpWhenSelfIsInHighRiskActivity(ctx context.Context, accountNumber string, nudgeType schema.NudgeType) error {

	logger := activity.GetLogger(ctx)
	logger.Info("Prepare the message context for following up hige risk symptoms")

	headings := map[string]string{}
	contents := map[string]string{}

	for key, lang := range background.OneSignalLanguageCode {
		loc := utils.NewLocalizer(lang)

		// translate heading
		heading, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: "notification.behavior_high_risk_follow_up.heading",
		})
		if err != nil {
			return err
		}

		headings[key] = heading

		// translate content
		content, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: "notification.behavior_high_risk_follow_up.content",
		})
		if err != nil {
			return err
		}

		contents[key] = content
	}

	if err := n.Background.NotifyAccountByText(accountNumber,
		headings, contents,
		map[string]interface{}{
			"notification_type": "BEHAVIOR_REPORT_ON_SELF_HIGH_RISK",
			"behaviors":         []schema.GoodBehaviorType{schema.SocialDistancing, schema.WearMask},
		},
	); err != nil {
		if !onesignal.IsErrAllPlayersNotSubscribed(err) {
			return err
		} else {
			logger.Warn("account is not subscribed in onesignal", zap.String("accountNumber", accountNumber))
		}
	}

	return n.mongo.UpdateAccountNudge(accountNumber, nudgeType)

}
