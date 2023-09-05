package service

const (
	// SourceName of this type should be listed within the handlers in the config
	SourceName = "source"
	// DestinationName of this type should be listed within the handlers in the config
	DestinationName = "destination"
)

type End string

const (
	RouteEnd   = End("route")
	HandlerEnd = End("handler")
	ServiceEnd = End("service")
)

type Proxy struct {
	Id       string `json:"id"`
	End      string `json:"end,omitempty"`      // only shown for a proxy
	Command  string `json:"command,omitempty"`  // if End is RouteEnd
	Category string `json:"category,omitempty"` // if End is HandlerEnd
}
