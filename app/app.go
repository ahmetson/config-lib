// Package app is the collection of the methods to create or load the app configuration from the engine.
// The app is the collection of the services.
//
// Depends on the engine.
// Managed by the handler.
package app

import (
	"fmt"
	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/ahmetson/config-lib/engine"
	"github.com/ahmetson/config-lib/service"
	"github.com/ahmetson/os-lib/arg"
	"github.com/ahmetson/os-lib/path"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

const (
	EnvConfigName = "CONFIG_NAME"
	EnvConfigPath = "CONFIG_PATH"
)

type App struct {
	Services   []*service.Service
	fileParams key_value.KeyValue
	filePath   string
	engine     engine.Interface
}

// New App configuration.
// If possible, it loads the configuration.
//
// Loading order:
// - create a default app config
// - if app.yml exists in the root, then load it.
// - if, environment variable exists, then load it.
// - if, a flag exists, then load it.
func New(configEngine engine.Interface) (*App, error) {
	// default app is empty
	app := &App{
		Services: make([]*service.Service, 0),
		engine:   configEngine,
	}

	execPath, err := path.CurrentDir()
	if err != nil {
		return nil, fmt.Errorf("path.CurrentDir: %w", err)
	}

	flagPath, fileExist, err := flagExist(execPath)
	if err != nil {
		return nil, fmt.Errorf("flagExist: %w", err)
	}

	// Load the configuration by flag parameter
	if fileExist {
		app.fileParams = flagPath
		services, err := read(flagPath, configEngine)
		if err != nil {
			return nil, fmt.Errorf("configEngine.Read: %w", err)
		}
		app.Services = services
		return app, nil
	}

	setDefault(execPath, configEngine)
	envPath, fileExist, err := envExist(configEngine)
	if err != nil {
		return nil, fmt.Errorf("envExist: %w", err)
	}

	// Load the configuration by environment parameter
	if fileExist {
		app.fileParams = envPath
		services, err := read(envPath, configEngine)
		if err != nil {
			return nil, fmt.Errorf("configEngine.Read: %w", err)
		}
		app.Services = services
		return app, nil
	}

	// File doesn't exist, let's write it.
	// Priority is the flag path.
	// If the user didn't pass the flags, then use an environment path.
	// The environment path will not be nil, since it will use the default path.
	app.fileParams = flagPath
	if flagPath == nil {
		app.fileParams = envPath
	}
	if envPath == nil {
		return nil, fmt.Errorf("envPath is nil")
	}

	if err := app.preparePath(); err != nil {
		return nil, fmt.Errorf("app.preparePath: %w", err)
	}
	if err := app.write(); err != nil {
		return nil, fmt.Errorf("app.write: %w", err)
	}

	return app, nil
}

func read(configParam key_value.KeyValue, configEngine engine.Interface) ([]*service.Service, error) {
	raw, err := configEngine.Read(configParam)
	if err != nil {
		return nil, fmt.Errorf("configEngine.Read: %w", err)
	}

	rawServices, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("rawServices is invalid: %v", rawServices)
	}

	services := make([]*service.Service, len(rawServices))
	for i, rawService := range rawServices {
		unmarshalled, err := service.UnmarshalService(rawService)
		if err != nil {
			return nil, fmt.Errorf("config.UnMarshalService(%d): %w", i, err)
		}
		services[i] = unmarshalled
	}

	return services, nil
}

// flagExist checks is there any configuration flag.
// If the configuration flag is set, it checks does it exist in the file system.
func flagExist(execPath string) (key_value.KeyValue, bool, error) {
	if !arg.FlagExist(service.ConfigFlag) {
		return nil, false, nil
	}

	configPath := arg.FlagValue(service.ConfigFlag)

	absPath := path.AbsDir(execPath, configPath)

	exists, err := path.FileExist(absPath)
	if err != nil {
		return nil, false, fmt.Errorf("path.FileExists('%s'): %w", absPath, err)
	}

	if !exists {
		return nil, false, fmt.Errorf("file (%s) not found", absPath)
	}

	dir, fileName := path.DirAndFileName(configPath)
	return engine.YamlPathParam(dir, fileName), true, nil
}

// envExist checks is there any configuration file path from env.
// If it exists, checks does it exist in the file system.
//
// In case if it doesn't exist, it will try to load the default configuration.
func envExist(configEngine engine.Interface) (key_value.KeyValue, bool, error) {
	if !configEngine.Exist(EnvConfigName) || !configEngine.Exist(EnvConfigPath) {
		return nil, false, nil
	}

	configName := configEngine.GetString(EnvConfigName)
	configPath := configEngine.GetString(EnvConfigPath)
	absPath := path.AbsDir(configPath, configName+".yml")
	exists, err := path.FileExist(absPath)
	if err != nil {
		return nil, false, fmt.Errorf("path.FileExists('%s'): %w", absPath, err)
	}

	if !exists {
		return nil, false, nil
	}

	return engine.YamlPathParam(configPath, configName), true, nil
}

// setDefault paths of the local file to load by default
func setDefault(execPath string, engine engine.Interface) {
	engine.SetDefault(EnvConfigName, "app")
	engine.SetDefault(EnvConfigPath, execPath)
}

// Service by id returned from the app configuration.
// If not found, return nil
func (a *App) Service(id string) *service.Service {
	for _, s := range a.Services {
		if s.Id == id {
			return s
		}
	}

	return nil
}

// ServiceByUrl returns the first service of an url type.
func (a *App) ServiceByUrl(url string) *service.Service {
	for _, s := range a.Services {
		if s.Url == url {
			return s
		}
	}

	return nil
}

// SetService sets a new service into the configuration.
// After setting, the app will write it to the file.
func (a *App) SetService(s *service.Service) error {
	a.Services = append(a.Services, s)

	if err := a.write(); err != nil {
		return fmt.Errorf("app.write: %w", err)
	}

	return nil
}

func (a *App) createYaml() key_value.KeyValue {
	var services = a.Services
	kv := key_value.Empty()
	kv.Set("services", services)

	return kv
}

func (a *App) preparePath() error {
	if a.fileParams == nil {
		return fmt.Errorf("a.fileParams nil")
	}
	name, err := a.fileParams.GetString("name")
	if err != nil {
		return fmt.Errorf("a.fileParams.GetString('name'): %w", err)
	}
	dirPath, err := a.fileParams.GetString("configPath")
	if err != nil {
		return fmt.Errorf("a.fileParams.GetString('configPath'): %w", err)
	}

	dirExist, err := path.DirExist(dirPath)
	if err != nil {
		return fmt.Errorf("path.DirExist('%s'): %w", dirPath, err)
	}
	if !dirExist {
		err = path.MakeDir(dirPath)
		if err != nil {
			return fmt.Errorf("path.MakeDir('%s'): %w", dirPath, err)
		}
	}
	a.filePath = filepath.Join(dirPath, name+".yml")

	return nil
}

// Writes the service as the yaml on the given path.
// If the path doesn't contain the file extension, it will through an error
func (a *App) write() error {
	kv := a.createYaml()

	appConfig, err := yaml.Marshal(kv.Map())
	if err != nil {
		return fmt.Errorf("yaml.Marshal: %w", err)
	}
	a.engine.Set("services", a.Services)

	f, _ := os.OpenFile(a.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	_, err = f.Write(appConfig)
	closeErr := f.Close()
	if closeErr != nil {
		if err != nil {
			return fmt.Errorf("%v: file.Close: %w", err, closeErr)
		} else {
			return fmt.Errorf("file.Close: %w", closeErr)
		}
	} else if err != nil {
		return fmt.Errorf("file.Write: %w", err)
	}

	return nil
}
