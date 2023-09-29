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
	"github.com/ahmetson/os-lib/path"
	"github.com/fsnotify/fsnotify"
	"path/filepath"
	"time"

	"github.com/ahmetson/os-lib/env"
	"github.com/spf13/viper"
)

// Dev context's configuration engine on viper.Viper
type Dev struct {
	viper *viper.Viper // used to keep default values

	// Passed as --secure command line arg.
	// If it's passed, then authentication is switched off.
	Secure       bool
	handleChange func(interface{}, error)
}

// NewDev creates a global config for the entire application.
//
// Automatically reads the command line arguments.
// Loads the environment variables.
//
// Logger should be a parent
func NewDev() (*Dev, error) {
	config := Dev{
		handleChange: nil,
	}

	// First, we load the environment variables
	err := env.LoadAnyEnv()
	if err != nil {
		return nil, fmt.Errorf("loading environment variables: %w", err)
	}

	// replace the values with the ones we fetched from environment variables
	config.viper = viper.New()
	config.viper.AutomaticEnv()

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
		return nil, fmt.Errorf("value.GetString(`name`): %w", err)
	}
	configType, err := value.StringValue("type")
	if err != nil {
		return nil, fmt.Errorf("value.GetString(`type`): %w", err)
	}
	configPath, err := value.StringValue("configPath")
	if err != nil {
		return nil, fmt.Errorf("value.GetString(`configPath`): %w", err)
	}
	config.viper.SetConfigName(name)
	config.viper.SetConfigType(configType)
	config.viper.AddConfigPath(configPath)

	err = config.viper.ReadInConfig()
	notFound := false
	_, notFound = err.(viper.ConfigFileNotFoundError)
	if err != nil && !notFound {
		return nil, fmt.Errorf("read '%s' failed: %w", config.viper.GetString("SERVICE_CONFIG_NAME"), err)
	} else if notFound {
		return nil, nil
	}
	services, ok := config.viper.Get("services").([]interface{})
	if !ok {
		return nil, fmt.Errorf("config.yml Service should be a list not a one object")
	}

	return services, nil
}

func (config *Dev) getServicePath() string {
	configName := config.viper.GetString("SERVICE_CONFIG_NAME")
	configPath := config.viper.GetString("SERVICE_CONFIG_PATH")

	return filepath.Join(configPath, configName+".yml")
}

func (config *Dev) getPath() key_value.KeyValue {
	configName := config.viper.GetString("SERVICE_CONFIG_NAME")
	configPath := config.viper.GetString("SERVICE_CONFIG_PATH")
	ext := "yaml"

	return key_value.New().Set("name", configName).Set("type", ext).Set("path", configPath)
}

// Watch tracks the config change in the file.
//
// Watch could be called only once. If it's already called, then it will skip it without an error.
//
// For production, we could call config.viper.WatchRemoteConfig() for example in etcd.
func (config *Dev) Watch(watchHandle func(interface{}, error)) error {
	if config.handleChange != nil {
		return nil
	}

	servicePath := config.getServicePath()

	exists, err := path.FileExist(servicePath)
	if err != nil {
		return fmt.Errorf("FileExist('%s'): %w", servicePath, err)
	}

	// set it after checking for errors
	config.handleChange = watchHandle

	if !exists {
		// wait file appearance, then call the watchChange
		go config.watchFileCreation()
	} else {
		config.watchChange()
	}

	return nil
}

// If the file not exists, then watch for its appearance.
func (config *Dev) watchFileCreation() {
	servicePath := config.getServicePath()
	for {
		exists, err := path.FileExist(servicePath)
		if err != nil {
			config.handleChange(nil, fmt.Errorf("watchFileCreation: FileExist: %w", err))
			break
		}
		if exists {
			serviceConfig, err := config.Load(config.getPath())
			if err != nil {
				config.handleChange(nil, fmt.Errorf("watchFileCreation: config.readFile: %w", err))
				break
			}

			config.handleChange(serviceConfig, nil)

			config.watchChange()
			break
		}
		time.Sleep(time.Millisecond * 200)
	}
}

// If file exists, then watch file deletion.
func (config *Dev) watchFileDeletion() {
	servicePath := config.getServicePath()
	for {
		exists, err := path.FileExist(servicePath)
		if err != nil {
			config.handleChange(nil, fmt.Errorf("watchFileDeletion: FileExist: %w", err))
			break
		}
		if !exists {
			config.handleChange(nil, nil)

			go config.watchFileCreation()
			break
		}
		time.Sleep(time.Millisecond * 200)
	}
}

func (config *Dev) watchChange() {
	go config.watchFileDeletion()
	// if file not exists, call the file appearance

	config.viper.WatchConfig()
	config.viper.OnConfigChange(func(e fsnotify.Event) {
		newConfig, err := config.Load(config.getPath())
		if err != nil {
			config.handleChange(nil, fmt.Errorf("watchChange: config.readFile: %w", err))
		} else {
			config.handleChange(newConfig, nil)
		}
	})
}

// SetDefaults sets the default config parameters.
func (config *Dev) SetDefaults(params key_value.KeyValue) {
	for name, value := range params {
		if value == nil {
			continue
		}
		// already set, don't use the default
		if config.viper.IsSet(name) {
			continue
		}
		config.SetDefault(name, value)
	}
}

// SetDefault sets the default config name to the value
func (config *Dev) SetDefault(name string, value interface{}) {
	config.viper.SetDefault(name, value)
}

func (config *Dev) Set(name string, value interface{}) {
	config.viper.Set(name, value)
}

// Exist Checks whether the config variable exists or not
// If the config exists or its default value exists, then returns true.
func (config *Dev) Exist(name string) bool {
	value := config.viper.GetString(name)
	return len(value) > 0
}

// GetString Returns the config request as a string
func (config *Dev) StringValue(name string) string {
	value := config.viper.GetString(name)
	return value
}

// GetUint64 Returns the config request as an unsigned 64-bit number
func (config *Dev) Uint64Value(name string) uint64 {
	value := config.viper.GetUint64(name)
	return value
}

// GetBool Returns the config request as a boolean
func (config *Dev) BoolValue(name string) bool {
	value := config.viper.GetBool(name)
	return value
}
