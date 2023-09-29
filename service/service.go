package service

import (
	"fmt"
	clientConfig "github.com/ahmetson/client-lib/config"
	handlerConfig "github.com/ahmetson/handler-lib/config"
	"slices"
	"strings"
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
	return &Service{
		Type:       serviceType,
		Id:         id,
		Url:        url,
		Handlers:   make([]*handlerConfig.Handler, 0),
		Extensions: make([]*clientConfig.Client, 0),
		Manager:    managerClient, // connecting to the service from other parents through dev context
	}
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
	if len(category) == 0 {
		return nil, fmt.Errorf("category argument is empty")
	}

	i := slices.IndexFunc(s.Handlers, func(e *handlerConfig.Handler) bool {
		return strings.Compare(e.Category, category) == 0
	})
	if i == -1 {
		return nil, fmt.Errorf("handler of '%s' category not found", category)
	}

	return s.Handlers[i], nil
}

// HandlersByCategory returns the multiple handlers of the given name.
// If the handlers don't exist, then it returns an error
func (s *Service) HandlersByCategory(category string) ([]*handlerConfig.Handler, error) {
	if len(category) == 0 {
		return nil, fmt.Errorf("category argument is empty")
	}

	handlers := make([]*handlerConfig.Handler, 0)
	i := 0

	for _, c := range s.Handlers {
		if strings.Compare(c.Category, category) == 0 {
			handlers = slices.Grow(handlers, 1)
			handlers = slices.Insert(handlers, i, c)
			i++
		}
	}

	if len(handlers) == 0 {
		return nil, fmt.Errorf("no '%s' handlers config", category)
	}
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

// SetExtension will set a new extension. If it exists, it will overwrite it
func (s *Service) SetExtension(extension *clientConfig.Client) {
	existing := s.ExtensionByUrl(extension.ServiceUrl)
	if existing == nil {
		s.Extensions = append(s.Extensions, extension)
	} else {
		*existing = *extension
	}
}

// SetHandler adds a new handler.
// If the handler by the same id exists, it will over-write that handler.
func (s *Service) SetHandler(handler *handlerConfig.Handler) {
	if s == nil {
		return
	}

	if len(s.Handlers) == 0 {
		s.Handlers = []*handlerConfig.Handler{handler}
		return
	}

	i := slices.IndexFunc(s.Handlers, func(h *handlerConfig.Handler) bool {
		return strings.Compare(h.Id, handler.Id) == 0
	})

	if i == -1 {
		s.Handlers = append(s.Handlers, handler)
		return
	}

	s.Handlers[i] = handler
}
