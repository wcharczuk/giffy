package spiffy

import (
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestConfigCreateDSN(t *testing.T) {
	assert := assert.New(t)

	cfg := &Config{
		DSN:      "foo",
		Host:     "bar",
		Username: "bailey",
		Password: "dog",
		Database: "blend",
		Schema:   "mortgages",
		SSLMode:  SSLModeVerifyCA,
	}

	assert.Equal("foo", cfg.CreateDSN())

	cfg = &Config{
		Host:     "bar",
		Port:     "1234",
		Username: "bailey",
		Password: "dog",
		Database: "blend",
		Schema:   "mortgages",
		SSLMode:  SSLModeVerifyCA,
	}

	assert.Equal("postgres://bailey:dog@bar:1234/blend?sslmode=verify-ca", cfg.CreateDSN())

	cfg = &Config{
		Host:     "bar",
		Port:     "1234",
		Username: "bailey",
		Password: "dog",
		Database: "blend",
		Schema:   "mortgages",
	}

	assert.Equal("postgres://bailey:dog@bar:1234/blend?sslmode=disable", cfg.CreateDSN())

	cfg = &Config{
		Host:     "bar",
		Port:     "1234",
		Username: "bailey",
		Password: "dog",
		Database: "blend",
	}

	assert.Equal("postgres://bailey:dog@bar:1234/blend?sslmode=disable", cfg.CreateDSN())

	cfg = &Config{
		Host:     "bar",
		Port:     "1234",
		Username: "bailey",
		Password: "dog",
	}

	assert.Equal("postgres://bailey:dog@bar:1234/postgres?sslmode=disable", cfg.CreateDSN())

	cfg = &Config{
		Host:     "bar",
		Port:     "1234",
		Username: "bailey",
	}

	assert.Equal("postgres://bailey@bar:1234/postgres?sslmode=disable", cfg.CreateDSN())

	cfg = &Config{
		Host: "bar",
		Port: "1234",
	}

	assert.Equal("postgres://bar:1234/postgres?sslmode=disable", cfg.CreateDSN())

	cfg = &Config{
		Host: "bar",
	}

	assert.Equal("postgres://bar:5432/postgres?sslmode=disable", cfg.CreateDSN())

	cfg = &Config{}
	assert.Equal("postgres://localhost:5432/postgres?sslmode=disable", cfg.CreateDSN())
}
