package awsutil

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/blend/go-sdk/ref"
)

// MustNewSession returns a new session and panics on error.
func MustNewSession(cfg Config) *session.Session {
	s, err := NewSession(cfg)
	if err != nil {
		panic(err)
	}
	return s
}

// NewSession creates a new aws session from a config.
func NewSession(cfg Config) (*session.Session, error) {
	var awsConfig aws.Config
	if !cfg.IsZero() {
		awsConfig.Region = ref.String(cfg.RegionOrDefault())
		awsConfig.Credentials = credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.Token)
	} else {
		awsConfig.Region = ref.String(cfg.RegionOrDefault())
	}
	sess, err := session.NewSession(&awsConfig)
	if err != nil {
		return nil, err
	}

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}
	return sess, nil
}
