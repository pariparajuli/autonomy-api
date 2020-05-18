package background

import (
	"context"

	"github.com/bitmark-inc/autonomy-api/external/onesignal"
	"github.com/spf13/viper"
)

// OneSignalLanguageCode is a mapping between onesignal language code and i18n language code
var OneSignalLanguageCode = map[string]string{
	"zh-Hant": "zh_tw",
	"en":      "en",
}

// notifyAccountsByTemplate will consolidate account numbers and submit notification requests
func (b *Background) NotifyAccountsByTemplate(accountNumbers []string, templateID string, data map[string]interface{}) error {
	filters := []map[string]string{}
	for i, a := range accountNumbers {
		if i%100 == 0 {
			filters = append(filters, map[string]string{
				"field":    "tag",
				"key":      "account_number",
				"relation": "=",
				"value":    a,
			})
		} else {
			filters = append(filters,
				map[string]string{"operator": "OR"},
				map[string]string{
					"field":    "tag",
					"key":      "account_number",
					"relation": "=",
					"value":    a,
				})
		}
		if i%100 == 99 {
			req := &onesignal.NotificationRequest{
				AppID:          viper.GetString("onesignal.appid"),
				TemplateID:     templateID,
				Filters:        filters,
				Data:           data,
				LocalChannelID: "important_alert",
			}
			if err := b.Onesignal.SendNotification(context.Background(), req); err != nil {
				return err
			}
			filters = []map[string]string{}
		}
	}
	// send rest of notification
	req := &onesignal.NotificationRequest{
		AppID:          viper.GetString("onesignal.appid"),
		TemplateID:     templateID,
		Filters:        filters,
		Data:           data,
		LocalChannelID: "important_alert",
	}
	return b.Onesignal.SendNotification(context.Background(), req)
}

// NotifyAccountByText will send message to an account by raw headings, contents and data
func (b *Background) NotifyAccountByText(accountNumber string, headings, contents map[string]string, data map[string]interface{}) error {
	filters := []map[string]string{
		{
			"field":    "tag",
			"key":      "account_number",
			"relation": "=",
			"value":    accountNumber,
		},
	}

	// send rest of notification
	req := &onesignal.NotificationRequest{
		AppID:          viper.GetString("onesignal.appid"),
		Headings:       headings,
		Contents:       contents,
		Filters:        filters,
		Data:           data,
		LocalChannelID: "important_alert",
	}
	return b.Onesignal.SendNotification(context.Background(), req)
}
