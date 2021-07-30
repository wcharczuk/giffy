package model

import (
	"context"
	"os"
	"testing"

	logger "github.com/blend/go-sdk/logger"
)

func defaultDB() *db.Connection {
	return testutil.DefaultDB()
}

func todo() context.Context {
	return context.TODO()
}

func TestMain(m *testing.M) {
	testutil.New(m,
		testutil.OptLog(logger.All()),
		testutil.OptWithDefaultDB(),
		testutil.OptBefore(
			func(ctx context.Context) error {
				return Schema().Apply(ctx, testutil.DefaultDB())
			},
			func(ctx context.Context) error {
				return Migrations().Apply(ctx, testutil.DefaultDB())
			},
		),
	).Run()
}
