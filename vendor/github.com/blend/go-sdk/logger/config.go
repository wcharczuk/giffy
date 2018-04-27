package logger

import (
	"strings"

	"github.com/blend/go-sdk/env"
)

var (
	// DefaultFlags are the default flags.
	DefaultFlags = []Flag{Fatal, Error, Warning, Info, WebRequest}
	// DefaultFlagSet is the default verbosity for a diagnostics agent inited from the environment.
	DefaultFlagSet = NewFlagSet(DefaultFlags...)

	// DefaultHiddenFlags are the default hidden flags.
	DefaultHiddenFlags []Flag
)

// NewConfigFromEnv returns a new config from the environment.
func NewConfigFromEnv() *Config {
	var config Config
	env.Env().ReadInto(&config)
	return &config
}

// Config is the logger config.
type Config struct {
	Heading       string   `json:"heading,omitempty" yaml:"heading,omitempty" env:"LOG_HEADING"`
	OutputFormat  string   `json:"format,omitempty" yaml:"format,omitempty" env:"LOG_FORMAT"`
	Flags         []string `json:"flags,omitempty" yaml:"flags,omitempty" env:"LOG_EVENTS,csv"`
	HiddenFlags   []string `json:"hidden,omitempty" yaml:"hidden,omitempty" env:"LOG_HIDDEN,csv"`
	RecoverPanics *bool    `json:"recoverPanics,omitempty" yaml:"recoverPanics,omitempty" env:"LOG_RECOVER"`

	TextOutput TextWriterConfig `json:"textOutput,omitempty" yaml:"textOutput,omitempty"`
	JSONOutput JSONWriterConfig `json:"jsonOutput,omitempty" yaml:"jsonOutput,omitempty"`
}

// GetHeading returns the writer heading.
func (c Config) GetHeading() string {
	if len(c.Heading) > 0 {
		return c.Heading
	}
	return ""
}

// GetOutputFormat returns the output format.
func (c Config) GetOutputFormat() OutputFormat {
	if len(c.OutputFormat) > 0 {
		return OutputFormat(strings.ToLower(c.OutputFormat))
	}
	return OutputFormatText
}

// GetFlags returns the enabled logger events.
func (c Config) GetFlags() []string {
	if len(c.Flags) > 0 {
		return c.Flags
	}
	return AsStrings(DefaultFlags...)
}

// GetHiddenFlags returns the enabled logger events.
func (c Config) GetHiddenFlags() []string {
	if len(c.HiddenFlags) > 0 {
		return c.HiddenFlags
	}
	return AsStrings(DefaultHiddenFlags...)
}

// GetRecoverPanics returns a field value or a default.
func (c Config) GetRecoverPanics(defaults ...bool) bool {
	if c.RecoverPanics != nil {
		return *c.RecoverPanics
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return DefaultRecoverPanics
}

// GetWriters returns the configured writers
func (c Config) GetWriters() []Writer {
	switch c.GetOutputFormat() {
	case OutputFormatJSON:
		return []Writer{NewJSONWriterFromConfig(&c.JSONOutput)}
	case OutputFormatText:
		return []Writer{NewTextWriterFromConfig(&c.TextOutput)}
	default:
		return []Writer{NewTextWriterFromConfig(&c.TextOutput)}
	}
}

// NewTextWriterConfigFromEnv returns a new text writer config from the environment.
func NewTextWriterConfigFromEnv() *TextWriterConfig {
	var config TextWriterConfig
	env.Env().ReadInto(&config)
	return &config
}

// TextWriterConfig is the config for a text writer.
type TextWriterConfig struct {
	ShowHeadings  *bool  `json:"showHeadings,omitempty" yaml:"showHeadings,omitempty" env:"LOG_SHOW_HEADINGS"`
	ShowTimestamp *bool  `json:"showTimestamp,omitempty" yaml:"showTimestamp,omitempty" env:"LOG_SHOW_TIMESTAMP"`
	UseColor      *bool  `json:"useColor,omitempty" yaml:"useColor,omitempty" env:"LOG_USE_COLOR"`
	TimeFormat    string `json:"timeFormat,omitempty" yaml:"timeFormat,omitempty" env:"LOG_TIME_FORMAT"`
}

// GetShowHeadings returns a field value or a default.
func (twc TextWriterConfig) GetShowHeadings(defaults ...bool) bool {
	if twc.ShowHeadings != nil {
		return *twc.ShowHeadings
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return DefaultTextWriterShowHeadings
}

// GetShowTimestamp returns a field value or a default.
func (twc TextWriterConfig) GetShowTimestamp(defaults ...bool) bool {
	if twc.ShowTimestamp != nil {
		return *twc.ShowTimestamp
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return DefaultTextWriterShowTimestamp
}

// GetUseColor returns a field value or a default.
func (twc TextWriterConfig) GetUseColor(defaults ...bool) bool {
	if twc.UseColor != nil {
		return *twc.UseColor
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return DefaultTextWriterUseColor
}

// GetTimeFormat returns a field value or a default.
func (twc TextWriterConfig) GetTimeFormat(defaults ...string) string {
	if len(twc.TimeFormat) > 0 {
		return twc.TimeFormat
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return DefaultTextTimeFormat
}

// NewJSONWriterConfigFromEnv returns a new json writer config from the environment.
func NewJSONWriterConfigFromEnv() *JSONWriterConfig {
	var config JSONWriterConfig
	env.Env().ReadInto(&config)
	return &config
}

// JSONWriterConfig is the config for a json writer.
type JSONWriterConfig struct {
	Pretty *bool `json:"pretty,omitempty" yaml:"pretty,omitempty" env:"LOG_JSON_PRETTY"`
}

// GetPretty returns a field value or a default.
func (jwc JSONWriterConfig) GetPretty(defaults ...bool) bool {
	if jwc.Pretty != nil {
		return *jwc.Pretty
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return DefaultJSONWriterPretty
}
