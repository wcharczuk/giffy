package model

import (
	"context"
	"testing"

	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/testutil"

	"github.com/wcharczuk/giffy/server/config"
)

func todo() context.Context {
	return context.TODO()
}

func TestMain(m *testing.M) {
	cfg := config.MustNewFromEnv()
	testutil.New(m,
		testutil.OptLog(logger.All()),
		testutil.OptWithDefaultDB(),
		testutil.OptBefore(
			func(ctx context.Context) error {
				return Schema(cfg).Apply(ctx, testutil.DefaultDB())
			},
			func(ctx context.Context) error {
				return Migrations(cfg).Apply(ctx, testutil.DefaultDB())
			},
		),
	).Run()
}
