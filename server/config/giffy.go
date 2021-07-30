package config

import (
	"encoding/base64"
	"fmt"

	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/web"
)

const (
	// EnvironmentDev is a service environment
	EnvironmentDev = "dev"
	// EnvironmentStaging is a service environment
	EnvironmentStaging = "staging"
	// EnvironmentProd is a service environment
	EnvironmentProd = "prod"
)

// MustNewFromEnv creates a new config from the environment.
// It will panic on error.
func MustNewFromEnv() *Giffy {
	var cfg Giffy
	if err := env.Env().ReadInto(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}

// NewFromEnv creates a new config from the environment.
func NewFromEnv() (*Giffy, error) {
	var cfg Giffy
	if err := env.Env().ReadInto(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Giffy is the root config for the server.
type Giffy struct {
	Name        string `json:"name" yaml:"name" env:"SERVICE_NAME"`
	Environment string `json:"env" yaml:"env" env:"SERVICE_ENV"`

	// If this user authenticates it is automatically made a super-admin.
	AdminUserEmail string `json:"adminUserEmail" yaml:"adminUserEmail" env:"ADMIN_USER_EMAIL"`
	EncryptionKey  string `json:"encryptionKey" yaml:"encryptionKey" env:"ENCRYPTION_KEY"`

	CloudFrontDNS string `json:"cloudfrontDNS" yaml:"cloudfrontDNS" env:"CLOUDFRONT_DNS"`
	S3Bucket      string `json:"s3Bucket" yaml:"s3Bucket" env:"S3_BUCKET"`

	SlackClientID          string `json:"slackClientID" yaml:"slackClientID" env:"SLACK_CLIENT_ID"`
	SlackClientSecret      string `json:"slackClientSecret" yaml:"slackClientSecret" env:"SLACK_CLIENT_SECRET"`
	SlackVerificationToken string `json:"slackVerificationToken" yaml:"slackVerificationToken" env:"SLACK_VERIFICATION_TOKEN"`

	Aws        Aws           `json:"aws" yaml:"aws"`
	DB         db.Config     `json:"db" yaml:"db"`
	GoogleAuth oauth.Config  `json:"googleAuth" yaml:"googleAuth"`
	Logger     logger.Config `json:"logger" yaml:"logger"`
	Web        web.Config    `json:"web" yaml:"web"`
}

// Resolve resolves the config.
func (g *Giffy) Resolve() error {
	return env.Env().ReadInto(g)
}

// GetEnvironment returns a property or a default.
func (g Giffy) GetEnvironment(inherited ...string) string {
	return configutil.CoalesceString(g.Environment, EnvironmentDev, inherited...)
}

// IsProduction returns if the current env is prodlike.
func (g Giffy) IsProduction() bool {
	return g.GetEnvironment() == EnvironmentProd
}

// GetS3Bucket gets a property or a default.
func (g Giffy) GetS3Bucket(inherited ...string) string {
	return configutil.CoalesceString(g.S3Bucket, fmt.Sprintf("giffy-%s", g.GetEnvironment()), inherited...)
}

// GetCloudFrontDNS returns the cdn.
func (g Giffy) GetCloudFrontDNS(inherited ...string) string {
	return configutil.CoalesceString(g.CloudFrontDNS, "", inherited...)
}

// GetEncryptionKey gets the config encryption key as a byte blob.
func (g Giffy) GetEncryptionKey() []byte {
	if len(g.EncryptionKey) > 0 {
		key, _ := base64.StdEncoding.DecodeString(g.EncryptionKey)
		return key
	}
	return nil
}

// GetAdminUserEmail returns the admin user email.
func (g Giffy) GetAdminUserEmail(inherited ...string) string {
	return configutil.CoalesceString(g.AdminUserEmail, "will.charczuk@gmail.com", inherited...)
}
