// Package app is the collection of the methods to create or load the app configuration from the engine.
// The app is the collection of the services.
//
// Depends on the engine.
// Managed by the handler.
package app

import (
	"fmt"
	"github.com/ahmetson/config-lib/service"
	"slices"
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
	ids         []string              // all service, handler ids must be unique.
}

// New App configuration.
// If possible, it loads the configuration.
//
// Loading order:
// - create a default app config
// - if app.yml exists in the root, then load it.
// - if, environment variable exists, then load it.
// - if, a flag exists, then load it.
func New() *App {
	appConfig := &App{
		ids: make([]string, 0),
	}
	appConfig.SetEmptyFields()
	return appConfig
}

// IdExist checks is the id unique within the application
func (a *App) IdExist(id string) bool {
	return slices.Contains(a.ids, id)
}

// SetId sets the id in the id list.
// Returns true if id exists.
func (a *App) SetId(id string) bool {
	// if the App was created directly, then ids will be nil
	if a.ids == nil {
		a.ids = []string{id}
		return true
	}

	if a.IdExist(id) {
		return false
	}
	a.ids = append(a.ids, id)
	return true
}

// RegisterId sets all ids of the all services and handlers.
// If there are duplicate id, then throw an error with detail information.
func (a *App) RegisterId() error {
	for _, s := range a.Services {
		if a.IdExist(s.Id) {
			return fmt.Errorf("the '%s' id of service is duplicate", s.Id)
		}
		a.SetId(s.Id)

		for _, h := range s.Handlers {
			if a.IdExist(h.Id) {
				return fmt.Errorf("the '%s' id of handler in '%s' service is duplicate", h.Id, s.Id)
			}

			a.SetId(s.Id)
		}
	}

	return nil
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

// SetEmptyFields sets empty value for nil fields.
// If the developer crated App directly, some fields might be nil
func (a *App) SetEmptyFields() {
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
	if a == nil {
		return fmt.Errorf("app struct is nil")
	}

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

	return nil
}
