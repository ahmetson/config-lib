package app

import (
	"fmt"
	"github.com/ahmetson/config-lib/engine"
	"github.com/ahmetson/config-lib/service"
	"github.com/ahmetson/log-lib"
	"github.com/ahmetson/os-lib/path"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"testing"

	"github.com/ahmetson/datatype-lib/data_type/key_value"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestAppSuite struct {
	suite.Suite
	envPath   string
	appConfig *App
	engine    engine.Interface
	logger    *log.Logger
	execPath  string
}

// Make sure that Account is set to five
// before each test
func (test *TestAppSuite) SetupTest() {
	s := test.Require

	logger, err := log.New("config_test", false)
	s().NoError(err)
	test.logger = logger

	dev, err := engine.NewDev()
	s().NoError(err)
	test.engine = dev

	test.execPath, err = path.CurrentDir()
	s().NoError(err)
}

func (test *TestAppSuite) createYaml(dir string, name string) {
	s := test.Require

	sampleService, err := service.Empty("id", "github.com/ahmetson/sample", service.IndependentType)
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

func (test *TestAppSuite) deleteYaml(dir string, name string) {
	s := test.Require

	filePath := filepath.Join(dir, name+".yml")

	exist, err := path.FileExist(filePath)
	s().NoError(err)

	if !exist {
		return
	}

	s().NoError(os.Remove(filePath))
}

// Test_10_setDefault checks the setup of the default parameters
func (test *TestAppSuite) Test_10_setDefault() {
	s := test.Require

	// No default was set
	s().False(test.engine.Exist(EnvConfigPath))
	s().False(test.engine.Exist(EnvConfigName))

	setDefault(test.execPath, test.engine)

	// After setting the default, the engine must return the parameters
	s().Equal(test.engine.GetString(EnvConfigPath), test.execPath)
	s().Equal(test.engine.GetString(EnvConfigName), "app")
}

// Test_11_envExist test the existence of the configuration file from the environment variables
func (test *TestAppSuite) Test_11_envExist() {
	s := test.Require

	// no data was set, so it must return a false
	envPath, exist, err := envExist(test.engine)
	s().NoError(err)
	s().False(exist)
	s().Nil(envPath)

	// Create a new default file
	setDefault(test.execPath, test.engine)

	configPath := test.engine.GetString(EnvConfigPath)
	configName := test.engine.GetString(EnvConfigName)

	// creating a default file
	test.createYaml(configPath, configName)

	// env exist must succeed by loading the default file
	params, exist, err := envExist(test.engine)
	s().NoError(err)
	s().True(exist)
	s().NotNil(params)
	loadedName, err := params.StringValue("name")
	s().NoError(err)
	s().Equal(configName, loadedName)

	// clean out
	test.deleteYaml(configPath, configName)

	// trying to create a custom file
	configPath = test.execPath
	configName = "sampleFile"
	test.engine.Set(EnvConfigName, configName)
	test.engine.Set(EnvConfigPath, configPath)
	test.createYaml(configPath, configName)

	params, exist, err = envExist(test.engine)
	s().NoError(err)
	s().True(exist)
	loadedName, err = params.StringValue("name")
	s().NoError(err)
	s().Equal(configName, loadedName)

	// clean out the files
	test.deleteYaml(configPath, configName)
}

// Test_12_flagExist test the existence of the configuration file from the arguments
func (test *TestAppSuite) Test_12_flagExist() {
	s := test.Require

	// no flag was given, so it must fail
	_, exist, err := flagExist(test.execPath)
	s().NoError(err)
	s().False(exist)

	configPath := test.execPath
	configName := "example"

	// we give a file name that doesn't exist, yet it must fail
	os.Args = append(os.Args, fmt.Sprintf(`--config=%s/%s.yml`, configPath, configName))
	_, exist, err = flagExist(test.execPath)
	s().Error(err)
	s().False(exist)

	// creating the file and checking flagExist must work
	test.createYaml(configPath, configName)

	// flagExist must work since the argument has the config flag and the file exists too
	params, exist, err := flagExist(test.execPath)
	s().NoError(err)
	s().True(exist)
	loadedName, err := params.StringValue("name")
	s().NoError(err)
	s().Equal(configName, loadedName)

	// clean out
	test.deleteYaml(configPath, configName)

	os.Args = os.Args[:len(os.Args)-1]
}

// Test_13_read tests the loading the services which tries to read the configuration file and parse it.
func (test *TestAppSuite) Test_13_read() {
	s := test.Require

	// let's add a config file from the environment variable
	// Create a new default file
	setDefault(test.execPath, test.engine)

	configPath := test.engine.GetString(EnvConfigPath)
	configName := test.engine.GetString(EnvConfigName)

	// creating a default file
	test.createYaml(configPath, configName)

	configParam, _, _ := envExist(test.engine)

	services, err := read(configParam, test.engine)
	s().NoError(err)
	s().Len(services, 1)

	// clean out
	test.deleteYaml(configPath, configName)
}

// Test_14_New tests the creation of the app from os environment or flag.
// It's the collection of all functions we tested earlier
func (test *TestAppSuite) Test_14_New() {
	s := test.Require

	// let's add a config file from the environment variable
	// Create a new default file
	setDefault(test.execPath, test.engine)

	configPath := test.execPath
	configName := "app"

	// creating a default file
	test.createYaml(configPath, configName)

	app, err := New(test.engine)
	s().NoError(err)
	s().Len(app.Services, 1)

	// clean out
	test.deleteYaml(configPath, configName)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestApp(t *testing.T) {
	suite.Run(t, new(TestAppSuite))
}
