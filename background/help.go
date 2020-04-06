package background

const (
	BROADCAST_NEW_HELP   = "763b85e1-0675-4277-ae33-7ba1de47b85c"
	NOTIFY_HELP_ACCEPTED = "abf98dc0-311f-4a1b-99a0-c8d4fe1cc9cf"
	NOTIFY_HELP_EXPIRED  = "4d36ad4f-13c5-4412-8640-2d5646e8ab56"
)

// BroadcastNewHelp is a background job to send notifications to the dynamic corhort of
// a user who just creates a new help request
func (m *BackgroundManager) BroadcastNewHelp(helpID string, accountNumbers []string) error {
	return m.notifyAccountsByTemplate(accountNumbers, BROADCAST_NEW_HELP, map[string]interface{}{
		"notification_type": "BROADCAST_NEW_HELP",
		"help_id":           helpID,
	})
}

// NotifyHelpAccepted is a background job to send notification to the user of a accepted
// request help
func (m *BackgroundManager) NotifyHelpAccepted(helpID string, accountNumber string) error {
	accountNumbers := []string{accountNumber}
	return m.notifyAccountsByTemplate(accountNumbers, NOTIFY_HELP_ACCEPTED, map[string]interface{}{
		"notification_type": "NOTIFY_HELP_ACCEPTED",
		"help_id":           helpID,
	})
}

// ExpireHelpRequests is a background job to check expired accounts and send
// notification to users
func (m *BackgroundManager) ExpireHelpRequests() error {
	return m.store.ExpireHelps()
}
