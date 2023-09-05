package service

import (
	"fmt"
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/ahmetson/config-lib/engine"
	handlerConfig "github.com/ahmetson/handler-lib/config"
	"path/filepath"
	"strings"
)

const (
	ConfigFlag = "config"
)

// Service type defined in the config
type Service struct {
	Type       Type
	Url        string
	Id         string
	Manager    *clientConfig.Client
	Handlers   []*handlerConfig.Handler
	Proxies    []*Proxy
	Extensions []*clientConfig.Client
	ProxyOrder []string
}

type Services []Service

func ManagerClient(id string, url string) (*clientConfig.Client, error) {
	newConfig, err := handlerConfig.NewHandler(handlerConfig.SyncReplierType, id+"_manager")
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
		Type:       serviceType,
		Id:         id,
		Url:        url,
		Handlers:   make([]*handlerConfig.Handler, 0),
		Proxies:    make([]*Proxy, 0),
		Extensions: make([]*clientConfig.Client, 0),
		ProxyOrder: []string{},
		Manager:    managerClient, // connecting to the service from other parents through dev context
	}, nil
}

func Read(engine engine.Interface) (*Service, error) {
	configName := engine.GetString("SERVICE_CONFIG_NAME")
	configPath := engine.GetString("SERVICE_CONFIG_PATH")
	configExt := "yaml"

	value := key_value.Empty().Set("name", configName).
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

// Handler returns the handler config by the handler name.
// If the handler doesn't exist, then it returns an error.
func (s *Service) Handler(name string) (*handlerConfig.Handler, error) {
	for _, c := range s.Handlers {
		if c.Category == name {
			return c, nil
		}
	}

	return nil, fmt.Errorf("'%s' handler was not found in '%s' service's config", name, s.Url)
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

func CreateYaml(configs ...*Service) key_value.KeyValue {
	var services = configs
	kv := key_value.Empty()
	kv.Set("Services", services)

	return kv
}

// validateServicePath returns an error if the path is not a valid .yml link
func validateServicePath(path string) error {
	if len(path) < 5 || len(filepath.Base(path)) < 5 {
		return fmt.Errorf("path is too short")
	}
	_, found := strings.CutSuffix(path, ".yml")
	if !found {
		return fmt.Errorf("the path should end with '.yml'")
	}

	return nil
}

//// WriteService writes the service as the yaml on the given path.
//// If the path doesn't contain the file extension, it will through an error
//func (ctx *Context) SetConfig(url string, service *config.Service) error {
//	path := ctx.ConfigurationPath(url)
//
//	if err := validateServicePath(path); err != nil {
//		return fmt.Errorf("validateServicePath: %w", err)
//	}
//
//	kv := CreateYaml(service)
//
//	serviceConfig, err := yaml.Marshal(kv.Map())
//	if err != nil {
//		return fmt.Errorf("failed to marshall config.Service: %w", err)
//	}
//
//	f, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
//	_, err = f.Write(serviceConfig)
//	closeErr := f.Close()
//	if err != nil {
//		return fmt.Errorf("failed to write service into the given path: %w", err)
//	} else if closeErr != nil {
//		return fmt.Errorf("failed to close the file descriptor: %w", closeErr)
//	} else {
//		return nil
//	}
//}
//

//
//// GetConfig on the given path.
//// If a path is not obsolete, then it should be relative to the executable.
//// The path should have the .yml extension
//func (ctx *Context) GetConfig(url string) (*config.Service, error) {
//	path := ctx.ConfigurationPath(url)
//
//	if err := validateServicePath(path); err != nil {
//		return nil, fmt.Errorf("validateServicePath: %w", err)
//	}
//
//	bytes, err := os.ReadFile(path)
//	if err != nil {
//		return nil, fmt.Errorf("os.ReadFile of %s: %w", path, err)
//	}
//
//	yamlConfig := createYaml()
//	kv := yamlConfig.Map()
//	err = yaml.Unmarshal(bytes, &kv)
//
//	if err != nil {
//		return nil, fmt.Errorf("yaml.Unmarshal of %s: %w", path, err)
//	}
//
//	fmt.Println("service", kv)
//
//	yamlConfig = key_value.NewDev(kv)
//	if err := yamlConfig.Exist("Services"); err != nil {
//		return nil, fmt.Errorf("no services in yaml: %w", err)
//	}
//
//	services, err := yamlConfig.GetKeyValueList("Services")
//	if err != nil {
//		return nil, fmt.Errorf("failed to get services as key value list: %w", err)
//	}
//
//	if len(services) == 0 {
//		return nil, fmt.Errorf("no services in the config")
//	}
//
//	var serviceConfig config.Service
//	err = services[0].Interface(&serviceConfig)
//	if err != nil {
//		return nil, fmt.Errorf("convert key value to Service: %w", err)
//	}
//
//	err = serviceConfig.PrepareService()
//	if err != nil {
//		return nil, fmt.Errorf("prepareService: %w", err)
//	}
//
//	return &serviceConfig, nil
//}
//
