package pipeline

//
// Preparation of the pipeline against service's other pipelines, handlers and service itself.
//
// It doesn't update the data, nor saves the updates.
// These updates are done by the service that calls these functions.

import (
	"fmt"
	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/ahmetson/config-lib"
	"github.com/ahmetson/config-lib/service"
	handlerConfig "github.com/ahmetson/handler-lib/config"
	"github.com/ahmetson/os-lib/net"
	"github.com/ahmetson/service-lib/config/service/converter"
	"github.com/ahmetson/service-lib/orchestra"
	"slices"
)

// PrepareAddingPipeline validates a new pipeline
// against the service parameters.
//
// Call this function before adding a pipeline into the service.
//
// - Validates the pipeline.
// - Ensures the proxies exist
// - Ensures that only one service pipeline exists.
// - Ensures that handlers exist.
func PrepareAddingPipeline(pipelines []*Pipeline, proxies []string, handlers key_value.KeyValue, pipeline *Pipeline) error {
	if !pipeline.HasLength() {
		return fmt.Errorf("no proxy")
	}
	if err := pipeline.ValidateHead(); err != nil {
		return fmt.Errorf("pipeline.ValidateHead: %w", err)
	}

	for _, proxyUrl := range pipeline.Head {
		if !slices.Contains(proxies, proxyUrl) {
			return fmt.Errorf("proxy '%s' url not required", proxyUrl)
		}
	}

	if pipeline.End.IsHandler() {
		if err := handlers.Exist(pipeline.End.Id); err != nil {
			return fmt.Errorf("service.Handlers.Exist('%s'): %w", pipeline.End.Id, err)
		}
	} else {
		if FindServiceEnd(pipelines) != nil {
			return fmt.Errorf("config.HasServicePipeline: service pipeline added already")
		}
	}

	return nil
}

func newSourceInstances(config *handlerConfig.Handler) []handlerConfig.Instance {
	amount := len(config.Instances)
	instances := make([]handlerConfig.Instance, 0, amount)

	for i := 0; i < amount; i++ {
		instances[i] = handlerConfig.Instance{
			Category: service.SourceName,
			Id:       fmt.Sprintf("%s-source", config.Instances[i].Id),
			Port:     uint64(net.GetFreePort()),
		}
	}

	return instances
}

func newDestinationInstances(config *handlerConfig.Handler) []handlerConfig.Instance {
	amount := len(config.Instances)
	instances := make([]handlerConfig.Instance, 0, amount)

	for i := 0; i < amount; i++ {
		instances[i] = handlerConfig.Instance{
			Category: service.DestinationName,
			Id:       fmt.Sprintf("%s-destination", config.Instances[i].Id),
			Port:     config.Instances[i].Port,
		}
	}

	return instances
}

// returns generated source and destination handlers.
// the parameters of the destination handler derived from config.
func newProxyHandlers(config *handlerConfig.Handler) []*handlerConfig.Handler {
	// set the source
	return []*handlerConfig.Handler{
		{
			Type:      config.Type,
			Category:  service.SourceName,
			Instances: newSourceInstances(config),
		},

		{
			Type:      config.Type,
			Category:  service.DestinationName,
			Instances: newDestinationInstances(config),
		},
	}
}

// rewriteDestinations removes the destination handlers. Then, set them based on handlers.
func rewriteHandlers(proxyConfig *config.Service, handlers []*handlerConfig.Handler) error {
	if len(handlers) == 0 {
		return fmt.Errorf("no destination handlers")
	}
	// two times more, source and destinationService for each handler
	proxyConfig.Handlers = make([]*handlerConfig.Handler, len(handlers)*2)
	set := 0

	// rewrite the destinations in the dependency
	for _, serviceHandler := range handlers {
		proxyHandlers := newProxyHandlers(serviceHandler)
		for _, _config := range proxyHandlers {
			proxyConfig.Handlers[set] = _config
			set++
		}
	}

	return nil
}

// LintHandlers updates the ports of the destination to match the service handlers.
// Return bool is true if ports were updated.
func LintHandlers(proxyDestinations []*handlerConfig.Handler, handlers []*handlerConfig.Handler) (bool, error) {
	updated := false

	// The order of the destinationService should match.
	// Check that ports match, if not then update the ports.
	for i, handler := range handlers {
		dest := proxyDestinations[i]

		amount := len(handler.Instances)
		if len(dest.Instances) != amount {
			return false, fmt.Errorf("proxy has %d instances, expecting  %d instances", len(dest.Instances), amount)
		}

		for j := 0; j < amount; j++ {
			if dest.Instances[j].Port != handler.Instances[j].Port {
				dest.Instances[j].Port = handler.Instances[j].Port
				updated = true
			}
		}

		if dest.Type != handler.Type {
			return false, fmt.Errorf("proxy #%d destination type %s mismatches service handler: %s", i, dest.Type, handler.Type)
		}
	}

	return updated, nil
}

// LintProxyToService returns the updated proxy config against the service config.
// Another function LintServiceToProxy updates the service config by proxy config.
//
// The proxyConfig will have the same number of destinations as handlers in the destinationService
func LintProxyToService(proxyConfig *config.Service, destinationService *config.Service) (bool, error) {
	if proxyConfig.Type != config.ProxyType {
		return false, fmt.Errorf("proxyConfig.Type is not proxy")
	}
	if destinationService.Type == config.ProxyType {
		return false, fmt.Errorf("destinationService.Type is proxy. call LintProxyToProxy()")
	}

	return lintDestinationsToHandlers(proxyConfig, destinationService.Handlers)
}

// LintProxyToProxy returns the updated proxy config against another proxy.
// Another function LintServiceToProxy updates the service config by proxy config.
func LintProxyToProxy(proxyConfig *config.Service, destinationService *config.Service) (bool, error) {
	if proxyConfig.Type != config.ProxyType {
		return false, fmt.Errorf("proxyConfig.Type is not proxy")
	}
	if destinationService.Type != config.ProxyType {
		return false, fmt.Errorf("destinationService.Type is not proxy. call LintProxyToService()")
	}

	sourceHandlers, err := destinationService.HandlersByCategory(service.SourceName)
	if err != nil {
		return false, fmt.Errorf("destinationService(%s).HandlersByCategory(%s): %w", destinationService.Id, service.SourceName, err)
	}

	return lintDestinationsToHandlers(proxyConfig, sourceHandlers)
}

func lintDestinationsToHandlers(proxyConfig *config.Service, handlers []*handlerConfig.Handler) (bool, error) {
	proxyDestinations, err := proxyConfig.HandlersByCategory(service.DestinationName)
	if err != nil {
		return false, fmt.Errorf("proxyConfig.HandlersByCategory('%s'): %w", service.DestinationName, err)
	}

	amount := len(handlers)

	updated := false
	// The service has more handlers or fewer than in the config.
	// Let's rewrite them
	if len(proxyDestinations) != amount {
		if err := rewriteHandlers(proxyConfig, handlers); err != nil {
			return false, fmt.Errorf("rewriteHandlers: %w", err)
		}
		updated = true
	} else {
		if updated, err = LintHandlers(proxyDestinations, handlers); err != nil {
			return false, fmt.Errorf("LintHandlers: %w", err)
		}
	}

	return updated, nil
}

// LintToHandlers lints the pipelines to each other.
//
// Rule for linting:
// If there is a pipeline with the service, then pipelines will lint through that.
func LintToHandlers(ctx orchestra.Interface, serviceConfig *config.Service, pipelines []*Pipeline) error {
	servicePipeline := FindServiceEnd(pipelines)
	serviceProxyConfig, err := ctx.GetConfig(servicePipeline.Beginning())
	if err != nil {
		return fmt.Errorf("orchestra.GetConfig('%s'): %w", servicePipeline.Beginning(), err)
	}
	handlerPipelines := FindHandlerEnds(pipelines)

	// lets lint the handler's last head destination to the service handler's source or
	// to the handler itself.
	for _, handlerPipeline := range handlerPipelines {
		if servicePipeline != nil {
			if err := lintLastToProxy(ctx, serviceProxyConfig, handlerPipeline); err != nil {
				return fmt.Errorf("lintLastToService: %w", err)
			}
		} else {
			if err := lintToLastHandler(ctx, serviceConfig, handlerPipeline); err != nil {
				return fmt.Errorf("lintToLastHandler: %w", err)
			}
		}

		if err := lintFront(ctx, handlerPipeline); err != nil {
			return fmt.Errorf("lintFront: %w", err)
		}
	}

	return nil
}

func lintLastToProxy(ctx orchestra.Interface, serviceConfig *config.Service, pipeline *Pipeline) error {
	// lets lint the handler's last head destination to the service handler's source or
	// to the handler itself.
	lastUrl := pipeline.HeadLast()

	lastConfig, err := ctx.GetConfig(lastUrl)
	if err != nil {
		return fmt.Errorf("dep.GetServiceConfig: %w", err)
	}

	updated, err := LintProxyToProxy(lastConfig, serviceConfig)
	if err != nil {
		return fmt.Errorf("LintProxyToService: %w", err)
	}

	if updated {
		err = ctx.SetConfig(lastUrl, lastConfig)
		if err != nil {
			return fmt.Errorf("failed to update source port in dependency porxy: '%s': %w", lastUrl, err)
		}

		converted, err := converter.ServiceToProxy(lastConfig)
		if err != nil {
			return fmt.Errorf("failed to convert the proxy")
		}

		serviceConfig.SetProxy(&converted)
	}

	return nil
}
func lintToLastHandler(ctx orchestra.Interface, serviceConfig *config.Service, pipeline *Pipeline) error {
	// lets lint the handler's last head destination to the service handler's source or
	// to the handler itself.
	lastUrl := pipeline.HeadLast()

	lastConfig, err := ctx.GetConfig(lastUrl)
	if err != nil {
		return fmt.Errorf("ctx.GetServiceConfig: %w", err)
	}

	// During the addition of the service, it should validate the handler
	handlerConfigs, _ := serviceConfig.HandlersByCategory(pipeline.End.Id)

	if len(lastConfig.Controllers) != 2 {
		return fmt.Errorf("lastConfig(%s).Handlers should have two proxies", lastConfig.Id)
	}

	destinationConfigs, err := lastConfig.GetControllers(service.DestinationName)
	if err != nil {
		return fmt.Errorf("lastConfig.HandlersByCategory('%s'): %w", service.DestinationName, err)
	}

	updated, err := LintHandlers(destinationConfigs, handlerConfigs)

	if err != nil {
		return fmt.Errorf("handlerPipeline.LintHandlers: %w", err)
	}

	if updated {
		err = ctx.SetConfig(lastUrl, lastConfig)
		if err != nil {
			return fmt.Errorf("failed to update source port in dependency porxy: '%s': %w", lastUrl, err)
		}
	}

	return nil
}
func lintLastToService(ctx orchestra.Interface, config *config.Service, pipeline *Pipeline) error {
	// bridge the proxies between the proxies
	if !pipeline.IsMultiHead() {
		return nil
	}

	// lets lint the service's last head's destination to this service
	lastUrl := pipeline.HeadLast()

	lastConfig, err := ctx.GetConfig(lastUrl)
	if err != nil {
		return fmt.Errorf("orchestra.GetServiceConfig: %w", err)
	}

	updated, err := LintProxyToService(lastConfig, config)
	if err != nil {
		return fmt.Errorf("handlerPipeline.LintProxyToService: %w", err)
	}

	if updated {
		err = ctx.SetConfig(lastUrl, lastConfig)
		if err != nil {
			return fmt.Errorf("failed to update source port in dependency porxy: '%s': %w", lastUrl, err)
		}

		converted, err := converter.ServiceToProxy(lastConfig)
		if err != nil {
			return fmt.Errorf("failed to convert the proxy")
		}

		config.SetProxy(&converted)
	}

	return nil
}

func lintFront(ctx orchestra.Interface, pipeline *Pipeline) error {
	// lets lint the service's last head's destination to this service
	lastUrl := pipeline.HeadLast()

	lastConfig, err := ctx.GetConfig(lastUrl)
	if err != nil {
		return fmt.Errorf("orchestra.GetServiceConfig: %w", err)
	}

	// make sure that they link to each other after linting the last head
	proxyUrls := pipeline.HeadFront()

	i := len(proxyUrls) - 1
	for ; i >= 0; i-- {
		proxyUrl := proxyUrls[i]
		proxyConfig, err := ctx.GetConfig(proxyUrl)
		if err != nil {
			return fmt.Errorf("controllerDep.GetServiceConfig: %w", err)
		}

		updated, err := LintProxyToProxy(proxyConfig, lastConfig)
		if err != nil {
			return fmt.Errorf("controllerPipeline.LintProxyToProxy(%s, %s): %w", proxyConfig.Id, lastConfig.Id, err)
		}

		if updated {
			err = ctx.SetConfig(proxyUrl, proxyConfig)
			if err != nil {
				return fmt.Errorf("failed to update source port in dependency porxy: '%s': %w", proxyUrl, err)
			}
		}

		lastUrl = proxyUrl
		lastConfig = proxyConfig
	}

	return nil
}

func LintToService(ctx orchestra.Interface, config *config.Service, pipeline *Pipeline) error {
	if err := lintLastToService(ctx, config, pipeline); err != nil {
		return fmt.Errorf("lintLastToService: %w", err)
	}

	if err := lintFront(ctx, pipeline); err != nil {
		return fmt.Errorf("lintFront: %w", err)
	}

	return nil
}
