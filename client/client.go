package client

import (
	"fmt"
	"github.com/ahmetson/client-lib"
	clientConfig "github.com/ahmetson/client-lib/config"
	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/ahmetson/common-lib/message"
	"github.com/ahmetson/config-lib"
	"github.com/ahmetson/config-lib/handler"
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

	Service(id string) (*config.Service, error)
	ServiceByUrl(url string) (*config.Service, error)
	SetService(s *config.Service) error
	GenerateHandler(handlerType handlerConfig.HandlerType, category string, internal bool) (*handlerConfig.Handler, error)

	Exist(name string) (bool, error)
	String(name string) (string, error)
	Uint64(name string) (uint64, error)
	Bool(name string) (bool, error)
}

func New() (*Client, error) {
	configHandler := handler.SocketConfig()
	socketType := handlerConfig.SocketType(configHandler.Type)
	c := clientConfig.New("", configHandler.Id, configHandler.Port, socketType).
		UrlFunc(handlerConfig.ExternalUrlByClient)

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

func (c *Client) Service(id string) (*config.Service, error) {
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

	var s config.Service
	err = raw.Interface(&s)
	if err != nil {
		return nil, fmt.Errorf("raw.Interface: %v", err)
	}

	return &s, nil
}

func (c *Client) ServiceByUrl(url string) (*config.Service, error) {
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

	var s config.Service
	err = raw.Interface(&s)
	if err != nil {
		return nil, fmt.Errorf("raw.Interface: %v", err)
	}

	return &s, nil
}

// SetService writes the service configuration into the app configuration.
// todo update the yaml file
func (c *Client) SetService(s *config.Service) error {
	req := message.Request{
		Command:    handler.SetService,
		Parameters: key_value.Empty().Set("service", s),
	}

	err := c.socket.Submit(&req)
	if err != nil {
		return fmt.Errorf("socket.Request('%s'): %w", handler.SetService, err)
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
