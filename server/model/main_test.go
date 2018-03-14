package model

import (
	"log"
	"os"
	"testing"

	"github.com/blendlabs/spiffy"
)

func TestMain(m *testing.M) {
	err := spiffy.OpenDefault(spiffy.NewFromConfig(spiffy.NewConfigFromEnv()))
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}
