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

// Load reads the config and returns it.
//
// In order to read it, call the following:
//
//	 config.Engine().SetConfigName(configName)
//		config.Engine().SetConfigType("yaml") // or json
//		config.Engine().AddConfigPath(configPath)
func (config *Dev) Load(value key_value.KeyValue) (interface{}, error) {
	name, err := value.StringValue("name")
	if err != nil {
		return nil, fmt.Errorf("value.StringValue(`name`): %w", err)
	}
	configType, err := value.StringValue("type")
	if err != nil {
		return nil, fmt.Errorf("value.StringValue(`type`): %w", err)
	}
	configPath, err := value.StringValue("configPath")
	if err != nil {
		return nil, fmt.Errorf("value.StringValue(`configPath`): %w", err)
	}
	config.SetConfigName(name)
	config.SetConfigType(configType)
	config.AddConfigPath(configPath)

	err = config.ReadInConfig()
	notFound := false
	_, notFound = err.(viper.ConfigFileNotFoundError)
	if err != nil && !notFound {
		return nil, fmt.Errorf("read '%s' failed: %w", config.GetString("SERVICE_CONFIG_NAME"), err)
	} else if notFound {
		return nil, nil
	}
	services, ok := config.Get("services").([]interface{})
	if !ok {
		return nil, fmt.Errorf("config.yml Service should be a list not a one object")
	}

	return services, nil
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
