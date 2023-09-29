package config

import (
	"github.com/ahmetson/config-lib/service"
	"github.com/ahmetson/datatype-lib/data_type/key_value"
)

// An Interface of the config engine.
// The configuration is represented as a key-value database.
type Interface interface {
	// Load the specific configuration on a remote file and add it into itself.
	//
	// The parameters are varied for each config engine.
	Load(key_value.KeyValue) (interface{}, error)

	SetDefaults(value key_value.KeyValue)
	SetDefault(string, interface{})
	Set(string, interface{})
	Exist(string) bool
	StringValue(string) string
	Uint64Value(string) uint64
	BoolValue(string) bool
	ServicesValue() ([]*service.Service, error)       // Reads all services. Call it after Load
	ProxyChainsValue() ([]*service.ProxyChain, error) // Reads all proxy chains. Call it after Load
}
