package config

import (
	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/env"
)

const (
	// DefaultAWSRegion is a default.
	DefaultAWSRegion = "us-east-1"
)

// NewAwsFromEnv returns a new aws config from the environment.
func NewAwsFromEnv() *Aws {
	var aws Aws
	env.Env().ReadInto(&aws)
	return &aws
}

// Aws is a config object.
type Aws struct {
	Region          string `json:"region,omitempty" yaml:"region,omitempty" env:"AWS_REGION"`
	AccessKeyID     string `json:"accessKeyID,omitempty" yaml:"accessKeyID,omitempty" env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" yaml:"secretAccessKey,omitempty" env:"AWS_SECRET_ACCESS_KEY"`
	SecurityToken   string `json:"securityToken" yaml:"securityToken" env:"AWS_SECURITY_TOKEN"`
}

// GetRegion gets a property or a default.
func (a Aws) GetRegion(defaults ...string) string {
	return configutil.CoalesceString(a.Region, DefaultAWSRegion, defaults...)
}

// GetAccessKeyID gets a property or a default.
func (a Aws) GetAccessKeyID(defaults ...string) string {
	return configutil.CoalesceString(a.AccessKeyID, "", defaults...)
}

// GetSecretAccessKey gets a property or a default.
func (a Aws) GetSecretAccessKey(defaults ...string) string {
	return configutil.CoalesceString(a.SecretAccessKey, "", defaults...)
}

// GetSecurityToken gets a property or a default.
func (a Aws) GetSecurityToken(defaults ...string) string {
	return configutil.CoalesceString(a.SecurityToken, "", defaults...)
}
