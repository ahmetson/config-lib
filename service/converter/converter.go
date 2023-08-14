package converter

import (
	"fmt"
	service2 "github.com/ahmetson/service-lib/service"
)

// ServiceToProxy returns the service in the proxy format
// so that it can be used as a proxy by other services.
//
// If the service has another proxy, then it will find it.
func ServiceToProxy(s *service2.Service) (service2.Proxy, error) {
	if s.Type != service2.ProxyType {
		return service2.Proxy{}, fmt.Errorf("only proxy type of service can be converted")
	}

	controllerConfig, err := s.GetController(service2.SourceName)
	if err != nil {
		return service2.Proxy{}, fmt.Errorf("no source controllerConfig: %w", err)
	}

	if len(controllerConfig.Instances) == 0 {
		return service2.Proxy{}, fmt.Errorf("no source instances")
	}

	instance := service2.Instance{
		Id: controllerConfig.Category + " instance 01",
	}

	if len(s.Proxies) == 0 {
		instance.Port = controllerConfig.Instances[0].Port
	} else {
		beginning, err := findPipelineBeginning(s, service2.SourceName)
		if err != nil {
			return service2.Proxy{}, fmt.Errorf("findPipelineBeginning: %w", err)
		}
		instance.Port = beginning.Instances[0].Port
	}

	converted := service2.Proxy{
		Url:       s.Url,
		Instances: []service2.Instance{instance},
	}

	return converted, nil
}

// findPipelineBeginning returns the beginning of the pipeline.
// If the contextType is not a default one, then it will search for the specific orchestra type.
func findPipelineBeginning(s *service2.Service, requiredEnd string) (*service2.Proxy, error) {
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
func ServiceToExtension(s *service2.Service) (service2.Extension, error) {
	if s.Type != service2.ExtensionType {
		return service2.Extension{}, fmt.Errorf("only proxy type of service can be converted")
	}

	controllerConfig, err := s.GetFirstController()
	if err != nil {
		return service2.Extension{}, fmt.Errorf("no controllerConfig: %w", err)
	}

	if len(controllerConfig.Instances) == 0 {
		return service2.Extension{}, fmt.Errorf("no server instances")
	}

	converted := service2.Extension{
		Url: s.Url,
		Id:  controllerConfig.Category + " instance 01",
	}

	if !s.HasProxy() {
		converted.Port = controllerConfig.Instances[0].Port
	} else {
		beginning, err := findPipelineBeginning(s, service2.SourceName)
		if err != nil {
			return service2.Extension{}, fmt.Errorf("findPipelineBeginning: %w", err)
		}
		converted.Port = beginning.Instances[0].Port
	}

	return converted, nil
}
