package engine

import (
	"github.com/ahmetson/datatype-lib/data_type/key_value"
)

// An Interface GetServiceConfig Engine based on viper.Viper
type Interface interface {
	// Read the specific configuration on a remote file and add it into itself
	Read(key_value.KeyValue) (interface{}, error)

	// Watch for the changes
	Watch(func(interface{}, error)) error

	SetDefaults(value key_value.KeyValue)
	SetDefault(string, interface{})
	Set(string, interface{})
	Exist(string) bool
	GetString(string) string
	GetUint64(string) uint64
	GetBool(string) bool
}
