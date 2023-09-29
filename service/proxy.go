package service

import (
	"slices"
)

// Proxy in the service.
//
// If two endpoints lint to the same route,
type Proxy struct {
	Id       string `json:"id" yaml:"id"`
	Url      string `json:"url" yaml:"url"`
	Category string `json:"category" yaml:"category"` // proxy category, example authr, valid, convert
}

// The Rule is the pattern matching rule to find the services, handlers and routes as the proxy destination.
type Rule struct {
	Urls             []string `json:"urls" yaml:"urls"`                           // Service url
	Categories       []string `json:"categories" yaml:"categories"`               // Handler category
	Commands         []string `json:"commands" yaml:"commands"`                   // Route command
	ExcludedCommands []string `json:"excluded_commands" yaml:"excluded_commands"` // Exclude this commands from routing
}

type Unit struct {
	Url      string `json:"url"`      // Service url
	Category string `json:"category"` // Handler category
	Command  string `json:"command"`  // Route command
}

type ProxyChain struct {
	Sources     []string `json:"sources,omitempty" yaml:"sources,omitempty"`
	Proxies     []*Proxy `json:"proxies" yaml:"proxies"`
	Destination *Rule    `json:"destination" yaml:"destination"` // only shown for a proxy
}

// The convertParam function converts the interface to the slice of strings.
// The parameter could be a string or []string.
//
// Returns nil if the raw parameter is not string or []string.
func convertParam(raw interface{}) []string {
	str, ok := raw.(string)
	if ok {
		return []string{str}
	}

	params, ok := raw.([]string)
	if ok {
		return params
	}

	return nil
}

// NewDestination returns a destination rule for routes.
// Returns nil if params are invalid.
//
// Any parameter could be a string or []string.
func NewDestination(params ...interface{}) *Rule {
	unit := &Rule{
		Urls:             make([]string, 0),
		ExcludedCommands: make([]string, 0),
	}

	if len(params) < 2 || len(params) > 3 {
		return nil
	} else if len(params) == 2 {
		categories := convertParam(params[0])
		if categories == nil {
			return nil
		}
		unit.Categories = categories
		commands := convertParam(params[1])
		if commands == nil {
			return nil
		}
		unit.Commands = commands
	} else {
		urls := convertParam(params[0])
		if urls == nil {
			return nil
		}
		unit.Urls = urls
		categories := convertParam(params[1])
		if categories == nil {
			return nil
		}
		unit.Categories = categories
		commands := convertParam(params[2])
		if commands == nil {
			return nil
		}
		unit.Commands = commands
	}

	return unit
}

// NewHandlerDestination returns a rule for the handler.
// Returns nil if parameters are invalid.
//
// Any parameter could be a string or []string.
func NewHandlerDestination(params ...interface{}) *Rule {
	unit := &Rule{
		Urls:             make([]string, 0),
		Commands:         make([]string, 0),
		ExcludedCommands: make([]string, 0),
	}

	if len(params) < 1 || len(params) > 2 {
		return nil
	} else if len(params) == 1 {
		categories := convertParam(params[0])
		if categories == nil {
			return nil
		}
		unit.Categories = categories
	} else {
		urls := convertParam(params[0])
		if urls == nil {
			return nil
		}
		unit.Urls = urls
		categories := convertParam(params[1])
		if categories == nil {
			return nil
		}
		unit.Categories = categories
	}

	return unit
}

// NewServiceDestination returns a rule for the service.
// Returns nil if parameter is invalid.
//
// A parameter could be a string or []string.
// If no parameter is given, returns an empty rule.
// In that case, set the urls later.
func NewServiceDestination(params ...interface{}) *Rule {
	unit := &Rule{
		Categories:       make([]string, 0),
		Commands:         make([]string, 0),
		ExcludedCommands: make([]string, 0),
	}

	if len(params) == 0 {
		unit.Urls = make([]string, 0)
		return unit
	}

	if len(params) != 1 {
		return nil
	}

	unit.Urls = convertParam(params[0])
	if unit.Urls == nil {
		return nil
	}

	return unit
}

// IsService returns true for the service destination.
// The rule is a service destination if Urls are given, but not Categories and Commands.
func (unit *Rule) IsService() bool {
	return len(unit.Categories) == 0 && len(unit.Commands) == 0 && len(unit.Urls) > 0
}

// IsHandler returns true for the handler destination.
// The rule is a handler destination if Urls and Categories are given, but not Commands.
func (unit *Rule) IsHandler() bool {
	return len(unit.Urls) > 0 && len(unit.Categories) > 0 && len(unit.Commands) == 0
}

// IsRoute returns true if for the route destination.
// The rule is a route destination if Urls, Categories and Commands are given
func (unit *Rule) IsRoute() bool {
	return len(unit.Urls) > 0 && len(unit.Categories) > 0 && len(unit.Commands) > 0
}

// IsValid returns true for a valid destination.
// The rule is considered valid if it's for route or handler or service.
//
// The empty rule is not a valid rule.
func (unit *Rule) IsValid() bool {
	return !unit.IsEmpty() && !unit.IsEmptyCommands() &&
		(unit.IsService() || unit.IsHandler() || unit.IsRoute())
}

// IsEmpty returns true if no fields are given.
//
// One way to create an empty parameter is to call NewServiceDestination() without any argument.
func (unit *Rule) IsEmpty() bool {
	return len(unit.Urls)+
		len(unit.Categories)+
		len(unit.Commands) == 0
}

// IsEmptyCommands returns true if all Commands are in the ExcludedCommands
func (unit *Rule) IsEmptyCommands() bool {
	if len(unit.ExcludedCommands) == 0 || len(unit.Categories) == 0 {
		return false
	}

	for _, command := range unit.Commands {
		if slices.Contains(unit.ExcludedCommands, command) {
			continue
		}
		// the command not in the excluded commands list
		return false
	}

	return true
}

// ExcludeCommands adds the list of commands as an exception for proxies
func (unit *Rule) ExcludeCommands(commands ...string) *Rule {
	for _, command := range commands {
		if slices.Contains(unit.ExcludedCommands, command) {
			continue
		}
		unit.ExcludedCommands = append(unit.ExcludedCommands, command)
	}
	return unit
}

func IsEqualRule(first *Rule, second *Rule) bool {
	return first != nil && second != nil
}

func IsProxySet(proxies []*Proxy, id string) bool {
	return slices.ContainsFunc(proxies, func(el *Proxy) bool {
		return el.Id == id
	})
}

// Chain returns the proxy chain for the given endpoint.
func Chain(proxyChains []*ProxyChain, rule *Rule) []*Proxy {
	for _, proxyChain := range proxyChains {
		if proxyChain.Destination == nil {
			continue
		}

		if IsEqualRule(proxyChain.Destination, rule) {
			return proxyChain.Proxies
		}
	}

	return []*Proxy{}
}

// IsHandlerEndExist returns true if there is a proxy that links to the given handler category
func IsHandlerEndExist(proxyChain *ProxyChain, category string) bool {
	return proxyChain.Destination != nil &&
		slices.Contains(proxyChain.Destination.Categories, category)
}

// IsEndpointExist returns true if the given endpoint exists in the proxy list
func IsEndpointExist(proxyChains []*ProxyChain, endpoint *Rule) bool {
	if endpoint == nil {
		return false
	}

	for _, proxyChain := range proxyChains {
		if IsEqualRule(proxyChain.Destination, endpoint) {
			return true
		}
	}

	return false
}

//
//// ValidProxyChain verifies that endpoints are set correctly.
////
//// If the proxy type is route, and there is a handler of the same type, then make sure that
//// there is a proxy chain of the end.
//func ValidProxyChain(proxies []*Proxy, proxyChains []*ProxyChain) error {
//	for _, proxy := range proxies {
//		if proxy.Units == nil {
//			return fmt.Errorf("proxy('%s').Units is nil. Call Proxy.SetEnd", proxy.Id)
//		}
//
//		if proxy.Units.Type != RouteEnd {
//			continue
//		}
//
//		id := proxy.Units.Url
//		category := proxy.Units.Category
//		if !IsHandlerEndExist(proxies, id, category) {
//			continue
//		}
//
//		proxyChain := Chain(proxyChains, proxy.Units)
//		if len(proxyChain) == 0 {
//			return fmt.Errorf("proxyChain(id: '%s') is 0. Set it by calling service.SetProxyChain()", proxy.Id)
//		}
//
//		// if proxy-chain has no route, return an error
//		if !IsEndpointExist(proxyChain, proxy.Units) {
//			return fmt.Errorf("IsEndpointExist in the proxy chain doesn't have the route endpoint('%s','%s')",
//				proxy.Units.Category, proxy.Units.Command)
//		}
//	}
//
//	return nil
//}
