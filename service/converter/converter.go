package converter

import (
	"fmt"
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/config-lib"
	"github.com/ahmetson/service-lib/config/service"
)

// ServiceToProxy returns the service in the proxy format
// so that it can be used as a proxy by other services.
//
// If the service has another proxy, then it will find it.
func ServiceToProxy(s *config.Service) (service.Proxy, error) {
	if s.Type != config.ProxyType {
		return service.Proxy{}, fmt.Errorf("only proxy type of service can be converted")
	}

	sourceConfig, err := s.Handler(service.SourceName)
	if err != nil {
		return service.Proxy{}, fmt.Errorf("no source sourceConfig: %w", err)
	}

	if len(sourceConfig.Instances) == 0 {
		return service.Proxy{}, fmt.Errorf("no source instances")
	}

	instance := clientConfig.Client{
		Id: sourceConfig.Category + " instance 01",
	}

	if len(s.Proxies) == 0 {
		instance.Port = sourceConfig.Instances[0].Port
	} else {
		beginning, err := findPipelineBeginning(s, service.SourceName)
		if err != nil {
			return service.Proxy{}, fmt.Errorf("findPipelineBeginning: %w", err)
		}
		instance.Port = beginning.Instances[0].Port
	}

	converted := service.Proxy{
		Url:       s.Url,
		Instances: []*clientConfig.Client{&instance},
	}

	return converted, nil
}

// findPipelineBeginning returns the beginning of the pipeline.
// If the contextType is not a default one, then it will search for the specific orchestra type.
func findPipelineBeginning(s *config.Service, requiredEnd string) (*service.Proxy, error) {
	for _, pipeline := range s.Pipelines {
		beginning := pipeline.Beginning()
		if !pipeline.HasBeginning() {
			return nil, fmt.Errorf("no pipeline beginning")
		}
		//end, err := s.Pipelines.GetString(beginning)
		//if err != nil {
		//	return nil, fmt.Errorf("pipeline '%s' get the end: %w", beginning, err)
		//}
		//
		//if strings.Compare(end, requiredEnd) != 0 {
		//	continue
		//}

		proxy := s.GetProxy(beginning)
		if proxy == nil {
			return nil, fmt.Errorf("invalid config. pipeline '%s' beginning not found in proxy list", beginning)
		}

		return proxy, nil
	}

	return nil, fmt.Errorf("no pipeline beginning '%s' end", requiredEnd)
}

// ServiceToExtension returns the service in the proxy format
// so that it can be used as a proxy
func ServiceToExtension(s *config.Service) (service.Extension, error) {
	if s.Type != config.ExtensionType {
		return service.Extension{}, fmt.Errorf("only proxy type of service can be converted")
	}

	firstHandler, err := s.FirstHandler()
	if err != nil {
		return service.Extension{}, fmt.Errorf("no firstHandler: %w", err)
	}

	if len(firstHandler.Instances) == 0 {
		return service.Extension{}, fmt.Errorf("no handler instances")
	}

	converted := service.Extension{
		Url: s.Url,
		Id:  firstHandler.Category + " instance 01",
	}

	if !s.HasProxy() {
		converted.Port = firstHandler.Instances[0].Port
	} else {
		beginning, err := findPipelineBeginning(s, service.SourceName)
		if err != nil {
			return service.Extension{}, fmt.Errorf("findPipelineBeginning: %w", err)
		}
		converted.Port = beginning.Instances[0].Port
	}

	return converted, nil
}
