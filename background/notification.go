package background

import (
	"context"

	"github.com/bitmark-inc/autonomy-api/external/onesignal"
)

type NotificationCenter interface {
	NotifyAccountByText(accountNumber string, headings, contents map[string]string, data map[string]interface{}) error
	NotifyAccountsByTemplate(accountNumbers []string, templateID string, data map[string]interface{}) error
}

type OnesignalNotificationCenter struct {
	appID  string
	client *onesignal.OneSignalClient
}

func NewOnesignalNotificationCenter(appID string, client *onesignal.OneSignalClient) *OnesignalNotificationCenter {
	return &OnesignalNotificationCenter{
		appID:  appID,
		client: client,
	}
}

func (o *OnesignalNotificationCenter) NotifyAccountByText(accountNumber string, headings, contents map[string]string, data map[string]interface{}) error {
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
		AppID:          o.appID,
		Headings:       headings,
		Contents:       contents,
		Filters:        filters,
		Data:           data,
		LocalChannelID: "important_alert",
	}
	return o.client.SendNotification(context.Background(), req)
}
func (o *OnesignalNotificationCenter) NotifyAccountsByTemplate(accountNumbers []string, templateID string, data map[string]interface{}) error {
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
				AppID:          o.appID,
				TemplateID:     templateID,
				Filters:        filters,
				Data:           data,
				LocalChannelID: "important_alert",
			}
			if err := o.client.SendNotification(context.Background(), req); err != nil {
				return err
			}
			filters = []map[string]string{}
		}
	}
	// send rest of notification
	req := &onesignal.NotificationRequest{
		AppID:          o.appID,
		TemplateID:     templateID,
		Filters:        filters,
		Data:           data,
		LocalChannelID: "important_alert",
	}
	return o.client.SendNotification(context.Background(), req)
}
