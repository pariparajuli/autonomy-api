package background

import (
	"context"

	"github.com/bitmark-inc/autonomy-api/external/onesignal"
	"github.com/spf13/viper"
)

// notifyAccountsByTemplate will consolidate account numbers and submit notification requests
func (m *BackgroundManager) notifyAccountsByTemplate(accountNumbers []string, templateID string, data map[string]interface{}) error {
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
				AppID:      viper.GetString("onesignal.appid"),
				TemplateID: templateID,
				Filters:    filters,
				Data:       data,
			}
			if err := m.onesignal.SendNotification(context.Background(), req); err != nil {
				return err
			}
			filters = []map[string]string{}
		}
	}
	// send rest of notification
	req := &onesignal.NotificationRequest{
		AppID:      viper.GetString("onesignal.appid"),
		TemplateID: templateID,
		Filters:    filters,
		Data:       data,
	}
	return m.onesignal.SendNotification(context.Background(), req)
}
