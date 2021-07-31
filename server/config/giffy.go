package config

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/blend/go-sdk/configmeta"
	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/giffy/server/awsutil"
)

// MustNewFromEnv creates a new config from the environment.
// It will panic on error.
func MustNewFromEnv() *Giffy {
	cfg := new(Giffy)
	if err := cfg.Resolve(context.Background()); err != nil {
		panic(err)
	}
	return cfg
}

// NewFromEnv creates a new config from the environment.
func NewFromEnv() (*Giffy, error) {
	cfg := new(Giffy)
	if err := cfg.Resolve(context.Background()); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Giffy is the root config for the server.
type Giffy struct {
	configmeta.Meta `yaml:",inline"`

	// If this user authenticates it is automatically made a super-admin.
	AdminUserEmail string `json:"adminUserEmail" yaml:"adminUserEmail"`
	EncryptionKey  string `json:"encryptionKey" yaml:"encryptionKey"`

	CloudFrontDNS string `json:"cloudfrontDNS" yaml:"cloudfrontDNS"`
	S3Bucket      string `json:"s3Bucket" yaml:"s3Bucket"`

	SlackClientID          string `json:"slackClientID" yaml:"slackClientID"`
	SlackClientSecret      string `json:"slackClientSecret" yaml:"slackClientSecret"`
	SlackAuthReturnURL     string `json:"slackAuthReturnURL" yaml:"slackAuthReturnURL"`
	SlackVerificationToken string `json:"slackVerificationToken" yaml:"slackVerificationToken" env:"SLACK_VERIFICATION_TOKEN"`

	Aws        awsutil.Config `json:"aws" yaml:"aws"`
	DB         db.Config      `json:"db" yaml:"db"`
	GoogleAuth oauth.Config   `json:"googleAuth" yaml:"googleAuth"`
	Logger     logger.Config  `json:"logger" yaml:"logger"`
	Web        web.Config     `json:"web" yaml:"web"`
}

// Resolve resolves the config.
func (g *Giffy) Resolve(ctx context.Context) error {
	return configutil.Resolve(ctx,
		(&g.Meta).Resolve,
		(&g.DB).Resolve,
		(&g.GoogleAuth).Resolve,
		(&g.Logger).Resolve,
		(&g.Web).Resolve,

		configutil.SetString(&g.AdminUserEmail, configutil.String(g.AdminUserEmail), configutil.String("will.charczuk@gmail.com")),
		configutil.SetString(&g.EncryptionKey, configutil.Env("ENCRYPTION_KEY"), configutil.String(g.EncryptionKey)),

		configutil.SetString(&g.CloudFrontDNS, configutil.Env("CLOUDFRONT_DNS"), configutil.String(g.CloudFrontDNS)),
		configutil.SetString(&g.S3Bucket, configutil.String(g.S3Bucket), configutil.StringFunc(g.ResolveS3Bucket)),

		configutil.SetString(&g.SlackClientID, configutil.Env("SLACK_CLIENT_ID"), configutil.String(g.SlackClientID)),
		configutil.SetString(&g.SlackClientSecret, configutil.Env("SLACK_CLIENT_SECRET"), configutil.String(g.SlackClientSecret)),
		configutil.SetString(&g.SlackAuthReturnURL, configutil.Env("SLACK_AUTH_RETURN_URL"), configutil.String(g.SlackAuthReturnURL)),
		configutil.SetString(&g.SlackVerificationToken, configutil.Env("SLACK_VERIFICATION_TOKEN"), configutil.String(g.SlackVerificationToken)),
	)
}

// GetS3Bucket gets a property or a default.
func (g Giffy) ResolveS3Bucket(_ context.Context) (*string, error) {
	bucket := fmt.Sprintf("giffy-%s", g.Meta.ServiceEnvOrDefault())
	return &bucket, nil
}

// GetEncryptionKey gets the config encryption key as a byte blob.
func (g Giffy) GetEncryptionKey() []byte {
	if len(g.EncryptionKey) > 0 {
		key, _ := base64.StdEncoding.DecodeString(g.EncryptionKey)
		return key
	}
	return nil
}
