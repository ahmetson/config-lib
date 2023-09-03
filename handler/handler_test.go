package handler

import (
	"fmt"
	"github.com/ahmetson/client-lib"
	"github.com/ahmetson/common-lib/message"
	"github.com/ahmetson/config-lib"
	"github.com/ahmetson/config-lib/app"
	handlerConfig "github.com/ahmetson/handler-lib/config"
	"github.com/ahmetson/log-lib"
	"github.com/ahmetson/os-lib/path"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ahmetson/common-lib/data_type/key_value"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestHandlerSuite struct {
	suite.Suite
	handler    *Handler
	logger     *log.Logger
	client     *client.Socket
	execPath   string
	serviceId  string
	serviceUrl string
}

// Make sure that Account is set to five
// before each test
func (test *TestHandlerSuite) SetupTest() {
	s := test.Require

	logger, err := log.New("config_test", false)
	s().NoError(err)
	test.logger = logger

	// Current directory
	test.execPath, err = path.CurrentDir()
	s().NoError(err)

	// createYaml uses this serviceId and serviceUrl
	test.serviceId = "id"
	test.serviceUrl = "github.com/ahmetson/sample"

	// app engine will load the yaml
	test.createYaml(test.execPath, "app")

	// The config handler
	handler, err := New()
	s().NoError(err)
	test.handler = handler

	err = test.handler.Start()
	s().NoError(err)
	time.Sleep(time.Millisecond * 200) // wait a bit for initialization

	config := SocketConfig()

	// Client that will send requests to the config handler
	zmqType := handlerConfig.SocketType(test.handler.handler.Type())
	socket, err := client.NewRaw(zmqType, fmt.Sprintf("inproc://%s", config.Id))
	s().NoError(err)
	test.client = socket

	// Make the file
	test.handler.engine.SetDefault(app.EnvConfigPath, test.execPath)
	test.handler.engine.SetDefault(app.EnvConfigName, "app")

	// Creating some random config parameters to fetch
	test.handler.engine.Set("bool", true)
	test.handler.engine.Set("string", "hello world")
	test.handler.engine.Set("uint64", uint64(123))
}

func (test *TestHandlerSuite) TearDownTest() {
	s := test.Require

	s().NoError(test.handler.Close())
	s().NoError(test.client.Close())

	time.Sleep(time.Millisecond * 200) // wait a bit for closing threads

	test.deleteYaml(test.execPath, "app")
}

func (test *TestHandlerSuite) createYaml(dir string, name string) {
	s := test.Require

	sampleService := config.Empty(test.serviceId, test.serviceUrl, config.IndependentType)
	kv := key_value.Empty().Set("services", []interface{}{sampleService})

	serviceConfig, err := yaml.Marshal(kv.Map())
	s().NoError(err)

	filePath := filepath.Join(dir, name+".yml")

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	s().NoError(err)
	_, err = f.Write(serviceConfig)
	s().NoError(err)

	s().NoError(f.Close())
}

func (test *TestHandlerSuite) deleteYaml(dir string, name string) {
	s := test.Require

	filePath := filepath.Join(dir, name+".yml")

	exist, err := path.FileExist(filePath)
	s().NoError(err)

	if !exist {
		return
	}

	s().NoError(os.Remove(filePath))
}

// Test_10_ServiceById fetching service by id
func (test *TestHandlerSuite) Test_10_ServiceById() {
	s := test.Require

	// No id parameter was given
	req := message.Request{Command: ServiceById, Parameters: key_value.Empty()}
	rep, err := test.client.Request(&req)
	s().NoError(err)
	s().False(rep.IsOK())

	// Unknown service must return failure
	req.Parameters.Set("id", "unknown")
	rep, err = test.client.Request(&req)
	s().NoError(err)
	s().False(rep.IsOK())

	// Successful request
	req.Parameters.Set("id", test.serviceId)
	rep, err = test.client.Request(&req)
	s().NoError(err)
	s().True(rep.IsOK())
}

// Test_11_ServiceByUrl fetching service by url
func (test *TestHandlerSuite) Test_11_ServiceByUrl() {
	s := test.Require

	// No id parameter was given
	req := message.Request{Command: ServiceByUrl, Parameters: key_value.Empty()}
	req.Parameters.Set("url", test.serviceUrl)
	rep, err := test.client.Request(&req)
	s().NoError(err)
	s().True(rep.IsOK())
}

// Test_12_SetService set a new service
func (test *TestHandlerSuite) Test_12_SetService() {
	s := test.Require

	sampleService := config.Empty(test.serviceId+"_2", test.serviceUrl+"_2", config.IndependentType)

	// No id parameter was given
	req := message.Request{Command: SetService, Parameters: key_value.Empty()}
	req.Parameters.Set("service", sampleService)
	_, err := test.client.Request(&req)
	s().NoError(err)

	// Validate that service was set in the config
	req.Command = ServiceById
	req.Parameters.Set("id", test.serviceId+"_2")
	rep, err := test.client.Request(&req)
	s().NoError(err)
	s().True(rep.IsOK())
}

// Test_13_onExist check parameter exists or not
func (test *TestHandlerSuite) Test_13_onExist() {
	s := test.Require

	param := "not_exist"

	// No id parameter was given
	req := message.Request{Command: ParamExist, Parameters: key_value.Empty()}
	req.Parameters.Set("name", param)
	reply, err := test.client.Request(&req)
	s().NoError(err)
	exist, err := reply.Parameters.GetBoolean("exist")
	s().NoError(err)
	s().False(exist)

	// This parameter set in the test setup, so must exist
	req.Parameters.Set("name", "bool")
	reply, err = test.client.Request(&req)
	s().NoError(err)
	exist, err = reply.Parameters.GetBoolean("exist")
	s().NoError(err)
	s().True(exist)
}

// Test_14_GetParam returns the string, uint or boolean parameters from config engine
func (test *TestHandlerSuite) Test_14_GetParam() {
	s := test.Require

	// No id parameter was given
	req := message.Request{Command: BoolParam, Parameters: key_value.Empty()}
	req.Parameters.Set("name", "bool")
	reply, err := test.client.Request(&req)
	s().NoError(err)
	value, err := reply.Parameters.GetBoolean("value")
	s().NoError(err)
	s().True(value)

	req.Command = StringParam
	req.Parameters.Set("name", "string")
	reply, err = test.client.Request(&req)
	s().NoError(err)
	valueStr, err := reply.Parameters.GetString("value")
	s().NoError(err)
	s().NotEmpty(valueStr)

	req.Command = Uint64Param
	req.Parameters.Set("name", "uint64")
	reply, err = test.client.Request(&req)
	s().NoError(err)
	valueUint64, err := reply.Parameters.GetUint64("value")
	s().NoError(err)
	s().NotZero(valueUint64)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestHandler(t *testing.T) {
	suite.Run(t, new(TestHandlerSuite))
}
