package config

import (
	"github.com/ahmetson/datatype-lib/data_type/key_value"
)

// An Interface of the config engine.
// The configuration is represented as a key-value database.
type Interface interface {
	SetDefaults(value key_value.KeyValue)
	Exist(string) bool
}
