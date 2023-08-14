package pipeline

//
// Preparation of the pipeline against service's other pipelines, controllers and service itself.
//
// It doesn't update the data, nor saves the updates.
// These updates are done by the service that calls these functions.

import (
	"fmt"
	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/ahmetson/os-lib/net"
	service2 "github.com/ahmetson/service-lib/service"
	"github.com/ahmetson/service-lib/service/converter"
	"github.com/ahmetson/service-lib/service/orchestra"
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
// - Ensures that controllers exist.
func PrepareAddingPipeline(pipelines []*Pipeline, proxies []string, controllers key_value.KeyValue, pipeline *Pipeline) error {
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

	if pipeline.End.IsController() {
		if err := controllers.Exist(pipeline.End.Id); err != nil {
			return fmt.Errorf("service.Controllers.Exist('%s'): %w", pipeline.End.Id, err)
		}
	} else {
		if FindServiceEnd(pipelines) != nil {
			return fmt.Errorf("config.HasServicePipeline: service pipeline added already")
		}
	}

	return nil
}

func newSourceInstances(config *service2.Controller) []service2.Instance {
	amount := len(config.Instances)
	instances := make([]service2.Instance, 0, amount)

	for i := 0; i < amount; i++ {
		instances[i] = service2.Instance{
			ControllerCategory: service2.SourceName,
			Id:                 fmt.Sprintf("%s-source", config.Instances[i].Id),
			Port:               uint64(net.GetFreePort()),
		}
	}

	return instances
}

func newDestinationInstances(config *service2.Controller) []service2.Instance {
	amount := len(config.Instances)
	instances := make([]service2.Instance, 0, amount)

	for i := 0; i < amount; i++ {
		instances[i] = service2.Instance{
			ControllerCategory: service2.DestinationName,
			Id:                 fmt.Sprintf("%s-destination", config.Instances[i].Id),
			Port:               config.Instances[i].Port,
		}
	}

	return instances
}

// returns generated source and destination controllers.
// the parameters of the destination server derived from config.
func newProxyControllers(config *service2.Controller) []*service2.Controller {
	// set the source
	return []*service2.Controller{
		{
			Type:      config.Type,
			Category:  service2.SourceName,
			Instances: newSourceInstances(config),
		},

		{
			Type:      config.Type,
			Category:  service2.DestinationName,
			Instances: newDestinationInstances(config),
		},
	}
}

// rewriteDestinations removes the destination controllers. Then, set them based on controllers.
func rewriteControllers(proxyConfig *service2.Service, controllers []*service2.Controller) error {
	if len(controllers) == 0 {
		return fmt.Errorf("no destination controllers")
	}
	// two times more, source and destinationService for each server
	proxyConfig.Controllers = make([]*service2.Controller, len(controllers)*2)
	set := 0

	// rewrite the destinations in the dependency
	for _, serviceController := range controllers {
		proxyControllers := newProxyControllers(serviceController)
		for _, config := range proxyControllers {
			proxyConfig.Controllers[set] = config
			set++
		}
	}

	return nil
}

// LintControllers updates the ports of the destination to match the service controllers.
// Return bool is true if ports were updated.
func LintControllers(proxyDestinations []*service2.Controller, serviceControllers []*service2.Controller) (bool, error) {
	updated := false

	// The order of the destinationService should match.
	// Check that ports match, if not then update the ports.
	for i, controllerConfig := range serviceControllers {
		dest := proxyDestinations[i]

		amount := len(controllerConfig.Instances)
		if len(dest.Instances) != amount {
			return false, fmt.Errorf("proxy has %d instances, expecting  %d instances", len(dest.Instances), amount)
		}

		for j := 0; j < amount; j++ {
			if dest.Instances[j].Port != controllerConfig.Instances[j].Port {
				dest.Instances[j].Port = controllerConfig.Instances[j].Port
				updated = true
			}
		}

		if dest.Type != controllerConfig.Type {
			return false, fmt.Errorf("proxy #%d destination type %s mismatches service server: %s", i, dest.Type, controllerConfig.Type)
		}
	}

	return updated, nil
}

// LintProxyToService returns the updated proxy config against the service config.
// Another function LintServiceToProxy updates the service config by proxy config.
//
// The proxyConfig will have the same number of destinations as controllers in the destinationService
func LintProxyToService(proxyConfig *service2.Service, destinationService *service2.Service) (bool, error) {
	if proxyConfig.Type != service2.ProxyType {
		return false, fmt.Errorf("proxyConfig.Type is not proxy")
	}
	if destinationService.Type == service2.ProxyType {
		return false, fmt.Errorf("destinationService.Type is proxy. call LintProxyToProxy()")
	}

	return lintDestinationsToControllers(proxyConfig, destinationService.Controllers)
}

// LintProxyToProxy returns the updated proxy config against another proxy.
// Another function LintServiceToProxy updates the service config by proxy config.
func LintProxyToProxy(proxyConfig *service2.Service, destinationService *service2.Service) (bool, error) {
	if proxyConfig.Type != service2.ProxyType {
		return false, fmt.Errorf("proxyConfig.Type is not proxy")
	}
	if destinationService.Type != service2.ProxyType {
		return false, fmt.Errorf("destinationService.Type is not proxy. call LintProxyToService()")
	}

	sourceControllers, err := destinationService.GetControllers(service2.SourceName)
	if err != nil {
		return false, fmt.Errorf("destinationService(%s).GetControllers(%s): %w", destinationService.Id, service2.SourceName, err)
	}

	return lintDestinationsToControllers(proxyConfig, sourceControllers)
}

func lintDestinationsToControllers(proxyConfig *service2.Service, controllers []*service2.Controller) (bool, error) {
	proxyDestinations, err := proxyConfig.GetControllers(service2.DestinationName)
	if err != nil {
		return false, fmt.Errorf("proxyConfig.GetControllers('%s'): %w", service2.DestinationName, err)
	}

	controllerAmount := len(controllers)

	updated := false
	// The service has more controllers or fewer than in the config.
	// Let's rewrite them
	if len(proxyDestinations) != controllerAmount {
		if err := rewriteControllers(proxyConfig, controllers); err != nil {
			return false, fmt.Errorf("rewriteControllers: %w", err)
		}
		updated = true
	} else {
		if updated, err = LintControllers(proxyDestinations, controllers); err != nil {
			return false, fmt.Errorf("LintControllers: %w", err)
		}
	}

	return updated, nil
}

// LintToControllers lints the pipelines to each other.
//
// Rule for linting:
// If there is a pipeline with the service, then pipelines will lint through that.
func LintToControllers(ctx orchestra.Interface, serviceConfig *service2.Service, pipelines []*Pipeline) error {
	servicePipeline := FindServiceEnd(pipelines)
	serviceProxyConfig, err := ctx.GetConfig(servicePipeline.Beginning())
	if err != nil {
		return fmt.Errorf("orchestra.GetConfig('%s'): %w", servicePipeline.Beginning(), err)
	}
	controllerPipelines := FindControllerEnds(pipelines)

	// lets lint the server's last head destination to the service server's source or
	// to the server itself.
	for _, controllerPipeline := range controllerPipelines {
		if servicePipeline != nil {
			if err := lintLastToProxy(ctx, serviceProxyConfig, controllerPipeline); err != nil {
				return fmt.Errorf("lintLastToService: %w", err)
			}
		} else {
			if err := lintLastToController(ctx, serviceConfig, controllerPipeline); err != nil {
				return fmt.Errorf("lintLastToController: %w", err)
			}
		}

		if err := lintFront(ctx, controllerPipeline); err != nil {
			return fmt.Errorf("lintFront: %w", err)
		}
	}

	return nil
}

func lintLastToProxy(ctx orchestra.Interface, serviceConfig *service2.Service, pipeline *Pipeline) error {
	// lets lint the server's last head destination to the service server's source or
	// to the server itself.
	lastUrl := pipeline.HeadLast()

	lastConfig, err := ctx.GetConfig(lastUrl)
	if err != nil {
		return fmt.Errorf("controllerDep.GetServiceConfig: %w", err)
	}

	updated, err := LintProxyToProxy(lastConfig, serviceConfig)
	if err != nil {
		return fmt.Errorf("controllerPipeline.LintProxyToService: %w", err)
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
func lintLastToController(ctx orchestra.Interface, serviceConfig *service2.Service, pipeline *Pipeline) error {
	// lets lint the server's last head destination to the service server's source or
	// to the server itself.
	lastUrl := pipeline.HeadLast()

	lastConfig, err := ctx.GetConfig(lastUrl)
	if err != nil {
		return fmt.Errorf("controllerDep.GetServiceConfig: %w", err)
	}

	// During the addition of the service, it should validate the server
	controllerConfigs, _ := serviceConfig.GetControllers(pipeline.End.Id)

	if len(lastConfig.Controllers) != 2 {
		return fmt.Errorf("lastConfig(%s).Controllers should have two proxies", lastConfig.Id)
	}

	destinationConfigs, err := lastConfig.GetControllers(service2.DestinationName)
	if err != nil {
		return fmt.Errorf("lastConfig.GetControllers('%s'): %w", service2.DestinationName, err)
	}

	updated, err := LintControllers(destinationConfigs, controllerConfigs)

	if err != nil {
		return fmt.Errorf("controllerPipeline.LintControllers: %w", err)
	}

	if updated {
		err = ctx.SetConfig(lastUrl, lastConfig)
		if err != nil {
			return fmt.Errorf("failed to update source port in dependency porxy: '%s': %w", lastUrl, err)
		}
	}

	return nil
}
func lintLastToService(ctx orchestra.Interface, config *service2.Service, pipeline *Pipeline) error {
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
		return fmt.Errorf("controllerPipeline.LintProxyToService: %w", err)
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

func LintToService(ctx orchestra.Interface, config *service2.Service, pipeline *Pipeline) error {
	if err := lintLastToService(ctx, config, pipeline); err != nil {
		return fmt.Errorf("lintLastToService: %w", err)
	}

	if err := lintFront(ctx, pipeline); err != nil {
		return fmt.Errorf("lintFront: %w", err)
	}

	return nil
}
