// Package handler is defining an entry point to interact with
// the configuration Engine.
package handler

import (
	"fmt"
	"github.com/ahmetson/config-lib/app"
	"github.com/ahmetson/config-lib/engine"
	"github.com/ahmetson/config-lib/service"
	"github.com/ahmetson/datatype-lib/data_type/key_value"
	"github.com/ahmetson/datatype-lib/message"
	"github.com/ahmetson/handler-lib/base"
	handlerConfig "github.com/ahmetson/handler-lib/config"
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
)

type Handler struct {
	Engine   *engine.Dev // todo make it private, for now it's used in the tests of other packages
	app      *app.App
	filePath string
	handler  base.Interface
}

// New handler of the config.
// The handler is initialized for use.
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

	filePath, fileExist, err := app.ReadFileParameters(dev)
	if err != nil {
		return nil, fmt.Errorf("app.ReadFileParameters: %w", err)
	}
	h.app = app.New()
	h.filePath = filePath

	// Load the configuration by flag parameter
	if fileExist {
		if err := app.Read(filePath, h.app); err != nil {
			return nil, fmt.Errorf("read('%s'): %w", filePath, err)
		}
	} else if err := h.writeInitialApp(); err != nil {
		return nil, fmt.Errorf("handler.writeInitialApp: %w", err)
	}

	h.handler = replier.New()
	h.handler.SetConfig(SocketConfig())
	if err := h.handler.SetLogger(logger); err != nil {
		return nil, fmt.Errorf("handler.SetLogger: %w", err)
	}
	if err := h.setRoutes(); err != nil {
		return nil, fmt.Errorf("handler.setRoutes: %w", err)
	}

	return h, nil
}

func (handler *Handler) writeInitialApp() error {
	// File doesn't exist, let's write it.
	// Priority is the flag path.
	// If the user didn't pass the flags, then use an environment path.
	// The environment path will not be nil, since it will use the default path.
	if err := app.MakeConfigDir(handler.filePath); err != nil {
		return fmt.Errorf("app.MakeConfigDir('%s'): %w", handler.filePath, err)
	}

	if err := app.Write(handler.filePath, handler.app); err != nil {
		return fmt.Errorf("app.Write('%s', 'h.app'): %w", handler.filePath, err)
	}

	return nil
}

func (handler *Handler) setRoutes() error {
	if err := handler.handler.Route(ServiceById, handler.onService); err != nil {
		return fmt.Errorf("handler.Route(%s): %w", ServiceById, err)
	}
	if err := handler.handler.Route(ServiceByUrl, handler.onServiceByUrl); err != nil {
		return fmt.Errorf("handler.Route(%s): %w", ServiceByUrl, err)
	}
	if err := handler.handler.Route(SetService, handler.onSetService); err != nil {
		return fmt.Errorf("handler.Route(%s): %w", SetService, err)
	}
	if err := handler.handler.Route(ParamExist, handler.onExist); err != nil {
		return fmt.Errorf("handler.Route(%s): %w", ParamExist, err)
	}
	if err := handler.handler.Route(StringParam, handler.onString); err != nil {
		return fmt.Errorf("handler.Route(%s): %w", StringParam, err)
	}
	if err := handler.handler.Route(Uint64Param, handler.onUint64); err != nil {
		return fmt.Errorf("handler.Route(%s): %w", Uint64Param, err)
	}
	if err := handler.handler.Route(BoolParam, handler.onBool); err != nil {
		return fmt.Errorf("handler.Route(%s): %w", BoolParam, err)
	}
	if err := handler.handler.Route(GenerateHandler, handler.onGenerateHandler); err != nil {
		return fmt.Errorf("handler.Route(%s): %w", GenerateHandler, err)
	}
	if err := handler.handler.Route(SetDefaultParam, handler.onSetDefault); err != nil {
		return fmt.Errorf("handler.Route(%s): %w", SetDefaultParam, err)
	}
	if err := handler.handler.Route(ServiceExist, handler.onServiceExist); err != nil {
		return fmt.Errorf("handler.Route(%s): %w", ServiceExist, err)
	}
	if err := handler.handler.Route(GenerateService, handler.onGenerateService); err != nil {
		return fmt.Errorf("handler.Route(%s): %w", GenerateService, err)
	}

	return nil
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
		return req.Fail(fmt.Sprintf("req.Parameters.StringValue('id'): %v", err))
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
		return req.Fail(fmt.Sprintf("req.Parameters.StringValue('url'): %v", err))
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
		return req.Fail(fmt.Sprintf("req.Parameters.StringValue('handler_type'): %v", err))
	}

	handlerType := handlerConfig.HandlerType(handlerTypeStr)
	if err := handlerConfig.IsValid(handlerType); err != nil {
		return req.Fail(fmt.Sprintf("handlerConfig.IsValid('%s'): %v", handlerTypeStr, err))
	}

	cat, err := req.RouteParameters().StringValue("category")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.StringValue('category'): %v", err))
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

	if err := app.Write(handler.filePath, handler.app); err != nil {
		return req.Fail(fmt.Sprintf("app.Write: %v", err))
	}

	return req.Ok(key_value.New())
}

// onExist checks is the given 'name' exists in the configuration.
func (handler *Handler) onExist(req message.RequestInterface) message.ReplyInterface {
	name, err := req.RouteParameters().StringValue("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.StringValue('name'): %v", err))
	}

	exist := handler.Engine.Exist(name)

	param := key_value.New().Set("exist", exist)
	return req.Ok(param)
}

// onSetDefault set the default parameter in the Engine.
func (handler *Handler) onSetDefault(req message.RequestInterface) message.ReplyInterface {
	name, err := req.RouteParameters().StringValue("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.StringValue('name'): %v", err))
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
		return req.Fail(fmt.Sprintf("req.Parameters.StringValue('id'): %v", err))
	}

	if handler.app.Service(id) != nil {
		return req.Fail(fmt.Sprintf("service('%s') exist", id))
	}

	url, err := req.RouteParameters().StringValue("url")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.StringValue('url'): %v", err))
	}

	typeStr, err := req.RouteParameters().StringValue("type")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.StringValue('type'): %v", err))
	}

	serviceType := service.Type(typeStr)
	if err := service.ValidateServiceType(serviceType); err != nil {
		return req.Fail(fmt.Sprintf("service.ValidateServiceType('%s'): %v", typeStr, err))
	}

	generatedManager, err := service.NewManager(id, url)
	if err != nil {
		return req.Fail(fmt.Sprintf("service.NewManager('%s', '%s'): %v", id, url, err))
	}

	generatedService := service.New(id, url, serviceType, generatedManager)

	params := key_value.New().Set("service", generatedService)
	return req.Ok(params)
}

// onString returns a string parameter from the Engine.
func (handler *Handler) onString(req message.RequestInterface) message.ReplyInterface {
	name, err := req.RouteParameters().StringValue("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.StringValue('name'): %v", err))
	}

	value := handler.Engine.GetString(name)

	param := key_value.New().Set("value", value)
	return req.Ok(param)
}

// onUint64 returns a string parameter from the Engine.
func (handler *Handler) onUint64(req message.RequestInterface) message.ReplyInterface {
	name, err := req.RouteParameters().StringValue("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.StringValue('name'): %v", err))
	}

	value := handler.Engine.GetUint64(name)

	param := key_value.New().Set("value", value)
	return req.Ok(param)
}

// onBool returns a string parameter from the Engine.
func (handler *Handler) onBool(req message.RequestInterface) message.ReplyInterface {
	name, err := req.RouteParameters().StringValue("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.StringValue('name'): %v", err))
	}

	value := handler.Engine.GetBool(name)

	param := key_value.New().Set("value", value)
	return req.Ok(param)
}

func (handler *Handler) Start() error {

	err := handler.handler.Start()
	if err != nil {
		return fmt.Errorf("handler.Start: %w", err)
	}

	return nil
}
