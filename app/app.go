// Package app is the collection of the methods to create or load the app configuration from the engine.
// The app is the collection of the services.
//
// Depends on the engine.
// Managed by the handler.
package app

import (
	"fmt"
	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/ahmetson/config-lib"
	"github.com/ahmetson/config-lib/engine"
	"github.com/ahmetson/os-lib/arg"
	"github.com/ahmetson/os-lib/path"
)

const (
	EnvConfigName = "CONFIG_NAME"
	EnvConfigPath = "CONFIG_PATH"
)

type App struct {
	Services []*config.Service
	Id       string
	ParentId string
	Url      string
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
		Services: make([]*config.Service, 0),
	}

	if arg.FlagExist(config.IdFlag) {
		app.Id = arg.FlagValue(config.IdFlag)
	}

	if arg.FlagExist(config.UrlFlag) {
		app.Url = arg.FlagValue(config.UrlFlag)
	}

	if arg.FlagExist(config.ParentFlag) {
		app.ParentId = arg.FlagValue(config.ParentFlag)
	}

	execPath, err := path.CurrentDir()
	if err != nil {
		return nil, fmt.Errorf("path.CurrentDir: %w", err)
	}

	configParam, exist, err := flagExist(execPath)
	if err != nil {
		return nil, fmt.Errorf("flagExist: %w", err)
	}

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

	if exist {
		services, err := read(configParam, configEngine)
		if err != nil {
			return nil, fmt.Errorf("configEngine.Read: %w", err)
		}
		app.Services = services
	}

	return app, nil
}

func read(configParam key_value.KeyValue, configEngine engine.Interface) ([]*config.Service, error) {
	raw, err := configEngine.Read(configParam)
	if err != nil {
		return nil, fmt.Errorf("configEngine.Read: %w", err)
	}

	rawServices, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("rawServices is invalid: %v", rawServices)
	}

	services := make([]*config.Service, len(rawServices))
	for i, rawService := range rawServices {
		unmarshalled, err := config.UnmarshalService(rawService)
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
	if !arg.FlagExist(config.ConfigFlag) {
		return nil, false, nil
	}

	configPath := arg.FlagValue(config.ConfigFlag)

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
func (a *App) Service(id string) *config.Service {
	for _, s := range a.Services {
		if s.Id == id {
			return s
		}
	}

	return nil
}

// ServiceByUrl returns the first service of an url type.
func (a *App) ServiceByUrl(url string) *config.Service {
	for _, s := range a.Services {
		if s.Url == url {
			return s
		}
	}

	return nil
}

func (a *App) SetService(s *config.Service) {
	a.Services = append(a.Services, s)
}

//
//// Prepare the services by validating, linting the configurations, as well as setting up the dependencies
//func Prepare(independent *service.Service) error {
//	if len(independent.Controllers) == 0 {
//		return fmt.Errorf("no Controllers. call service.AddController")
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
//	independent.Config = config.Empty(independent.Id(), independent.Url(), config.IndependentType)
//
//	if err := createHandlerConfiguration(independent); err != nil {
//		return fmt.Errorf("createHandlerConfiguration: %w", err)
//	}
//
//	return nil
//}
//
//func createHandlerConfiguration(independent *service.Service) error {
//	// validate the Controllers
//	for category, controllerInterface := range independent.Controllers {
//		c := controllerInterface.(handler.Interface)
//
//		controllerConfig := handlerConfig.NewController(c.ControllerType(), category)
//
//		sourceInstance, err := handlerConfig.NewInstance(controllerConfig.Category)
//		if err != nil {
//			return fmt.Errorf("service.NewInstance: %w", err)
//		}
//		controllerConfig.Instances = append(controllerConfig.Instances, *sourceInstance)
//		independent.Config.SetController(controllerConfig)
//	}
//	return nil
//}
//
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
//	if serviceConfig.Id != independent.Config.Id {
//		return fmt.Errorf("service type is overwritten. expected '%s', not '%s'", independent.Config.Type, serviceConfig.Type)
//	}
//
//	// validate the Controllers
//	for category, controllerInterface := range independent.Controllers {
//		c := controllerInterface.(handler.Interface)
//
//		controllerConfig, err := serviceConfig.GetController(category)
//		if err != nil {
//			return fmt.Errorf("serviceConfig.GetController(%s): %w", category, err)
//		}
//
//		if controllerConfig.Type != c.ControllerType() {
//			return fmt.Errorf("handler expected to be of '%s' type, not '%s'", c.ControllerType(), controllerConfig.Type)
//		}
//
//		if len(controllerConfig.Instances) == 0 {
//			return fmt.Errorf("missing %s handler instances", category)
//		}
//
//		if controllerConfig.Instances[0].Port == 0 {
//			return fmt.Errorf("the port should not be 0 in the source")
//		}
//	}
//
//	independent.Config = serviceConfig
//
//	return nil
//}
