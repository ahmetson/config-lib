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
)

const (
	EnvConfigName = "CONFIG_NAME"
	EnvConfigPath = "CONFIG_PATH"
)

type App struct {
	Services []*service.Service
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
	}

	execPath, err := path.CurrentDir()
	if err != nil {
		return nil, fmt.Errorf("path.CurrentDir: %w", err)
	}

	configParam, exist, err := flagExist(execPath)
	if err != nil {
		return nil, fmt.Errorf("flagExist: %w", err)
	}

	// Load the configuration by flag parameter
	if exist {
		services, err := read(configParam, configEngine)
		if err != nil {
			return nil, fmt.Errorf("configEngine.Read: %w", err)
		}
		app.Services = services
		return app, nil
	}

	setDefault(execPath, configEngine)
	configParam, exist, err = envExist(configEngine)
	if err != nil {
		return nil, fmt.Errorf("envExist: %w", err)
	}

	// Load the configuration by environment parameter
	if exist {
		services, err := read(configParam, configEngine)
		if err != nil {
			return nil, fmt.Errorf("configEngine.Read: %w", err)
		}
		app.Services = services
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
	return engine.AppConfig(dir, fileName), true, nil
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

	return engine.AppConfig(configPath, configName), true, nil
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

func (a *App) SetService(s *service.Service) {
	a.Services = append(a.Services, s)
}

//
//// Prepare the services by validating, linting the configurations, as well as setting up the dependencies
//func Prepare(independent *service.Service) error {
//	if len(independent.Handlers) == 0 {
//		return fmt.Errorf("no Handlers. call service.AddHandler()")
//	}
//
//	// validate the service itself
//	if err := createConfiguration(independent); err != nil {
//		return fmt.Errorf("independent.createConfiguration: %w", err)
//	}
//
//	if err := fillConfiguration(independent); err != nil {
//		return fmt.Errorf("independent.fillConfiguration: %w", err)
//	}
//
//	exist, err := config.FileExist()
//	if err != nil {
//		return fmt.Errorf("config.FileExist: %w", err)
//	}
//	if !exist {
//		if err := writeConfiguration(independent); err != nil {
//			return fmt.Errorf("writeConfiguration: %w", err)
//		}
//	}
//
//	return nil
//}
//
//func createConfiguration(independent *service.Service) error {
//	independent.SocketConfig = config.Empty(independent.Id(), independent.Url(), config.IndependentType)
//
//	if err := createHandlerConfiguration(independent); err != nil {
//		return fmt.Errorf("createHandlerConfiguration: %w", err)
//	}
//
//	return nil
//}
//
//func createHandlerConfiguration(internal bool, independent *config.Service) error {
//	// validate the Handlers
//	for category, config := range independent.Handlers {
//		newConfig := handlerConfig.NewHandler(config.Type, config.Category)
//
//		sourceInstance, err := handlerConfig.NewInstance(newConfig.Category)
//		if err != nil {
//			return fmt.Errorf("service.NewInstance: %w", err)
//		}
//		newConfig.Instances = append(newConfig.Instances, *sourceInstance)
//		independent.SocketConfig.SetHandler(newConfig)
//	}
//	return nil
//}

//func fillConfiguration(independent *service.Service) error {
//	exist, err := config.FileExist()
//	if err != nil {
//		return fmt.Errorf("config.Exist: %w", err)
//	}
//	if !exist {
//		return nil
//	}
//
//	config.SetDefault(independent.ConfigEngine())
//	config.RegisterPath(independent.ConfigEngine())
//
//	serviceConfig, err := config.Read(independent.ConfigEngine())
//	if err != nil {
//		return fmt.Errorf("config.Read: %w", err)
//	}
//
//	if serviceConfig.Id != independent.SocketConfig.Id {
//		return fmt.Errorf("service type is overwritten. expected '%s', not '%s'", independent.SocketConfig.Type, serviceConfig.Type)
//	}
//
//	// validate the Handlers
//	for category, raw := range independent.Handlers {
//		c := raw.(handler.Interface)
//
//		handlerConfig, err := serviceConfig.HandlerByCategory(category)
//		if err != nil {
//			return fmt.Errorf("serviceConfig.HandlerByCategory(%s): %w", category, err)
//		}
//
//		if handlerConfig.Type != c.HandlerType() {
//			return fmt.Errorf("handler expected to be of '%s' type, not '%s'", c.HandlerType(), handlerConfig.Type)
//		}
//
//		if len(handlerConfig.Instances) == 0 {
//			return fmt.Errorf("missing %s handler instances", category)
//		}
//
//		if handlerConfig.Instances[0].Port == 0 {
//			return fmt.Errorf("the port should not be 0 in the source")
//		}
//	}
//
//	independent.SocketConfig = serviceConfig
//
//	return nil
//}
