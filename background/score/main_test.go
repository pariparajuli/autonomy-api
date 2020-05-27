package score

import (
	"os"
	"testing"

	"github.com/bitmark-inc/autonomy-api/background/nudge"
	"github.com/bitmark-inc/autonomy-api/mocks"
)

var testWorker *ScoreUpdateWorker
var mongoMock *mocks.MockMongoStore

func TestMain(m *testing.M) {
	nudge.NewNudgeWorker("test", mongoMock).Register() // register for cross worker reference
	testWorker = NewScoreUpdateWorker("test", mongoMock)
	testWorker.Register()
	os.Exit(m.Run())
}
