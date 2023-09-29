package service

import (
	"fmt"
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/datatype-lib/data_type/key_value"
	handlerConfig "github.com/ahmetson/handler-lib/config"
)

const (
	ManagerCategory = "manager"
	ConfigFlag      = "config"
)

// Service type defined in the config.
//
// Fields
//   - Type is the type of service. For example, ProxyType, IndependentType or ExtensionType
//   - Url to the service source code
//   - Id within the application
//   - Manager parameter of the service
//   - Handlers that are listed in the service
//   - Extensions that this service depends on
type Service struct {
	Type       Type                     `json:"type" yaml:"type"`
	Url        string                   `json:"url" yaml:"url"`
	Id         string                   `json:"id" yaml:"id"`
	Manager    *clientConfig.Client     `json:"manager" yaml:"manager"`
	Handlers   []*handlerConfig.Handler `json:"handlers" yaml:"handlers"`
	Extensions []*clientConfig.Client   `json:"extensions" yaml:"extensions"`
}

// ManagerId generates a service manager id.
// If no service id, then it will return an error
func ManagerId(serviceId string) string {
	if len(serviceId) == 0 {
		return ""
	}
	return serviceId + "_manager"
}

// NewManager generates a service manager configuration
func NewManager(id string, url string) (*clientConfig.Client, error) {
	if len(id) == 0 || len(url) == 0 {
		return nil, fmt.Errorf("id or url parameter is empty")
	}
	// HewHandler allocates a free port.
	newConfig, err := handlerConfig.NewHandler(handlerConfig.SyncReplierType, ManagerCategory)
	if err != nil {
		return nil, fmt.Errorf("handlerConfig.NewHandler: %w", err)
	}

	socketType := handlerConfig.SocketType(handlerConfig.SyncReplierType)

	managerClient := &clientConfig.Client{
		Id:         ManagerId(id),
		ServiceUrl: url,
		Port:       newConfig.Port,
		TargetType: socketType,
	}

	managerClient.UrlFunc(clientConfig.Url)
	return managerClient, nil
}

// New generates a service configuration.
// It also generates the manager client
func New(id string, url string, serviceType Type, managerClient *clientConfig.Client) *Service {
	//managerClient, err := NewManager(id, url)
	//if err != nil {
	//	return nil, fmt.Errorf("NewManager('%s', '%s'): %w", id, url, err)
	//}

	return &Service{
		Type:       serviceType,
		Id:         id,
		Url:        url,
		Handlers:   make([]*handlerConfig.Handler, 0),
		Extensions: make([]*clientConfig.Client, 0),
		Manager:    managerClient, // connecting to the service from other parents through dev context
	}
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

//// Proxy returns the proxy by its url.
//// If it doesn't exist, returns nil
//func (s *Service) Proxy(id string) *Proxy {
//	for _, p := range s.Proxies {
//		if p.Id == id {
//			return p
//		}
//	}
//
//	return nil
//}

//// SetProxy will set a new proxy.
//// If it exists, it will overwrite it
//func (s *Service) SetProxy(proxy *Proxy) {
//	existing := s.Proxy(proxy.Id)
//	if existing == nil {
//		s.Proxies = append(s.Proxies, proxy)
//	} else {
//		*existing = *proxy
//	}
//}

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
