// Package watch listens for the live update of the data.
//
// Todo improve the function call cycle (fileCreated -> fileChange -> fileDelete)
package watch

import (
	"fmt"
	"github.com/ahmetson/config-lib/engine"
	"github.com/ahmetson/datatype-lib/data_type/key_value"
	"github.com/ahmetson/os-lib/path"
	"github.com/fsnotify/fsnotify"
	"path/filepath"
	"time"
)

// Watch tracks the config change in the file.
//
// Watch could be called only once. If it's already called, then it will skip it without an error.
//
// For production, we could call config.viper.WatchRemoteConfig() for example in etcd.
func Watch(config *engine.Dev, watchHandle func(interface{}, error)) error {
	if config.HandleChange != nil {
		return nil
	}

	servicePath := getServicePath(config)

	exists, err := path.FileExist(servicePath)
	if err != nil {
		return fmt.Errorf("FileExist('%s'): %w", servicePath, err)
	}

	// set it after checking for errors
	config.HandleChange = watchHandle

	if !exists {
		// wait file appearance, then call the watchChange
		go watchFileCreation(config)
	} else {
		watchChange(config)
	}

	return nil
}

// If the file not exists, then watch for its appearance.
func watchFileCreation(config *engine.Dev) {
	servicePath := getServicePath(config)
	for {
		exists, err := path.FileExist(servicePath)
		if err != nil {
			config.HandleChange(nil, fmt.Errorf("watchFileCreation: FileExist: %w", err))
			break
		}
		if exists {
			serviceConfig, err := config.Load(getPath(config))
			if err != nil {
				config.HandleChange(nil, fmt.Errorf("watchFileCreation: config.readFile: %w", err))
				break
			}

			config.HandleChange(serviceConfig, nil)

			watchChange(config)
			break
		}
		time.Sleep(time.Millisecond * 200)
	}
}

// If file exists, then watch file deletion.
func watchFileDeletion(config *engine.Dev) {
	servicePath := getServicePath(config)
	for {
		exists, err := path.FileExist(servicePath)
		if err != nil {
			config.HandleChange(nil, fmt.Errorf("watchFileDeletion: FileExist: %w", err))
			break
		}
		if !exists {
			config.HandleChange(nil, nil)

			go watchFileCreation(config)
			break
		}
		time.Sleep(time.Millisecond * 200)
	}
}

func watchChange(config *engine.Dev) {
	go watchFileDeletion(config)
	// if file not exists, call the file appearance

	config.Viper.WatchConfig()
	config.Viper.OnConfigChange(func(e fsnotify.Event) {
		newConfig, err := config.Load(getPath(config))
		if err != nil {
			config.HandleChange(nil, fmt.Errorf("watchChange: config.readFile: %w", err))
		} else {
			config.HandleChange(newConfig, nil)
		}
	})
}

func getServicePath(config *engine.Dev) string {
	configName := config.GetString("SERVICE_CONFIG_NAME")
	configPath := config.GetString("SERVICE_CONFIG_PATH")

	return filepath.Join(configPath, configName+".yml")
}

func getPath(config *engine.Dev) key_value.KeyValue {
	configName := config.GetString("SERVICE_CONFIG_NAME")
	configPath := config.GetString("SERVICE_CONFIG_PATH")
	ext := "yaml"

	return key_value.New().Set("name", configName).Set("type", ext).Set("path", configPath)
}
