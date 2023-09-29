package app

import (
	"fmt"
	"github.com/ahmetson/config-lib/engine"
	"github.com/ahmetson/config-lib/service"
	"github.com/ahmetson/datatype-lib/data_type/key_value"
	"github.com/ahmetson/os-lib/arg"
	"github.com/ahmetson/os-lib/path"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

//
// Yaml file operations
//

// The readFileParameters returns the file path.
// First it reads from a flag, then from environment variable.
// Lastly, read the default file.
func readFileParameters(configEngine *engine.Dev) (key_value.KeyValue, bool, error) {
	// default app is empty
	execPath, err := path.CurrentDir()
	if err != nil {
		return nil, false, fmt.Errorf("path.CurrentDir: %w", err)
	}

	flagFileParams, fileExist, err := flagExist(execPath)
	if err != nil {
		return nil, false, fmt.Errorf("flagExist: %w", err)
	}

	// Load the configuration by flag parameter
	if fileExist {
		//if err := makeConfigDir(flagFileParams); err != nil {
		//	return nil, fmt.Errorf("app.makeConfigDir: %w", err)
		//}

		return flagFileParams, true, nil
	}

	setDefault(execPath, configEngine)
	envFileParams, fileExist, err := envExist(configEngine)
	if err != nil {
		return nil, true, fmt.Errorf("envExist: %w", err)
	}

	// Load the configuration by environment parameter
	if fileExist {
		//if err := makeConfigDir(envFileParams); err != nil {
		//	return nil, fmt.Errorf("app.makeConfigDir: %w", err)
		//}

		return envFileParams, true, nil
	}

	// flag is priority, if not given then
	fileParams := flagFileParams
	if flagFileParams == nil {
		fileParams = envFileParams
	}

	if fileParams == nil {
		return nil, false, fmt.Errorf("file parameter is nil")
	}

	return fileParams, false, nil
}

// filePath must be absolute
func read(filePath string) (*App, error) {
	var appConfig App

	info, err := os.Stat(filePath)
	if err != nil {
		if err == os.ErrNotExist {
			return &appConfig, nil
		}
		return nil, fmt.Errorf("os.Stat('%s'): %w", filePath, err)
	}

	f, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("os.OpenFile('%s'): %w", filePath, err)
	}

	buf := make([]byte, info.Size())
	_, err = f.Read(buf)
	// All this long piece of code is needed
	// to catch the close error.
	closeErr := f.Close()
	if closeErr != nil {
		if err != nil {
			return nil, fmt.Errorf("%v: file.Close: %w", err, closeErr)
		} else {
			return nil, fmt.Errorf("file.Close: %w", closeErr)
		}
	} else if err != nil {
		return nil, fmt.Errorf("file.Write: %w", err)
	}

	err = yaml.Unmarshal(buf, &appConfig)
	if err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal: %w", err)
	}
	appConfig.setNewField()

	return &appConfig, nil
}

// flagExist checks is there any configuration flag.
// If the configuration flag is set, it checks does it exist in the file system.
func flagExist(execPath string) (key_value.KeyValue, bool, error) {
	if !arg.FlagExist(service.ConfigFlag) {
		return nil, false, nil
	}

	configPath := arg.FlagValue(service.ConfigFlag)

	absPath := path.AbsDir(execPath, configPath)

	exists, err := path.FileExist(absPath)
	if err != nil {
		return nil, false, fmt.Errorf("path.FileExists('%s'): %w", absPath, err)
	}

	if !exists {
		return nil, false, fmt.Errorf("file (%s) not found", absPath)
	}

	dir, fileName := path.DirAndFileName(absPath)
	return engine.YamlPathParam(dir, fileName), true, nil
}

// envExist checks is there any configuration file path from env.
// If it exists, checks does it exist in the file system.
//
// In case if it doesn't exist, it will try to load the default configuration.
func envExist(configEngine *engine.Dev) (key_value.KeyValue, bool, error) {
	if !configEngine.Exist(EnvConfigName) || !configEngine.Exist(EnvConfigPath) {
		return nil, false, nil
	}

	configName := configEngine.GetString(EnvConfigName)
	configPath := configEngine.GetString(EnvConfigPath)
	absPath := path.AbsDir(configPath, configName+".yml")
	exists, err := path.FileExist(absPath)
	if err != nil {
		return nil, false, fmt.Errorf("path.FileExists('%s'): %w", absPath, err)
	}

	envPath := engine.YamlPathParam(configPath, configName)
	return envPath, exists, nil
}

// setDefault paths of the local file to load by default
func setDefault(execPath string, engine *engine.Dev) {
	engine.SetDefault(EnvConfigName, "app")
	engine.SetDefault(EnvConfigPath, execPath)
}

// makeConfigDir converts fileParams to the full file path.
// if the configuration file is stored in the nested directory, then those directories are created.
func makeConfigDir(fileParams key_value.KeyValue) error {
	if fileParams == nil {
		return fmt.Errorf("fileParams nil")
	}
	dirPath, err := fileParams.StringValue("configPath")
	if err != nil {
		return fmt.Errorf("a.fileParams.StringValue('configPath'): %w", err)
	}

	dirExist, err := path.DirExist(dirPath)
	if err != nil {
		return fmt.Errorf("path.DirExist('%s'): %w", dirPath, err)
	}
	if !dirExist {
		err = path.MakeDir(dirPath)
		if err != nil {
			return fmt.Errorf("path.MakeDir('%s'): %w", dirPath, err)
		}
	}

	return nil
}

func fileParamsToPath(fileParams key_value.KeyValue) string {
	name, _ := fileParams.StringValue("name")
	dirPath, _ := fileParams.StringValue("configPath")
	return filepath.Join(dirPath, name+".yml")
}

// Writes the service as the yaml on the given path.
// If the path doesn't contain the file extension, it will through an error
func write(filePath string, data *App) error {
	data.setNewField()
	appConfig, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("yaml.Marshal: %w", err)
	}

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("os.OpenFile('%s'): %w", filePath, err)
	}

	_, err = f.Write(appConfig)
	closeErr := f.Close()
	if closeErr != nil {
		if err != nil {
			return fmt.Errorf("%v: file.Close: %w", err, closeErr)
		} else {
			return fmt.Errorf("file.Close: %w", closeErr)
		}
	} else if err != nil {
		return fmt.Errorf("file.Write: %w", err)
	}

	return nil
}
