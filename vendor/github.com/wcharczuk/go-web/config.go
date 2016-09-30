package web

import (
	"os"
	"reflect"
)

const (
	// TagNameEnvironmentVariableName is the struct tag for what environment variable to use to populate a field.
	TagNameEnvironmentVariableName = "env"
	// TagNameEnvironmentVariableDefault is the struct tag for what to use if the environment variable is empty.
	TagNameEnvironmentVariableDefault = "env_default"
)

// MockEnvVar mocks an environment variable.
func MockEnvVar(varName, varValue string) (reset func()) {
	oldValue := os.Getenv(varName)
	os.Setenv(varName, varValue)
	return func() {
		os.Setenv(varName, oldValue)
	}
}

// Initialized is a type that can be initialized.
type Initialized interface {
	Initialize() error
}

// ReadConfigFromEnvironment reads a config from the environment.
func ReadConfigFromEnvironment(reference interface{}) (interface{}, error) {
	objectMeta := reflectType(reference)

	var field reflect.StructField
	var tag string
	var envValue string
	var defaultValue string
	var err error
	for x := 0; x < objectMeta.NumField(); x++ {
		field = objectMeta.Field(x)
		tag = field.Tag.Get(TagNameEnvironmentVariableName)

		if len(tag) > 0 {
			envValue = os.Getenv(tag)
			if len(envValue) > 0 {
				err = setValueByName(reference, field.Name, envValue)
				if err != nil {
					return reference, err
				}
			} else {
				defaultValue = field.Tag.Get(TagNameEnvironmentVariableDefault)
				if len(defaultValue) > 0 {
					err = setValueByName(reference, field.Name, defaultValue)
					if err != nil {
						return reference, err
					}
				}
			}
		}
	}

	if typed, isTyped := reference.(Initialized); isTyped {
		return typed, typed.Initialize()
	}

	return reference, nil
}
