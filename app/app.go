// Package app is the collection of the methods to create or load the app configuration from the engine.
// The app is the collection of the services.
//
// Depends on the engine.
// Managed by the handler.
package app

import (
	"fmt"
	"github.com/ahmetson/config-lib/engine"
	"github.com/ahmetson/config-lib/service"
	"github.com/ahmetson/datatype-lib/data_type/key_value"
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

// App is the configuration of the entire application.
// Consists the supported services and proxy chains.
//
// Fields
//   - Services in the application
//   - ProxyChains list of proxies that targets to the services
type App struct {
	Services    []*service.Service    `json:"services" yaml:"services"`
	ProxyChains []*service.ProxyChain `json:"proxy_chains" yaml:"proxy_chains"`
	filePath    string
}

// New App configuration.
// If possible, it loads the configuration.
//
// Loading order:
// - create a default app config
// - if app.yml exists in the root, then load it.
// - if, environment variable exists, then load it.
// - if, a flag exists, then load it.
func New(configEngine *engine.Dev) (*App, error) {
	// default app is empty
	execPath, err := path.CurrentDir()
	if err != nil {
		return nil, fmt.Errorf("path.CurrentDir: %w", err)
	}

	flagFileParams, fileExist, err := flagExist(execPath)
	if err != nil {
		return nil, fmt.Errorf("flagExist: %w", err)
	}

	// Load the configuration by flag parameter
	if fileExist {
		if err := makeConfigDir(flagFileParams); err != nil {
			return nil, fmt.Errorf("app.makeConfigDir: %w", err)
		}

		filePath := fileParamsToPath(flagFileParams)
		app, err := read(filePath)
		if err != nil {
			return nil, fmt.Errorf("configEngine.Load: %w", err)
		}
		app.filePath = filePath

		return app, nil
	}

	setDefault(execPath, configEngine)
	envFileParams, fileExist, err := envExist(configEngine)
	if err != nil {
		return nil, fmt.Errorf("envExist: %w", err)
	}

	// Load the configuration by environment parameter
	if fileExist {
		if err := makeConfigDir(envFileParams); err != nil {
			return nil, fmt.Errorf("app.makeConfigDir: %w", err)
		}
		filePath := fileParamsToPath(envFileParams)

		app, err := read(filePath)
		if err != nil {
			return nil, fmt.Errorf("configEngine.Load: %w", err)
		}
		app.filePath = filePath

		return app, nil
	}

	app := &App{
	}

	// File doesn't exist, let's write it.
	// Priority is the flag path.
	// If the user didn't pass the flags, then use an environment path.
	// The environment path will not be nil, since it will use the default path.
	if flagFileParams == nil {
		app.fileParams = envFileParams
	}
	if envFileParams == nil {
		return nil, fmt.Errorf("envFileParams is nil")
	}

	if err := makeConfigDir(flagFileParams); err != nil {
		return nil, fmt.Errorf("app.makeConfigDir: %w", err)
	}
	app.filePath = fileParamsToPath(app.fileParams)
	if err := app.write(); err != nil {
		return nil, fmt.Errorf("app.write: %w", err)
	}

	return app, nil
}

// filePath must be absolute
func read(filePath string) (*App, error) {
	var appConfig App

	info, err := os.Stat(filePath)
	if err != nil {
		if err == os.ErrNotExist {
			return &appConfig, nil
		}
		return nil, fmt.Errorf("os.Stat('%s'): %w", filePath, err)
	}

	f, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("os.OpenFile('%s'): %w", filePath, err)
	}

	buf := make([]byte, info.Size())
	_, err = f.Read(buf)
	// All this long piece of code is needed
	// to catch the close error.
	closeErr := f.Close()
	if closeErr != nil {
		if err != nil {
			return nil, fmt.Errorf("%v: file.Close: %w", err, closeErr)
		} else {
			return nil, fmt.Errorf("file.Close: %w", closeErr)
		}
	} else if err != nil {
		return nil, fmt.Errorf("file.Write: %w", err)
	}

	err = yaml.Unmarshal(buf, &appConfig)
	if err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal: %w", err)
	}
	if appConfig.Services == nil {
		appConfig.Services = make([]*service.Service, 0)
	}
	if appConfig.ProxyChains == nil {
		appConfig.ProxyChains = make([]*service.ProxyChain, 0)
	}

	return &appConfig, nil
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

	dir, fileName := path.DirAndFileName(absPath)
	return engine.YamlPathParam(dir, fileName), true, nil
}

// envExist checks is there any configuration file path from env.
// If it exists, checks does it exist in the file system.
//
// In case if it doesn't exist, it will try to load the default configuration.
func envExist(configEngine *engine.Dev) (key_value.KeyValue, bool, error) {
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

	envPath := engine.YamlPathParam(configPath, configName)
	return envPath, exists, nil
}

// setDefault paths of the local file to load by default
func setDefault(execPath string, engine *engine.Dev) {
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
	found := false
	for i, old := range a.Services {
		if old.Id == s.Id {
			found = true
			a.Services[i] = s
			break
		}
	}
	if !found {
		a.Services = append(a.Services, s)
	}

	if err := a.write(); err != nil {
		return fmt.Errorf("app.write: %w", err)
	}

	return nil
}

func (a *App) createYaml() key_value.KeyValue {
	var services = a.Services
	kv := key_value.New()
	kv.Set("services", services)

	return kv
}

// makeConfigDir converts fileParams to the full file path.
// if the configuration file is stored in the nested directory, then those directories are created.
func makeConfigDir(fileParams key_value.KeyValue) error {
	if fileParams == nil {
		return fmt.Errorf("fileParams nil")
	}
	dirPath, err := fileParams.StringValue("configPath")
	if err != nil {
		return fmt.Errorf("a.fileParams.StringValue('configPath'): %w", err)
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

	return nil
}

func fileParamsToPath(fileParams key_value.KeyValue) string {
	name, _ := fileParams.StringValue("name")
	dirPath, _ := fileParams.StringValue("configPath")
	return filepath.Join(dirPath, name+".yml")
}

// Writes the service as the yaml on the given path.
// If the path doesn't contain the file extension, it will through an error
func (a *App) write() error {
	kv := a.createYaml()

	appConfig, err := yaml.Marshal(kv.Map())
	if err != nil {
		return fmt.Errorf("yaml.Marshal: %w", err)
	}

	f, err := os.OpenFile(a.filePath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("os.OpenFile('%s'): %w", a.filePath, err)
	}

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
