package config

import (
	"fmt"

	logger "github.com/blendlabs/go-logger"
	util "github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/env"
	web "github.com/blendlabs/go-web"
	"github.com/blendlabs/spiffy"
)

const (
	// EnvironmentDev is a service environment
	EnvironmentDev = "dev"
	// EnvironmentStaging is a service environment
	EnvironmentStaging = "staging"
	// EnvironmentProd is a service environment
	EnvironmentProd = "prod"
)

// NewFromEnv creates a new config from the environment.
func NewFromEnv() *Giffy {
	var cfg Giffy
	env.Env().ReadInto(&cfg)
	return &cfg
}

// Giffy is the root config for the server.
type Giffy struct {
	Name        string `json:"name" yaml:"name" env:"SERVICE_NAME"`
	Environment string `json:"envionment" yaml:"environment" env:"SERVICE_ENV"`

	// If this user authenticates it is automatically made a super-admin.
	AdminUserEmail string `json:"adminUserEmail" yaml:"adminUserEmail"`
	EncryptionKey  string `json:"encryptionKey" yaml:"encryptionKey"`

	S3Bucket string `json:"s3Bucket" yaml:"s3Bucket"`

	FacebookClientID     string `json:"facebookClientID" yaml:"facebookClientID"`
	FacebookClientSecret string `json:"facebookClientSecret" yaml:"facebookClientSecret"`

	GoogleClientID     string `json:"googleClientID" yaml:"googleClientID"`
	GoogleClientSecret string `json:"googleClientSecret" yaml:"googleClientSecret"`

	SlackClientID          string `json:"slackClientID" yaml:"slackClientID"`
	SlackClientSecret      string `json:"slackClientSecret" yaml:"slackClientSecret"`
	SlackVerificationToken string `json:"slackVerificationToken" yaml:"slackVerificationToken"`

	Logger logger.Config `json:"logger" yaml:"logger"`
	Web    web.Config    `json:"web" yaml:"web"`
	DB     spiffy.Config `json:"db" yaml:"db"`
	Aws    Aws           `json:"aws" yaml:"aws"`
}

// GetEnvironment returns a property or a default.
func (g Giffy) GetEnvironment(inherited ...string) string {
	return util.Coalesce.String(g.Environment, EnvironmentDev, inherited...)
}

// IsProduction returns if the current env is prodlike.
func (g Giffy) IsProduction() bool {
	return g.GetEnvironment() == EnvironmentProd
}

// GetS3Bucket gets a property or a default.
func (g Giffy) GetS3Bucket(inherited ...string) string {
	return util.Coalesce.String(g.S3Bucket, fmt.Sprintf("giffy-%s", g.GetEnvironment()), inherited...)
}

// GetEncryptionKey gets the config encryption key as a byte blob.
func (g Giffy) GetEncryptionKey() []byte {
	if len(g.EncryptionKey) > 0 {
		key, _ := util.Base64.Decode(g.EncryptionKey)
		return key
	}
	return nil
}
