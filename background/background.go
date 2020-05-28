package background

import (
	"github.com/bitmark-inc/autonomy-api/external/onesignal"
)

// Background is a struct to maintain common clients
// and functions for all background workers
type Background struct {
	Onesignal *onesignal.OneSignalClient
}
