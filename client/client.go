package client

import (
	"fmt"
	"github.com/ahmetson/client-lib"
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/ahmetson/common-lib/message"
	"github.com/ahmetson/config-lib/handler"
	"github.com/ahmetson/config-lib/service"
	handlerConfig "github.com/ahmetson/handler-lib/config"
	"time"
)

type Client struct {
	socket *client.Socket
}

type Interface interface {
	Close() error
	Timeout(duration time.Duration)
	Attempt(attempt uint8)

	Service(id string) (*service.Service, error)
	ServiceByUrl(url string) (*service.Service, error)
	SetService(s *service.Service) error
	GenerateHandler(handlerType handlerConfig.HandlerType, category string, internal bool) (*handlerConfig.Handler, error)

	Exist(name string) (bool, error)
	String(name string) (string, error)
	Uint64(name string) (uint64, error)
	Bool(name string) (bool, error)
	SetDefault(name string, value interface{}) error
	ServiceExist(id string) (bool, error)
	ServiceExistByUrl(url string) (bool, error)
	GenerateService(id string, url string, serviceType service.Type) (*service.Service, error)
}

func New() (*Client, error) {
	configHandler := handler.SocketConfig()
	socketType := handlerConfig.SocketType(configHandler.Type)
	c := clientConfig.New("", configHandler.Id, configHandler.Port, socketType).
		UrlFunc(clientConfig.Url)

	socket, err := client.New(c)
	if err != nil {
		return nil, fmt.Errorf("client.New: %w", err)
	}

	return &Client{socket: socket}, nil
}

func (c *Client) Close() error {
	if c.socket != nil {
		return c.socket.Close()
	}

	return nil
}

func (c *Client) Timeout(duration time.Duration) {
	c.socket.Timeout(duration)
}

func (c *Client) Attempt(attempt uint8) {
	c.socket.Attempt(attempt)
}

func (c *Client) Service(id string) (*service.Service, error) {
	req := message.Request{
		Command:    handler.ServiceById,
		Parameters: key_value.Empty().Set("id", id),
	}

	rep, err := c.socket.Request(&req)
	if err != nil {
		return nil, fmt.Errorf("socket.Request('%s'): %w", handler.ServiceById, err)
	}

	if !rep.IsOK() {
		return nil, fmt.Errorf("replied an error: %s", rep.Message)
	}

	raw, err := rep.Parameters.GetKeyValue("service")
	if err != nil {
		return nil, fmt.Errorf("rep.Parameters.GetKeyValue('service'): %v", err)
	}

	var s service.Service
	err = raw.Interface(&s)
	if err != nil {
		return nil, fmt.Errorf("raw.Interface: %v", err)
	}

	return &s, nil
}

func (c *Client) ServiceByUrl(url string) (*service.Service, error) {
	req := message.Request{
		Command:    handler.ServiceByUrl,
		Parameters: key_value.Empty().Set("url", url),
	}

	rep, err := c.socket.Request(&req)
	if err != nil {
		return nil, fmt.Errorf("socket.Request('%s'): %w", handler.ServiceByUrl, err)
	}

	if !rep.IsOK() {
		return nil, fmt.Errorf("replied an error: %s", rep.Message)
	}

	raw, err := rep.Parameters.GetKeyValue("service")
	if err != nil {
		return nil, fmt.Errorf("rep.Parameters.GetKeyValue('service'): %v", err)
	}

	var s service.Service
	err = raw.Interface(&s)
	if err != nil {
		return nil, fmt.Errorf("raw.Interface: %v", err)
	}

	return &s, nil
}

// SetService writes the service configuration into the app configuration.
// todo update the yaml file
func (c *Client) SetService(s *service.Service) error {
	req := message.Request{
		Command:    handler.SetService,
		Parameters: key_value.Empty().Set("service", s),
	}

	err := c.socket.Submit(&req)
	if err != nil {
		return fmt.Errorf("socket.Submit('%s'): %w", handler.SetService, err)
	}

	return nil
}

// GenerateHandler creates a configuration that could be added into the service
func (c *Client) GenerateHandler(handlerType handlerConfig.HandlerType, category string, internal bool) (*handlerConfig.Handler, error) {
	req := message.Request{
		Command: handler.GenerateHandler,
		Parameters: key_value.Empty().
			Set("internal", internal).
			Set("category", category).
			Set("handler_type", handlerType),
	}

	rep, err := c.socket.Request(&req)
	if err != nil {
		return nil, fmt.Errorf("socket.Request('%s'): %w", handler.SetService, err)
	}

	if !rep.IsOK() {
		return nil, fmt.Errorf("replied an error: %s", rep.Message)
	}

	raw, err := rep.Parameters.GetKeyValue("handler")
	if err != nil {
		return nil, fmt.Errorf("rep.Parameters.GetKeyValue('handler'): %v", err)
	}

	var h handlerConfig.Handler
	err = raw.Interface(&h)
	if err != nil {
		return nil, fmt.Errorf("raw.Interface: %v", err)
	}

	return &h, nil
}

// GenerateService creates a configuration of a service
func (c *Client) GenerateService(id string, url string, serviceType service.Type) (*service.Service, error) {
	req := message.Request{
		Command: handler.GenerateService,
		Parameters: key_value.Empty().
			Set("id", id).
			Set("url", url).
			Set("type", serviceType),
	}

	rep, err := c.socket.Request(&req)
	if err != nil {
		return nil, fmt.Errorf("socket.Request('%s'): %w", handler.SetService, err)
	}

	if !rep.IsOK() {
		return nil, fmt.Errorf("replied an error: %s", rep.Message)
	}

	raw, err := rep.Parameters.GetKeyValue("service")
	if err != nil {
		return nil, fmt.Errorf("rep.Parameters.GetKeyValue('handler'): %v", err)
	}

	var s service.Service
	err = raw.Interface(&s)
	if err != nil {
		return nil, fmt.Errorf("raw.Interface: %v", err)
	}

	return &s, nil
}

// Exist checks whether the given parameter exists in the config
func (c *Client) Exist(name string) (bool, error) {
	req := message.Request{
		Command:    handler.ParamExist,
		Parameters: key_value.Empty().Set("name", name),
	}

	rep, err := c.socket.Request(&req)
	if err != nil {
		return false, fmt.Errorf("socket.Request('%s'): %w", handler.ParamExist, err)
	}

	if !rep.IsOK() {
		return false, fmt.Errorf("replied an error: %s", rep.Message)
	}

	exist, err := rep.Parameters.GetBoolean("exist")
	if err != nil {
		return false, fmt.Errorf("rep.Parameters.GetKeyValue('service'): %v", err)
	}

	return exist, nil
}

// String parameter from config engine
func (c *Client) String(name string) (string, error) {
	req := message.Request{
		Command:    handler.StringParam,
		Parameters: key_value.Empty().Set("name", name),
	}

	rep, err := c.socket.Request(&req)
	if err != nil {
		return "", fmt.Errorf("socket.Request('%s'): %w", handler.StringParam, err)
	}

	if !rep.IsOK() {
		return "", fmt.Errorf("replied an error: %s", rep.Message)
	}

	value, err := rep.Parameters.GetString("value")
	if err != nil {
		return "", fmt.Errorf("rep.Parameters.GetString('value'): %v", err)
	}

	return value, nil
}

// Uint64 parameter from config engine
func (c *Client) Uint64(name string) (uint64, error) {
	req := message.Request{
		Command:    handler.Uint64Param,
		Parameters: key_value.Empty().Set("name", name),
	}

	rep, err := c.socket.Request(&req)
	if err != nil {
		return 0, fmt.Errorf("socket.Request('%s'): %w", handler.Uint64Param, err)
	}

	if !rep.IsOK() {
		return 0, fmt.Errorf("replied an error: %s", rep.Message)
	}

	value, err := rep.Parameters.GetUint64("value")
	if err != nil {
		return 0, fmt.Errorf("rep.Parameters.GetUint64('value'): %v", err)
	}

	return value, nil
}

// Bool parameter from config engine
func (c *Client) Bool(name string) (bool, error) {
	req := message.Request{
		Command:    handler.BoolParam,
		Parameters: key_value.Empty().Set("name", name),
	}

	rep, err := c.socket.Request(&req)
	if err != nil {
		return false, fmt.Errorf("socket.Request('%s'): %w", handler.BoolParam, err)
	}

	if !rep.IsOK() {
		return false, fmt.Errorf("replied an error: %s", rep.Message)
	}

	value, err := rep.Parameters.GetBoolean("value")
	if err != nil {
		return false, fmt.Errorf("rep.Parameters.GetBoolean('value'): %v", err)
	}

	return value, nil
}

// SetDefault sets the default value
func (c *Client) SetDefault(name string, value interface{}) error {
	req := message.Request{
		Command:    handler.SetDefaultParam,
		Parameters: key_value.Empty().Set("name", name).Set("value", value),
	}

	err := c.socket.Submit(&req)
	if err != nil {
		return fmt.Errorf("socket.Submit('%s'): %w", handler.ParamExist, err)
	}

	return nil
}

// ServiceExist checks whether the service exists or not
func (c *Client) ServiceExist(id string) (bool, error) {
	return c.serviceExist("id", id)
}

// ServiceExistByUrl checks whether the service exists or not by its url
func (c *Client) ServiceExistByUrl(url string) (bool, error) {
	return c.serviceExist("url", url)
}

// ServiceExist checks whether the service exists or not by the parameter
func (c *Client) serviceExist(name string, value string) (bool, error) {
	req := message.Request{
		Command:    handler.ServiceExist,
		Parameters: key_value.Empty().Set(name, value),
	}

	reply, err := c.socket.Request(&req)
	if err != nil {
		return false, fmt.Errorf("socket.Request('%s'): %w", handler.ServiceExist, err)
	}

	if !reply.IsOK() {
		return false, fmt.Errorf("reply.Message: %s", reply.Message)
	}

	exist, err := reply.Parameters.GetBoolean("exist")
	if err != nil {
		return false, fmt.Errorf("reply.Parameters.GetBoolean('exist'): %w", err)
	}

	return exist, nil
}
