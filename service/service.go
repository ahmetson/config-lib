package service

import (
	"fmt"
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/ahmetson/config-lib/engine"
	handlerConfig "github.com/ahmetson/handler-lib/config"
)

const (
	ManagerCategory = "manager"
	ConfigFlag      = "config"
)

// Service type defined in the config
type Service struct {
	Type        Type
	Url         string
	Id          string
	Manager     *clientConfig.Client
	Handlers    []*handlerConfig.Handler
	Proxies     []*Proxy
	ProxyChains []*ProxyChain
	Extensions  []*clientConfig.Client
}

type Services []Service

func ManagerClient(_ string, url string) (*clientConfig.Client, error) {
	newConfig, err := handlerConfig.NewHandler(handlerConfig.SyncReplierType, ManagerCategory)
	if err != nil {
		return nil, fmt.Errorf("handlerConfig.NewHandler: %w", err)
	}

	socketType := handlerConfig.SocketType(handlerConfig.SyncReplierType)

	return &clientConfig.Client{
		Id:         newConfig.Id,
		ServiceUrl: url,
		Port:       newConfig.Port,
		TargetType: socketType,
	}, nil
}

func Empty(id string, url string, serviceType Type) (*Service, error) {
	managerClient, err := ManagerClient(id, url)
	if err != nil {
		return nil, fmt.Errorf("ManagerClient('%s', '%s'): %w", id, url, err)
	}

	return &Service{
		Type:        serviceType,
		Id:          id,
		Url:         url,
		Handlers:    make([]*handlerConfig.Handler, 0),
		Proxies:     make([]*Proxy, 0),
		ProxyChains: make([]*ProxyChain, 0),
		Extensions:  make([]*clientConfig.Client, 0),
		Manager:     managerClient, // connecting to the service from other parents through dev context
	}, nil
}

func Read(engine engine.Interface) (*Service, error) {
	configName := engine.GetString("SERVICE_CONFIG_NAME")
	configPath := engine.GetString("SERVICE_CONFIG_PATH")
	configExt := "yaml"

	value := key_value.New().Set("name", configName).
		Set("type", configExt).
		Set("configPath", configPath)

	file, err := engine.Read(value)
	if err != nil {
		return nil, fmt.Errorf("engine.ReadValue(%s/%s.%s): %w", configPath, configName, configExt, err)
	}

	serv, ok := file.(*Service)
	if !ok {
		return nil, fmt.Errorf("'%s/%s.%s' not a valid Service", configPath, configName, configExt)
	}

	return serv, nil
}

func (s *Service) PrepareService() error {
	err := s.ValidateTypes()
	if err != nil {
		return fmt.Errorf("service.ValidateTypes: %w", err)
	}
	err = s.Lint()
	if err != nil {
		return fmt.Errorf("service.Lint: %w", err)
	}

	return nil
}

// UnmarshalService decodes the yaml into the config.
func UnmarshalService(raw interface{}) (*Service, error) {
	if raw == nil {
		return nil, nil
	}

	kv, err := key_value.NewFromInterface(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to convert raw config service into map: %w", err)
	}

	var serviceConfig Service
	err = kv.Interface(&serviceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert raw config service to config.Service: %w", err)
	}
	err = serviceConfig.PrepareService()
	if err != nil {
		return nil, fmt.Errorf("prepareService: %w", err)
	}

	return &serviceConfig, nil
}

// Lint sets the reference to the parent from the child.
func (s *Service) Lint() error {
	return nil
}

// ValidateTypes the parameters of the service
func (s *Service) ValidateTypes() error {
	if err := ValidateServiceType(s.Type); err != nil {
		return fmt.Errorf("identity.ValidateServiceType: %v", err)
	}

	for _, c := range s.Handlers {
		if err := handlerConfig.IsValid(c.Type); err != nil {
			return fmt.Errorf("handlerConfig.IsValid: %v", err)
		}
	}

	return nil
}

// HandlerByCategory returns the handler config by the handler category.
// If the handler doesn't exist, then it returns an error.
func (s *Service) HandlerByCategory(category string) (*handlerConfig.Handler, error) {
	for _, c := range s.Handlers {
		if c.Category == category {
			return c, nil
		}
	}

	return nil, fmt.Errorf("'%s' category of handler was not found in '%s' service's config", category, s.Url)
}

// HandlersByCategory returns the multiple handlers of the given name.
// If the handlers don't exist, then it returns an error
func (s *Service) HandlersByCategory(name string) ([]*handlerConfig.Handler, error) {
	handlers := make([]*handlerConfig.Handler, 0, len(s.Handlers))
	count := 0

	for _, c := range s.Handlers {
		if c.Category == name {
			handlers[count] = c
			count++
		}
	}

	if len(handlers) == 0 {
		return nil, fmt.Errorf("no '%s' handlers config", name)
	}
	return handlers, nil
}

// FirstHandler returns the handler without requiring its name.
// If the service doesn't have handlers, then it will return an error.
func (s *Service) FirstHandler() (*handlerConfig.Handler, error) {
	if len(s.Handlers) == 0 {
		return nil, fmt.Errorf("service '%s' doesn't have any handlers in yaml file", s.Url)
	}

	handlers := s.Handlers[0]
	return handlers, nil
}

// ExtensionByUrl returns the first occurred extension config by the url.
// If the extension doesn't exist, then it returns nil
func (s *Service) ExtensionByUrl(url string) *clientConfig.Client {
	for _, e := range s.Extensions {
		if e.ServiceUrl == url {
			return e
		}
	}

	return nil
}

// Proxy returns the proxy by its url. If it doesn't exist, returns nil
func (s *Service) Proxy(id string) *Proxy {
	for _, p := range s.Proxies {
		if p.Id == id {
			return p
		}
	}

	return nil
}

// SetProxy will set a new proxy. If it exists, it will overwrite it
func (s *Service) SetProxy(proxy *Proxy) {
	existing := s.Proxy(proxy.Id)
	if existing == nil {
		s.Proxies = append(s.Proxies, proxy)
	} else {
		*existing = *proxy
	}
}

// SetExtension will set a new extension. If it exists, it will overwrite it
func (s *Service) SetExtension(extension *clientConfig.Client) {
	existing := s.ExtensionByUrl(extension.ServiceUrl)
	if existing == nil {
		s.Extensions = append(s.Extensions, extension)
	} else {
		*existing = *extension
	}
}

// SetHandler adds a new handler. If the handler by the same name exists, it will add a new copy.
func (s *Service) SetHandler(handler *handlerConfig.Handler) {
	s.Handlers = append(s.Handlers, handler)
}

//func (s *Service) SetPipeline(pipeline *pipeline.Pipeline) {
//s.Pipelines = append(s.Pipelines, pipeline)
//}

func (s *Service) HasProxy() bool {
	return len(s.Proxies) > 0
}
