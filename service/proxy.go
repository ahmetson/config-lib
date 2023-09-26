package service

import "fmt"

//
// Usage in the service:
//
// p := serviceConfig.NewProxy("id", "url")
// end := serviceConfig.NewHandlerEnd("category")
// p.SetEnd(end)
//
// service.SetProxy(p)
//
// --- internal
// The service passes the proxy to the dependency manager.
// The dependency manager installs the proxy.
// The installed proxy runs with the configuration file passed from the service.
// The installed proxy accepts a --destination and --id.
// From the destination, installed proxy gets the destination service.
//
// Proxy requests from destination the next Endpoint.

type EndType string

const (
	RouteEnd   = EndType("route")
	HandlerEnd = EndType("handler")
)

// Proxy in the service.
//
// If two endpoints lint to the same route,
type Proxy struct {
	Id       string    `json:"id"`
	Url      string    `json:"url"`           // in the beginning, use this to initialize the proxy automatically.
	Endpoint *Endpoint `json:"end,omitempty"` // only shown for a proxy
}

// Endpoint of the proxy pipe
type Endpoint struct {
	Type     EndType `json:"type"`              // either RouteEnd or HandlerEnd
	Url      string  `json:"url"`               // The id of the service
	Category string  `json:"value"`             // if Type is RouteEnd, this is the command id
	Command  string  `json:"command,omitempty"` // if provided, then it's a RouteEnd
}

type ProxyChain struct {
	Endpoint *Endpoint `json:"endpoint"`
	Proxies  []*Proxy  `json:"proxies"`
}

func NewProxy(id string, url string) *Proxy {
	return &Proxy{
		Id:  id,
		Url: url,
	}
}

func NewHandlerEnd(id string, category string) *Endpoint {
	return &Endpoint{
		Type:     HandlerEnd,
		Url:      id,
		Category: category,
		Command:  "",
	}
}

func NewRouteEnd(id string, category string, command string) *Endpoint {
	return &Endpoint{
		Type:     RouteEnd,
		Url:      id,
		Category: category,
		Command:  command,
	}
}

func (p *Proxy) SetEnd(end *Endpoint) *Proxy {
	p.Endpoint = end
	return p
}

func IsEqualEnd(first *Endpoint, second *Endpoint) bool {
	return first != nil && second != nil &&
		first.Type == second.Type &&
		first.Url == second.Url &&
		first.Category == second.Category &&
		first.Command == second.Command
}

func IsProxySet(proxies []*Proxy, id string) bool {
	for _, proxy := range proxies {
		if proxy.Id == id {
			return true
		}
	}

	return false
}

// Chain returns the proxy chain for the given endpoint.
func Chain(proxyChains []*ProxyChain, end *Endpoint) []*Proxy {
	for _, proxyChain := range proxyChains {
		if proxyChain.Endpoint == nil {
			continue
		}

		if IsEqualEnd(proxyChain.Endpoint, end) {
			return proxyChain.Proxies
		}
	}

	return []*Proxy{}
}

// IsHandlerEndExist returns true if there is a proxy that links to the given handler category
func IsHandlerEndExist(proxies []*Proxy, id string, category string) bool {
	for _, proxy := range proxies {
		if proxy.Endpoint != nil &&
			proxy.Endpoint.Url == id &&
			proxy.Endpoint.Category == category {
			return true
		}
	}

	return false
}

// IsEndpointExist returns true if the given endpoint exists in the proxy list
func IsEndpointExist(proxies []*Proxy, endpoint *Endpoint) bool {
	if endpoint == nil {
		return false
	}

	for _, proxy := range proxies {
		if IsEqualEnd(proxy.Endpoint, endpoint) {
			return true
		}
	}

	return false
}

// ValidProxyChain verifies that endpoints are set correctly.
//
// If the proxy type is route, and there is a handler of the same type, then make sure that
// there is a proxy chain of the end.
func ValidProxyChain(proxies []*Proxy, proxyChains []*ProxyChain) error {
	for _, proxy := range proxies {
		if proxy.Endpoint == nil {
			return fmt.Errorf("proxy('%s').Endpoint is nil. Call Proxy.SetEnd", proxy.Id)
		}

		if proxy.Endpoint.Type != RouteEnd {
			continue
		}

		id := proxy.Endpoint.Url
		category := proxy.Endpoint.Category
		if !IsHandlerEndExist(proxies, id, category) {
			continue
		}

		proxyChain := Chain(proxyChains, proxy.Endpoint)
		if len(proxyChain) == 0 {
			return fmt.Errorf("proxyChain(id: '%s') is 0. Set it by calling service.SetProxyChain()", proxy.Id)
		}

		// if proxy-chain has no route, return an error
		if !IsEndpointExist(proxyChain, proxy.Endpoint) {
			return fmt.Errorf("IsEndpointExist in the proxy chain doesn't have the route endpoint('%s','%s')",
				proxy.Endpoint.Category, proxy.Endpoint.Command)
		}
	}

	return nil
}
