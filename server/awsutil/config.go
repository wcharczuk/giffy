package awsutil

import (
	"context"

	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/env"
)

// Assert Config implements configutil.Resolver
var (
	_ configutil.Resolver = (*Config)(nil)
)

const (
	// EnvVarAWSRegion is an environment variable.
	EnvVarAWSRegion = "AWS_REGION"
	// DefaultRegion a default.
	DefaultRegion = "us-east-1"
)

// Config is a config object.
type Config struct {
	Region          string `json:"region,omitempty" yaml:"region,omitempty" env:"AWS_REGION_NAME"`
	AccessKeyID     string `json:"accessKeyID,omitempty" yaml:"accessKeyID,omitempty" env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" yaml:"secretAccessKey,omitempty" env:"AWS_SECRET_ACCESS_KEY"`
	Token           string `json:"token,omitempty" yaml:"token,omitempty" env:"AWS_SECURITY_TOKEN"`
}

// Resolve implements configutil.Resolver.
func (a *Config) Resolve(ctx context.Context) error {
	return env.GetVars(ctx).ReadInto(a)
}

// IsZero returns if the aws config is set.
func (a Config) IsZero() bool {
	return a.AccessKeyID == "" || a.SecretAccessKey == ""
}

// RegionOrDefault gets a property or a default.
func (a Config) RegionOrDefault() string {
	if a.Region != "" {
		return a.Region
	}
	return DefaultRegion
}
