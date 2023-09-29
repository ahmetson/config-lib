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
	fileParams, fileExist, err := readFileParameters(configEngine)
	if err != nil {
		return nil, fmt.Errorf("flagExist: %w", err)
	}

	// Load the configuration by flag parameter
	if fileExist {
		filePath := fileParamsToPath(fileParams)
		app, err := read(filePath)
		if err != nil {
			return nil, fmt.Errorf("read('%s'): %w", filePath, err)
		}
		app.filePath = filePath

		return app, nil
	}

	if fileParams == nil {
		return nil, fmt.Errorf("file parameter is nil")
	}
	// File doesn't exist, let's write it.
	// Priority is the flag path.
	// If the user didn't pass the flags, then use an environment path.
	// The environment path will not be nil, since it will use the default path.
	if err := makeConfigDir(fileParams); err != nil {
		return nil, fmt.Errorf("app.makeConfigDir: %w", err)
	}

	app := &App{
		filePath: fileParamsToPath(fileParams),
	}

	if err := write(app.filePath, app); err != nil {
		return nil, fmt.Errorf("app.write: %w", err)
	}

	return app, nil
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

// setNewField sets empty value for nil fields
func (a *App) setNewField() {
	if a.Services == nil {
		a.Services = make([]*service.Service, 0)
	}
	if a.ProxyChains == nil {
		a.ProxyChains = make([]*service.ProxyChain, 0)
	}
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

	if err := write(a.filePath, a); err != nil {
		return fmt.Errorf("app.write: %w", err)
	}

	return nil
}
