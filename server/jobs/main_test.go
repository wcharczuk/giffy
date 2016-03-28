package jobs

import (
	"os"
	"testing"

	"github.com/wcharczuk/giffy/server/core"
)

func TestMain(m *testing.M) {
	core.DBInit()
	os.Exit(m.Run())
}
