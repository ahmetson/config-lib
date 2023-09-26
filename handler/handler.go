// Package handler is defining an entry point to interact with
// the configuration Engine.
package handler

import (
	"fmt"
	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/ahmetson/common-lib/message"
	"github.com/ahmetson/config-lib/app"
	"github.com/ahmetson/config-lib/engine"
	"github.com/ahmetson/config-lib/service"
	"github.com/ahmetson/handler-lib/base"
	handlerConfig "github.com/ahmetson/handler-lib/config"
	"github.com/ahmetson/handler-lib/manager_client"
	"github.com/ahmetson/handler-lib/replier"
	"github.com/ahmetson/log-lib"
)

const (
	Id              = "dev_config_handler" // only one instance of config Engine can be in the service
	ServiceById     = "service"
	ServiceByUrl    = "service-by-url"
	ServiceExist    = "service-exist"
	SetService      = "set-service"
	ParamExist      = "param-exist"
	StringParam     = "string-param"
	Uint64Param     = "uint64-param"
	BoolParam       = "bool-param"
	GenerateHandler = "generate-handler"
	SetDefaultParam = "set-default"
	GenerateService = "generate-service"
	Close           = "close"
)

type Handler struct {
	Engine  engine.Interface
	app     *app.App
	handler base.Interface
}

// New handler of the config
func New() (*Handler, error) {
	h := &Handler{}

	logger, err := log.New("config", false)
	if err != nil {
		return nil, fmt.Errorf("log.New('config'): %w", err)
	}

	dev, err := engine.NewDev()
	if err != nil {
		return nil, fmt.Errorf("Engine.NewDev: %w", err)
	}
	h.Engine = dev

	allConfig, err := app.New(dev)
	if err != nil {
		return nil, fmt.Errorf("app.New: %w", err)
	}
	h.app = allConfig

	h.handler = replier.New()
	h.handler.SetConfig(SocketConfig())
	if err := h.handler.SetLogger(logger); err != nil {
		return nil, fmt.Errorf("handler.SetLogger: %w", err)
	}
	if err := h.handler.Route(ServiceById, h.onService); err != nil {
		return nil, fmt.Errorf("handler.Route(%s): %w", ServiceById, err)
	}
	if err := h.handler.Route(ServiceByUrl, h.onServiceByUrl); err != nil {
		return nil, fmt.Errorf("handler.Route(%s): %w", ServiceByUrl, err)
	}
	if err := h.handler.Route(SetService, h.onSetService); err != nil {
		return nil, fmt.Errorf("handler.Route(%s): %w", SetService, err)
	}
	if err := h.handler.Route(ParamExist, h.onExist); err != nil {
		return nil, fmt.Errorf("handler.Route(%s): %w", ParamExist, err)
	}
	if err := h.handler.Route(StringParam, h.onString); err != nil {
		return nil, fmt.Errorf("handler.Route(%s): %w", StringParam, err)
	}
	if err := h.handler.Route(Uint64Param, h.onUint64); err != nil {
		return nil, fmt.Errorf("handler.Route(%s): %w", Uint64Param, err)
	}
	if err := h.handler.Route(BoolParam, h.onBool); err != nil {
		return nil, fmt.Errorf("handler.Route(%s): %w", BoolParam, err)
	}
	if err := h.handler.Route(GenerateHandler, h.onGenerateHandler); err != nil {
		return nil, fmt.Errorf("handler.Route(%s): %w", GenerateHandler, err)
	}
	if err := h.handler.Route(SetDefaultParam, h.onSetDefault); err != nil {
		return nil, fmt.Errorf("handler.Route(%s): %w", SetDefaultParam, err)
	}
	if err := h.handler.Route(ServiceExist, h.onServiceExist); err != nil {
		return nil, fmt.Errorf("handler.Route(%s): %w", ServiceExist, err)
	}
	if err := h.handler.Route(GenerateService, h.onGenerateService); err != nil {
		return nil, fmt.Errorf("handler.Route(%s): %w", GenerateService, err)
	}
	if err := h.handler.Route(Close, h.onClose); err != nil {
		return nil, fmt.Errorf("handler.Route(%s): %w", Close, err)
	}

	return h, nil
}

// SocketConfig parameter of the handler
func SocketConfig() *handlerConfig.Handler {
	return handlerConfig.NewInternalHandler(handlerConfig.ReplierType, Id)
}

// onServiceExist checks whether the service exist or not.
// Either checks by the 'id' or by the 'url' parameter.
//
// Returns 'exist' parameter.
func (handler *Handler) onServiceExist(req message.RequestInterface) message.ReplyInterface {
	id, err := req.RouteParameters().StringValue("id")
	if err == nil {
		s := handler.app.Service(id)
		exist := s != nil
		params := key_value.New().Set("exist", exist)
		return req.Ok(params)
	}

	url, err := req.RouteParameters().StringValue("url")
	if err == nil {
		s := handler.app.ServiceByUrl(url)
		exist := s != nil
		params := key_value.New().Set("exist", exist)
		return req.Ok(params)
	}

	return req.Fail(fmt.Sprintf("need 'id' or 'url' parameter"))
}

// onService returns a service by service id
func (handler *Handler) onService(req message.RequestInterface) message.ReplyInterface {
	id, err := req.RouteParameters().StringValue("id")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('id'): %v", err))
	}

	s := handler.app.Service(id)
	if s == nil {
		return req.Fail(fmt.Sprintf("service('%s') not found", id))
	}

	params := key_value.New().Set("service", s)
	return req.Ok(params)
}

// onServiceByUrl returns a first occurred service by its url
func (handler *Handler) onServiceByUrl(req message.RequestInterface) message.ReplyInterface {
	url, err := req.RouteParameters().StringValue("url")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('url'): %v", err))
	}

	s := handler.app.ServiceByUrl(url)
	if s == nil {
		return req.Fail(fmt.Sprintf("serviceByUrl('%s') not found", url))
	}

	params := key_value.New().Set("service", s)
	return req.Ok(params)
}

// onGenerateHandler generates the handler parameters
func (handler *Handler) onGenerateHandler(req message.RequestInterface) message.ReplyInterface {
	internal, err := req.RouteParameters().BoolValue("internal")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetBoolean('internal'): %v", err))
	}

	handlerTypeStr, err := req.RouteParameters().StringValue("handler_type")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('handler_type'): %v", err))
	}

	handlerType := handlerConfig.HandlerType(handlerTypeStr)
	if err := handlerConfig.IsValid(handlerType); err != nil {
		return req.Fail(fmt.Sprintf("handlerConfig.IsValid('%s'): %v", handlerTypeStr, err))
	}

	cat, err := req.RouteParameters().StringValue("category")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('category'): %v", err))
	}

	if len(cat) == 0 {
		return req.Fail("the 'category' is empty")
	}

	var generatedConfig *handlerConfig.Handler
	if internal {
		generatedConfig = handlerConfig.NewInternalHandler(handlerType, cat)
	} else {
		generatedConfig, err = handlerConfig.NewHandler(handlerType, cat)
		if err != nil {
			return req.Fail(fmt.Sprintf("handlerConfig.NewHandler(handler_type: '%s', cat: '%s'): %v", handlerTypeStr, cat, err))
		}
	}

	params := key_value.New().Set("handler", generatedConfig)
	return req.Ok(params)
}

// onServiceByUrl updates the service parameters.
func (handler *Handler) onSetService(req message.RequestInterface) message.ReplyInterface {
	raw, err := req.RouteParameters().NestedValue("service")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetKetValue('service'): %v", err))
	}
	var s service.Service
	err = raw.Interface(&s)
	if err != nil {
		return req.Fail(fmt.Sprintf("raw.Interface: %v", err))
	}

	err = handler.app.SetService(&s)
	if err != nil {
		return req.Fail(fmt.Sprintf("app.SetService: %v", err))
	}

	return req.Ok(key_value.New())
}

// onExist checks is the given 'name' exists in the configuration.
func (handler *Handler) onExist(req message.RequestInterface) message.ReplyInterface {
	name, err := req.RouteParameters().StringValue("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('name'): %v", err))
	}

	exist := handler.Engine.Exist(name)

	param := key_value.New().Set("exist", exist)
	return req.Ok(param)
}

// onSetDefault set the default parameter in the Engine.
func (handler *Handler) onSetDefault(req message.RequestInterface) message.ReplyInterface {
	name, err := req.RouteParameters().StringValue("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('name'): %v", err))
	}
	value, ok := req.RouteParameters()["value"]
	if !ok {
		return req.Fail("req.Parameters['value'] not found")
	}

	handler.Engine.SetDefault(name, value)

	param := key_value.New()
	return req.Ok(param)
}

// onGenerateService generates the service parameters
//
// todo write the service into the yaml
func (handler *Handler) onGenerateService(req message.RequestInterface) message.ReplyInterface {
	id, err := req.RouteParameters().StringValue("id")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('id'): %v", err))
	}

	if handler.app.Service(id) != nil {
		return req.Fail(fmt.Sprintf("service('%s') exist", id))
	}

	url, err := req.RouteParameters().StringValue("url")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('url'): %v", err))
	}

	typeStr, err := req.RouteParameters().StringValue("type")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('type'): %v", err))
	}

	serviceType := service.Type(typeStr)
	if err := service.ValidateServiceType(serviceType); err != nil {
		return req.Fail(fmt.Sprintf("service.ValidateServiceType('%s'): %v", typeStr, err))
	}

	generatedService, err := service.Empty(id, url, serviceType)
	if err != nil {
		return req.Fail(fmt.Sprintf("service.Empty('%s', '%s', '%s'): %v", id, url, serviceType, err))
	}

	params := key_value.New().Set("service", generatedService)
	return req.Ok(params)
}

// onString returns a string parameter from the Engine.
func (handler *Handler) onString(req message.RequestInterface) message.ReplyInterface {
	name, err := req.RouteParameters().StringValue("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('name'): %v", err))
	}

	value := handler.Engine.GetString(name)

	param := key_value.New().Set("value", value)
	return req.Ok(param)
}

// onUint64 returns a string parameter from the Engine.
func (handler *Handler) onUint64(req message.RequestInterface) message.ReplyInterface {
	name, err := req.RouteParameters().StringValue("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('name'): %v", err))
	}

	value := handler.Engine.GetUint64(name)

	param := key_value.New().Set("value", value)
	return req.Ok(param)
}

// onBool returns a string parameter from the Engine.
func (handler *Handler) onBool(req message.RequestInterface) message.ReplyInterface {
	name, err := req.RouteParameters().StringValue("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('name'): %v", err))
	}

	value := handler.Engine.GetBool(name)

	param := key_value.New().Set("value", value)
	return req.Ok(param)
}

// onClose receives a close signal to close the handler and underlying engine.
func (handler *Handler) onClose(req message.RequestInterface) message.ReplyInterface {

	managerClient, err := manager_client.New(SocketConfig())
	if err != nil {
		return req.Fail(fmt.Sprintf("handler.Close: %v", err))
	}
	err = managerClient.Close()
	if err != nil {
		return req.Fail(fmt.Sprintf("managerClient.Close: %v", err))
	}

	param := key_value.Empty()
	return req.Ok(param)
}

func (handler *Handler) Start() error {

	err := handler.handler.Start()
	if err != nil {
		return fmt.Errorf("handler.Start: %w", err)
	}

	return nil
}
