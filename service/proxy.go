package service

import "github.com/ahmetson/client-lib/config"

const (
	// SourceName of this type should be listed within the handlers in the config
	SourceName = "source"
	// DestinationName of this type should be listed within the handlers in the config
	DestinationName = "destination"
)

type Proxy struct {
	Url       string
	Instances []*config.Client
}
