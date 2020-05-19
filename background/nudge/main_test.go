package nudge

import (
	"os"
	"testing"

	"github.com/bitmark-inc/autonomy-api/mocks"
)

var nudgeWorker *NudgeWorker
var mongoMock *mocks.MockMongoStore

func TestMain(m *testing.M) {
	nudgeWorker = NewNudgeWorker("test", mongoMock)
	nudgeWorker.Register()
	os.Exit(m.Run())
}
