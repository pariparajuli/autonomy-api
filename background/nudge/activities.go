package nudge

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bitmark-inc/autonomy-api/background"
	"github.com/bitmark-inc/autonomy-api/schema"
	"github.com/bitmark-inc/autonomy-api/utils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

const SymptomFollowUpExpiry = 24 * time.Hour

func (n *NudgeWorker) getLastSymptomReport(accountNumber string) *schema.SymptomReportData {
	symptoms, err := n.mongo.GetReportedSymptoms(accountNumber, time.Now().Unix(), 1, "")
	if err != nil {
		return nil
	}

	if len(symptoms) > 0 {
		return symptoms[0]
	}

	return nil
}

// SymptomsNeedFollowUpActivity is an activity that determine if an account contains symptoms to follow up
func (n *NudgeWorker) SymptomsNeedFollowUpActivity(ctx context.Context, accountNumber string) ([]schema.Symptom, error) {
	logger := activity.GetLogger(ctx)

	p, err := n.mongo.GetProfile(accountNumber)
	if err != nil {
		return nil, err
	}

	// get account timezone
	accountLocation := time.FixedZone("UTC+8", int((8 * time.Hour).Seconds()))

	accountNow := time.Now().In(accountLocation)
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
		lastNudgeSinceToday := p.LastSymptomNudged.Sub(accountToday.UTC())
		logger.Info("Last nudge sent since today", zap.Any("lastNudgeSinceToday", lastNudgeSinceToday))

		if lastNudgeSinceToday < 8*time.Hour { // last notified time is before this morning
			return append(report.OfficialSymptoms, report.CustomizedSymptoms...), nil
		} else if lastNudgeSinceToday >= 8*time.Hour && lastNudgeSinceToday < 12*time.Hour { // last notified time is in this morning
			if accountCurrentHour >= 13 || accountCurrentHour < 17 {
				return append(report.OfficialSymptoms, report.CustomizedSymptoms...), nil
			}
		}
	}

	logger.Info("No symptoms for following", zap.String("account_number", accountNumber))
	return nil, nil
}

// commaSeparatedSymptoms will return a string of symptoms separate by commas
func commaSeparatedSymptoms(lang string, sourceSymptoms []schema.Symptom) string {
	loc := utils.NewLocalizer(lang)

	symptomsNames := make([]string, 0)

	for _, s := range sourceSymptoms {
		if name, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: fmt.Sprintf("symptoms.%s.name", s.ID),
		}); err == nil {
			symptomsNames = append(symptomsNames, name)
		} else {
			symptomsNames = append(symptomsNames, s.Name)
		}
	}

	return strings.Join(symptomsNames, ", ")
}

// SymptomFollowUpMessage returns headings and contents in a map where its keys are languages
func (n *NudgeWorker) SymptomFollowUpMessage(sourceSymptoms []schema.Symptom) (map[string]string, map[string]string, error) {
	headings := map[string]string{}
	contents := map[string]string{}

	for key, lang := range background.OneSignalLanguageCode {
		loc := utils.NewLocalizer(lang)

		// translate heading
		heading, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: "notification.symptom_follow_up.heading",
		})
		if err != nil {
			return nil, nil, err
		}

		headings[key] = heading

		// translate content
		content, err := loc.Localize(&i18n.LocalizeConfig{
			MessageID: "notification.symptom_follow_up.content",
			TemplateData: map[string]interface{}{
				"Symptoms": commaSeparatedSymptoms(lang, sourceSymptoms),
			},
		})
		if err != nil {
			return nil, nil, err
		}

		contents[key] = content
	}

	return headings, contents, nil
}

// NotifySymptomFollowUpActivity send notifications to accounts those have symptoms to be followed
func (n *NudgeWorker) NotifySymptomFollowUpActivity(ctx context.Context, accountNumber string, symptoms []schema.Symptom) error {
	logger := activity.GetLogger(ctx)

	logger.Info("Prepare the message context for following up symptoms", zap.Any("symptoms", symptoms))

	var symptomsIDs = make([]schema.SymptomType, 0)
	for _, s := range symptoms {
		symptomsIDs = append(symptomsIDs, s.ID)
	}

	headings, contents, err := n.SymptomFollowUpMessage(symptoms)
	if err != nil {
		logger.Error("can not generate symptoms follow-up message", zap.Error(err))
	}

	if err := n.Background.NotifyAccountByText(accountNumber,
		headings, contents,
		map[string]interface{}{
			"notification_type": "ACCOUNT_SYMPTOM_FOLLOW_UP",
			"symptoms":          symptomsIDs,
		},
	); err != nil {
		return err
	}

	return n.mongo.UpdateAccountSymptomNudge(accountNumber)
}
