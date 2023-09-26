package client

import (
	"github.com/ahmetson/config-lib/app"
	"github.com/ahmetson/config-lib/handler"
	"github.com/ahmetson/config-lib/service"
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
type TestClientSuite struct {
	suite.Suite
	handler    *handler.Handler
	logger     *log.Logger
	client     *Client
	execPath   string
	serviceId  string
	serviceUrl string
}

// Make sure that Account is set to five
// before each test
func (test *TestClientSuite) SetupTest() {
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
	h, err := handler.New()
	s().NoError(err)
	test.handler = h

	s().NoError(test.handler.Start())
	time.Sleep(time.Millisecond * 200) // wait a bit for initialization

	// Client that will send requests to the config handler
	c, err := New()
	s().NoError(err)
	test.client = c

	// Make the file
	test.handler.Engine.SetDefault(app.EnvConfigPath, test.execPath)
	test.handler.Engine.SetDefault(app.EnvConfigName, "app")

	// Creating some random config parameters to fetch
	test.handler.Engine.Set("bool", true)
	test.handler.Engine.Set("string", "hello world")
	test.handler.Engine.Set("uint64", uint64(123))
}

func (test *TestClientSuite) TearDownTest() {
	s := test.Require

	if test.client != nil {
		s().NoError(test.client.Close())
		time.Sleep(time.Millisecond * 200) // wait a bit for closing threads
	}

	test.deleteYaml(test.execPath, "app")
}

func (test *TestClientSuite) createYaml(dir string, name string) {
	s := test.Require

	sampleService, err := service.Empty(test.serviceId, test.serviceUrl, service.IndependentType)
	s().NoError(err)
	kv := key_value.New().Set("services", []interface{}{sampleService})

	serviceConfig, err := yaml.Marshal(kv.Map())
	s().NoError(err)

	filePath := filepath.Join(dir, name+".yml")

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	s().NoError(err)
	_, err = f.Write(serviceConfig)
	s().NoError(err)

	s().NoError(f.Close())
}

func (test *TestClientSuite) deleteYaml(dir string, name string) {
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
func (test *TestClientSuite) Test_10_ServiceById() {
	s := test.Require

	// No id parameter was given
	_, err := test.client.Service("")
	s().Error(err)

	// Unknown service must return failure
	_, err = test.client.Service("unknown")
	s().Error(err)

	// Successful request
	returnedService, err := test.client.Service(test.serviceId)
	s().NoError(err)
	s().NotNil(returnedService)
}

// Test_11_ServiceByUrl fetching service by url
func (test *TestClientSuite) Test_11_ServiceByUrl() {
	s := test.Require

	// No id parameter was given
	returnedService, err := test.client.ServiceByUrl(test.serviceUrl)
	s().NoError(err)
	s().NotNil(returnedService)
}

// Test_12_SetService set a new service
func (test *TestClientSuite) Test_12_SetService() {
	s := test.Require

	sampleService, err := service.Empty(test.serviceId+"_2", test.serviceUrl+"_2", service.IndependentType)
	s().NoError(err)

	// No id parameter was given
	err = test.client.SetService(sampleService)
	s().NoError(err)

	// Validate that service was set in the config
	returnedService, err := test.client.Service(test.serviceId + "_2")
	s().NoError(err)
	s().Equal(sampleService.Url, returnedService.Url)
}

// Test_13_onExist check parameter exists or not
func (test *TestClientSuite) Test_13_onExist() {
	s := test.Require

	param := "not_exist"

	// No id parameter was given
	exist, err := test.client.Exist(param)
	s().NoError(err)
	s().False(exist)

	// This parameter set in the test setup, so must exist
	exist, err = test.client.Exist("bool")
	s().NoError(err)
	s().True(exist)
}

// Test_14_GetParam returns the string, uint or boolean parameters from the config engine
func (test *TestClientSuite) Test_14_GetParam() {
	s := test.Require

	// bool
	value, err := test.client.Bool("bool")
	s().NoError(err)
	s().True(value)

	valueStr, err := test.client.String("string")
	s().NoError(err)
	s().NotEmpty(valueStr)

	valueUint64, err := test.client.Uint64("uint64")
	s().NoError(err)
	s().NotZero(valueUint64)
}

// Test_15_GenerateHandler generate a handler
func (test *TestClientSuite) Test_15_GenerateHandler() {
	s := test.Require

	handlerType := handlerConfig.ReplierType
	category := "database"

	// Generate the internal handler configuration
	h, err := test.client.GenerateHandler(handlerType, category, true)
	s().NoError(err)
	s().Zero(h.Port)
	s().Equal(handlerType, h.Type)
	s().Equal(category, h.Category)

	// Generate the tcp handler configuration
	h, err = test.client.GenerateHandler(handlerType, category, false)
	s().NoError(err)
	s().NotZero(h.Port)
	s().Equal(handlerType, h.Type)
	s().Equal(category, h.Category)
}

// Test_16_Close close a handler
func (test *TestClientSuite) Test_16_Close() {
	s := test.Require

	// Close the handler
	err := test.client.Close()
	s().NoError(err)

	// Closed already, so the test suite doesn't have to close them.
	test.handler = nil
	test.client = nil
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestHandler(t *testing.T) {
	suite.Run(t, new(TestClientSuite))
}
