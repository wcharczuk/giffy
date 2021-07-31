package jobs

import (
	"context"
	"testing"

	logger "github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/testutil"
	"github.com/wcharczuk/giffy/server/config"
	"github.com/wcharczuk/giffy/server/model"
)

func TestMain(m *testing.M) {
	cfg := config.MustNewFromEnv()
	testutil.New(m,
		testutil.OptLog(logger.All()),
		testutil.OptWithDefaultDB(),
		testutil.OptBefore(
			func(ctx context.Context) error {
				return model.Schema(cfg).Apply(ctx, testutil.DefaultDB())
			},
			func(ctx context.Context) error {
				return model.Migrations(cfg).Apply(ctx, testutil.DefaultDB())
			},
		),
	).Run()
}
