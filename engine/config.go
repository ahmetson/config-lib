// Package engine defines a config engine for the entire app in the dev context.
//
// The config features:
//   - reads the command line arguments for the app such as authentication enabled or not.
//   - automatically loads the environment variables files.
//   - Allows setting default variables if user didn't define them.
package engine

import (
	"fmt"
	"github.com/ahmetson/datatype-lib/data_type/key_value"
	"github.com/ahmetson/os-lib/env"
	"github.com/spf13/viper"
)

// Dev context's configuration engine on viper.Viper
type Dev struct {
	*viper.Viper // used to keep default values

	// Passed as --secure command line arg.
	// If it's passed, then authentication is switched off.
	Secure       bool
	HandleChange func(interface{}, error)
}

// NewDev creates a global config for the entire application.
//
// Automatically reads the command line arguments.
// Loads the environment variables.
func NewDev() (*Dev, error) {
	// First, we load the environment variables
	err := env.LoadAnyEnv()
	if err != nil {
		return nil, fmt.Errorf("loading environment variables: %w", err)
	}

	// replace the values with the ones we fetched from environment variables
	config := Dev{
		Viper:        viper.New(),
		HandleChange: nil,
	}
	config.AutomaticEnv()

	return &config, nil
}

// YamlPathParam creates a file parameter.
func YamlPathParam(configPath string, configName string) key_value.KeyValue {
	return key_value.New().
		Set("name", configName).
		Set("configPath", configPath).
		Set("type", "yaml")
}

// SetDefaults sets the default config parameters.
func (config *Dev) SetDefaults(params key_value.KeyValue) {
	for name, value := range params {
		if value == nil {
			continue
		}
		// already set, don't use the default
		if config.IsSet(name) {
			continue
		}
		config.SetDefault(name, value)
	}
}

// Exist Checks whether the config variable exists or not
// If the config exists or its default value exists, then returns true.
func (config *Dev) Exist(name string) bool {
	return config.Viper.Get(name) != nil
}
