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

type SourceService struct {
	*Proxy
	Manager *clientConfig.Client   `json:"manager" yaml:"manager"`
	Clients []*clientConfig.Client `json:"clients" yaml:"clients"`
}

// The Source defines the proxy
type Source struct {
	Proxies []*SourceService `json:"proxies" yaml:"proxies"`
	Rule    *Rule            `json:"rule,omitempty" yaml:"rule,omitempty"`
}

// Service type defined in the config.
//
// Fields
//   - Type is the type of service. For example, ProxyType, IndependentType or ExtensionType
//   - Url to the service source code
//   - Id within the application
//   - Manager parameter of the service
//   - Handlers that are listed in the service
//   - Extensions that this service depends on
//   - Sources that are can access to this service
type Service struct {
	Type       Type                     `json:"type" yaml:"type"`
	Url        string                   `json:"url" yaml:"url"`
	Id         string                   `json:"id" yaml:"id"`
	Manager    *clientConfig.Client     `json:"manager" yaml:"manager"`
	Handlers   []*handlerConfig.Handler `json:"handlers" yaml:"handlers"`
	Extensions []*clientConfig.Client   `json:"extensions,omitempty" yaml:"extensions,omitempty"`
	Sources    []*Source                `json:"sources,omitempty" yaml:"sources,omitempty"`
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

// IsSourceExist returns true if the sources contains the rule
func IsSourceExist(sources []*Source, rule *Rule) bool {
	return slices.ContainsFunc(sources, func(s *Source) bool {
		return IsEqualRule(s.Rule, rule)
	})
}

// IsEqualSourceService returns true if the fields of both structs match.
func IsEqualSourceService(first *SourceService, second *SourceService) bool {
	if first == nil || second == nil {
		return false
	}

	if len(first.Clients) != len(second.Clients) {
		return false
	}

	if !IsEqualProxy(first.Proxy, second.Proxy) {
		return false
	}

	if !clientConfig.IsEqual(first.Manager, second.Manager) {
		return false
	}

	for i := range first.Clients {
		if !slices.ContainsFunc(second.Clients, func(c *clientConfig.Client) bool {
			return clientConfig.IsEqual(first.Clients[i], c)
		}) {
			return false
		}
	}

	return true
}

func IsSourceServiceExist(proxies []*SourceService, id string) bool {
	return SourceServiceIndex(proxies, id) > -1
}

func SourceServiceIndex(proxies []*SourceService, id string) int {
	return slices.IndexFunc(proxies, func(el *SourceService) bool {
		return el.Id == id
	})
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

// SetServiceSource updates the source.
// Returns true if the Service was updated.
func (s *Service) SetServiceSource(rule *Rule, source *SourceService) bool {
	if s == nil || rule == nil || source == nil {
		return false
	}

	if !IsSourceExist(s.Sources, rule) {
		source := Source{
			Proxies: []*SourceService{source},
			Rule:    rule,
		}

		s.Sources = append(s.Sources, &source)
		return true
	}

	i := slices.IndexFunc(s.Sources, func(s *Source) bool {
		return IsEqualRule(s.Rule, rule)
	})

	if !IsSourceServiceExist(s.Sources[i].Proxies, source.Id) {
		s.Sources[i].Proxies = append(s.Sources[i].Proxies, source)
		return true
	}

	proxyIndex := SourceServiceIndex(s.Sources[i].Proxies, source.Id)

	if IsEqualSourceService(s.Sources[i].Proxies[proxyIndex], source) {
		return false
	}

	s.Sources[i].Proxies[proxyIndex] = source
	return true
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

// SourceExist returns true if the proxy id exists in the sources
func (s *Service) SourceExist(id string) bool {
	return s.SourceById(id) != nil
}

// SourceById returns the manager parameter of the proxy
func (s *Service) SourceById(id string) *SourceService {
	if s == nil || len(s.Sources) == 0 {
		return nil
	}

	for i := range s.Sources {
		source := s.Sources[i]

		if len(source.Proxies) == 0 {
			continue
		}

		found := slices.IndexFunc(source.Proxies, func(proxy *SourceService) bool {
			return proxy != nil && proxy.Id == id
		})
		if found > -1 {
			return source.Proxies[found]
		}
	}

	return nil
}
