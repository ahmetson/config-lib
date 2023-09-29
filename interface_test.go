package config

import (
	"github.com/ahmetson/config-lib/engine"
	"github.com/ahmetson/os-lib/path"
	"os"
	"path/filepath"
	"testing"

	"github.com/ahmetson/datatype-lib/data_type/key_value"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestEngineInterfaceSuite struct {
	suite.Suite
	envPath   string
	appConfig *engine.Dev
}

// Make sure that Account is set to five
// before each test
func (suite *TestEngineInterfaceSuite) SetupTest() {
	os.Args = append(os.Args, "--plain")
	os.Args = append(os.Args, "--security-debug")
	os.Args = append(os.Args, "--number-key=5")

	envFile := "TRUE_KEY=true\n" +
		"FALSE_KEY=false\n" +
		"STRING_KEY=hello world\n" +
		"NUMBER_KEY=123\n" +
		"FLOAT_KEY=75.321\n"
	execPath, err := path.CurrentDir()
	suite.Require().NoError(err)

	suite.envPath = filepath.Join(execPath, ".test.env")

	os.Args = append(os.Args, suite.envPath)

	file, err := os.Create(suite.envPath)
	suite.Require().NoError(err)
	_, err = file.WriteString(envFile)
	suite.Require().NoError(err, "failed to write the data into: "+suite.envPath)
	err = file.Close()
	suite.Require().NoError(err, "delete the dump file: "+suite.envPath)

	suite.Require().NoError(err)
	appConfig, err := engine.NewDev()
	suite.Require().NoError(err)
	suite.appConfig = appConfig

}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *TestEngineInterfaceSuite) TestRun() {
	suite.Require().False(suite.appConfig.Secure)

	var confInterface Interface = suite.appConfig

	suite.Require().False(confInterface.Exist("TURKISH_KEY"))
	defaultConfig := key_value.New().
		// never will be written since env is already written
		Set("STRING_KEY", "salam").
		Set("TURKISH_KEY", "salam")

	confInterface.SetDefaults(defaultConfig)
	suite.Require().True(confInterface.Exist("TURKISH_KEY"))
	suite.Require().Equal(confInterface.GetString("TURKISH_KEY"), "salam")

	key := "NOT_FOUND"
	value := "random_text"
	suite.Require().False(confInterface.Exist(key))
	suite.Require().Empty(confInterface.GetString(key))
	confInterface.SetDefault(key, value)
	suite.Require().Equal(confInterface.GetString(key), value)

	suite.Require().True(confInterface.Exist("TRUE_KEY"))
	suite.Require().True(confInterface.GetBool("TRUE_KEY"))
	suite.Require().True(confInterface.Exist("FALSE_KEY"))
	suite.Require().False(confInterface.GetBool("FALSE_KEY"))
	suite.Require().Equal(confInterface.GetString("STRING_KEY"), "hello world")
	suite.Require().Equal(uint64(123), confInterface.GetUint64("NUMBER_KEY"))
	suite.Require().True(confInterface.Exist("FLOAT_KEY"))
	suite.Require().Equal(confInterface.GetString("FLOAT_KEY"), "75.321")
	suite.Require().Empty(confInterface.GetUint64("FLOAT_KEY"))

	err := os.Remove(suite.envPath)
	suite.Require().NoError(err, "delete the dump file: "+suite.envPath)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestInterface(t *testing.T) {
	suite.Run(t, new(TestEngineInterfaceSuite))
}
