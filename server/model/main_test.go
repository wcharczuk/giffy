package model

import (
	"context"
	"os"
	"testing"

	logger "github.com/blend/go-sdk/logger"
	"github.com/wcharczuk/giffy/server/core"
)

func testCtx() context.Context {
	return context.TODO()
}

func TestMain(m *testing.M) {
	if err := core.Setwd("../../"); err != nil {
		logger.All().SyncFatalExit(err)
	}
	if err := core.InitTest(); err != nil {
		logger.All().SyncFatalExit(err)
	}
	os.Exit(m.Run())
}
