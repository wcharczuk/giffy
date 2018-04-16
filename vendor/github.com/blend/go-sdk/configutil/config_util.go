package configutil

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/yaml"
)

const (
	// EnvVarConfigPath is the env var for configs.
	EnvVarConfigPath = "CONFIG_PATH"

	// ExtensionJSON is a file extension.
	ExtensionJSON = ".json"
	// ExtensionYAML is a file extension.
	ExtensionYAML = ".yaml"
	// ExtensionYML is a file extension.
	ExtensionYML = ".yml"

	// ErrConfigPathUnset is a common error.
	ErrConfigPathUnset Error = "config path unset"

	// ErrInvalidConfigExtension is a common error.
	ErrInvalidConfigExtension Error = "config extension invalid"
)

// Path returns the config path.
func Path(defaults ...string) string {
	if env.Env().Has(EnvVarConfigPath) {
		return env.Env().String(EnvVarConfigPath)
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return ""
}

// Deserialize deserializes a config.
func Deserialize(ext string, r io.Reader, ref Any) error {
	switch strings.ToLower(ext) {
	case ExtensionJSON:
		return exception.Wrap(json.NewDecoder(r).Decode(ref))
	case ExtensionYAML, ExtensionYML:
		contents, err := ioutil.ReadAll(r)
		if err != nil {
			return exception.Wrap(err)
		}
		return exception.Wrap(yaml.Unmarshal(contents, ref))
	default:
		return exception.NewFromErr(ErrInvalidConfigExtension).WithMessagef("extension: %s", ext)
	}
}

// Read reads a config from a default path (or inferred path from the environment).
func Read(ref Any, defaultPath ...string) error {
	return ReadFromPath(ref, Path(defaultPath...))
}

// ReadFromPath reads a config from a given path.
func ReadFromPath(ref Any, path string) error {
	defer env.Env().ReadInto(ref)

	if len(path) == 0 {
		return exception.NewFromErr(ErrConfigPathUnset)
	}

	f, err := os.Open(path)
	if err != nil {
		return exception.Wrap(err)
	}
	defer f.Close()

	return Deserialize(filepath.Ext(path), f, ref)
}

// ReadFromReader reads a config from a given reader.
func ReadFromReader(ref Any, r io.Reader, ext string) error {
	defer env.Env().ReadInto(ref)
	return Deserialize(ext, r, ref)
}

// IsIgnored returns if we should ignore the config read error.
func IsIgnored(err error) bool {
	if err == nil {
		return true
	}
	return !IsNotExist(err) && !IsConfigPathUnset(err) && !IsInvalidConfigExtension(err)
}

// IsNotExist returns if an error is an os.ErrNotExist.
func IsNotExist(err error) bool {
	if typed, isTyped := err.(exception.Exception); isTyped {
		err = typed.Inner()
	}
	return os.IsNotExist(err)
}

// IsConfigPathUnset returns if an error is an ErrConfigPathUnset.
func IsConfigPathUnset(err error) bool {
	if typed, isTyped := err.(*exception.Ex); isTyped {
		err = typed.Inner()
	}
	return err == ErrConfigPathUnset
}

// IsInvalidConfigExtension returns if an error is an ErrInvalidConfigExtension.
func IsInvalidConfigExtension(err error) bool {
	if typed, isTyped := err.(*exception.Ex); isTyped {
		err = typed.Inner()
	}
	return err == ErrInvalidConfigExtension
}
