package engine

import (
	"github.com/ahmetson/datatype-lib/data_type/key_value"
)

// An Interface of the config engine.
// The configuration is represented as a key-value database.
type Interface interface {
	// Load the specific configuration on a remote file and add it into itself.
	//
	// The parameters are varied for each config engine.
	Load(key_value.KeyValue) (interface{}, error)

	// Watch for the changes.
	// For watching, it requires Load to be collected first.
	Watch(func(interface{}, error)) error

	SetDefaults(value key_value.KeyValue)
	SetDefault(string, interface{})
	Set(string, interface{})
	Exist(string) bool
	StringValue(string) string
	Uint64Value(string) uint64
	BoolValue(string) bool
}
