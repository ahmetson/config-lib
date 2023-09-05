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
	"github.com/ahmetson/handler-lib/replier"
	"github.com/ahmetson/log-lib"
)

const (
	Id              = "config" // only one instance of config Engine can be in the service
	ServiceById     = "service"
	ServiceByUrl    = "service-by-url"
	SetService      = "set-service"
	ParamExist      = "param-exist"
	StringParam     = "string-param"
	Uint64Param     = "uint64-param"
	BoolParam       = "bool-param"
	GenerateHandler = "generate-handler"
	SetDefaultParam = "set-default"
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

	return h, nil
}

// SocketConfig parameter of the handler
func SocketConfig() *handlerConfig.Handler {
	return handlerConfig.NewInternalHandler(handlerConfig.ReplierType, Id)
}

// onService returns a service by service id
func (handler *Handler) onService(req message.Request) message.Reply {
	id, err := req.Parameters.GetString("id")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('id'): %v", err))
	}

	s := handler.app.Service(id)
	if s == nil {
		return req.Fail(fmt.Sprintf("service('%s') not found", id))
	}

	params := key_value.Empty().Set("service", s)
	return req.Ok(params)
}

// onServiceByUrl returns a first occurred service by its url
func (handler *Handler) onServiceByUrl(req message.Request) message.Reply {
	url, err := req.Parameters.GetString("url")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('url'): %v", err))
	}

	s := handler.app.ServiceByUrl(url)
	if s == nil {
		return req.Fail(fmt.Sprintf("serviceByUrl('%s') not found", url))
	}

	params := key_value.Empty().Set("service", s)
	return req.Ok(params)
}

// onGenerateHandler generates the handler parameters
func (handler *Handler) onGenerateHandler(req message.Request) message.Reply {
	internal, err := req.Parameters.GetBoolean("internal")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetBoolean('internal'): %v", err))
	}

	handlerTypeStr, err := req.Parameters.GetString("handler_type")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('handler_type'): %v", err))
	}

	handlerType := handlerConfig.HandlerType(handlerTypeStr)
	if err := handlerConfig.IsValid(handlerType); err != nil {
		return req.Fail(fmt.Sprintf("handlerConfig.IsValid('%s'): %v", handlerTypeStr, err))
	}

	cat, err := req.Parameters.GetString("category")
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

	params := key_value.Empty().Set("handler", generatedConfig)
	return req.Ok(params)
}

// onServiceByUrl updates the service parameters.
func (handler *Handler) onSetService(req message.Request) message.Reply {
	raw, err := req.Parameters.GetKeyValue("service")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetKetValue('service'): %v", err))
	}
	var s service.Service
	err = raw.Interface(&s)
	if err != nil {
		return req.Fail(fmt.Sprintf("raw.Interface: %v", err))
	}

	handler.app.SetService(&s)

	return req.Ok(key_value.Empty())
}

// onExist checks is the given 'name' exists in the configuration.
func (handler *Handler) onExist(req message.Request) message.Reply {
	name, err := req.Parameters.GetString("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('name'): %v", err))
	}

	exist := handler.Engine.Exist(name)

	param := key_value.Empty().Set("exist", exist)
	return req.Ok(param)
}

// onSetDefault set the default parameter in the Engine.
func (handler *Handler) onSetDefault(req message.Request) message.Reply {
	name, err := req.Parameters.GetString("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('name'): %v", err))
	}
	value, ok := req.Parameters["value"]
	if !ok {
		return req.Fail("req.Parameters['value'] not found")
	}

	handler.Engine.SetDefault(name, value)

	param := key_value.Empty()
	return req.Ok(param)
}

// onString returns a string parameter from the Engine.
func (handler *Handler) onString(req message.Request) message.Reply {
	name, err := req.Parameters.GetString("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('name'): %v", err))
	}

	value := handler.Engine.GetString(name)

	param := key_value.Empty().Set("value", value)
	return req.Ok(param)
}

// onUint64 returns a string parameter from the Engine.
func (handler *Handler) onUint64(req message.Request) message.Reply {
	name, err := req.Parameters.GetString("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('name'): %v", err))
	}

	value := handler.Engine.GetUint64(name)

	param := key_value.Empty().Set("value", value)
	return req.Ok(param)
}

// onBool returns a string parameter from the Engine.
func (handler *Handler) onBool(req message.Request) message.Reply {
	name, err := req.Parameters.GetString("name")
	if err != nil {
		return req.Fail(fmt.Sprintf("req.Parameters.GetString('name'): %v", err))
	}

	value := handler.Engine.GetBool(name)

	param := key_value.Empty().Set("value", value)
	return req.Ok(param)
}

func (handler *Handler) Start() error {
	return handler.handler.Start()
}

func (handler *Handler) Close() error {
	return handler.handler.Close()
}
