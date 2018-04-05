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
	Label         string   `json:"label" yaml:"label" env:"LOG_LABEL"`
	OutputFormat  string   `json:"format" yaml:"format" env:"LOG_FORMAT"`
	Flags         []string `json:"flags" yaml:"flags" env:"LOG_EVENTS,csv"`
	HiddenFlags   []string `json:"hidden" yaml:"hidden" env:"LOG_HIDDEN,csv"`
	RecoverPanics *bool    `json:"recoverPanics" yaml:"recoverPanics" env:"LOG_RECOVER"`

	TextOutput TextWriterConfig `json:"textOutput" yaml:"textOutput"`
	JSONOutput JSONWriterConfig `json:"jsonOutput" yaml:"jsonOutput"`
}

// GetLabel returns the writer label.
func (c Config) GetLabel() string {
	if len(c.Label) > 0 {
		return c.Label
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
	ShowLabel  *bool  `json:"showLabel" yaml:"showLabel" env:"LOG_SHOW_LABEL"`
	ShowTime   *bool  `json:"showTime" yaml:"showTime" env:"LOG_SHOW_TIME"`
	UseColor   *bool  `json:"useColor" yaml:"useColor" env:"LOG_USE_COLOR"`
	Label      string `json:"label" yaml:"label" env:"LOG_LABEL"`
	TimeFormat string `json:"timeFormat" yaml:"timeFormat" env:"LOG_TIME_FORMAT"`
}

// GetShowLabel returns a field value or a default.
func (twc TextWriterConfig) GetShowLabel(defaults ...bool) bool {
	if twc.ShowLabel != nil {
		return *twc.ShowLabel
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return DefaultTextWriterShowLabel
}

// GetShowTime returns a field value or a default.
func (twc TextWriterConfig) GetShowTime(defaults ...bool) bool {
	if twc.ShowTime != nil {
		return *twc.ShowTime
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return DefaultTextWriterShowTime
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

// GetLabel returns a field value or a default.
func (twc TextWriterConfig) GetLabel(defaults ...string) string {
	if len(twc.Label) > 0 {
		return twc.Label
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return ""
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
	Pretty *bool `json:"pretty" yaml:"pretty" env:"LOG_JSON_PRETTY"`
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
