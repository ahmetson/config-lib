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
	ServiceId string `json:"service_id"` // Service id
	HandlerId string `json:"handler_id"` // Handler id
	Command   string `json:"command"`    // Route command
}

type ProxyChain struct {
	Sources     []string `json:"sources,omitempty" yaml:"sources,omitempty"`
	Proxies     []*Proxy `json:"proxies" yaml:"proxies"`
	Destination *Rule    `json:"destination" yaml:"destination"` // only shown for a proxy
}

// The anyToStringSlice function converts the interface to the slice of strings.
// The parameter could be a string or []string.
//
// Returns nil if the raw parameter is not string or []string.
func anyToStringSlice(raw interface{}) []string {
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
		Urls:             []string{},
		ExcludedCommands: []string{},
	}

	if len(params) < 2 || len(params) > 3 {
		return nil
	}
	if len(params) == 3 {
		urls := anyToStringSlice(params[0])
		if urls == nil {
			return nil
		}
		urls = slices.Compact(urls)
		unit.Urls = make([]string, 0, len(urls))
		unit.Urls = append(unit.Urls, urls...)
	}

	index := len(params) - 2

	categories := anyToStringSlice(params[index])
	if categories == nil {
		return nil
	}

	categories = slices.Compact(categories)
	unit.Categories = make([]string, 0, len(categories))
	unit.Categories = append(unit.Categories, categories...)

	index++

	commands := anyToStringSlice(params[index])
	if commands == nil {
		return nil
	}

	commands = slices.Compact(commands)
	unit.Commands = make([]string, 0, len(commands))
	unit.Commands = append(unit.Commands, commands...)

	return unit
}

// NewHandlerDestination returns a rule for the handler.
// Returns nil if parameters are invalid.
//
// Any parameter could be a string or []string.
func NewHandlerDestination(params ...interface{}) *Rule {
	unit := Rule{
		Urls:             []string{},
		Commands:         []string{},
		ExcludedCommands: []string{},
	}

	if len(params) < 1 || len(params) > 2 {
		return nil
	}

	if len(params) == 2 {
		urls := anyToStringSlice(params[0])
		if urls == nil {
			return nil
		}
		urls = slices.Compact(urls)
		unit.Urls = make([]string, 0, len(urls))
		unit.Urls = append(unit.Urls, urls...)
	}

	index := len(params) - 1

	categories := anyToStringSlice(params[index])
	if categories == nil {
		return nil
	}

	categories = slices.Compact(categories)
	unit.Categories = make([]string, 0, len(categories))
	unit.Categories = append(unit.Categories, categories...)

	return &unit
}

// NewServiceDestination returns a rule for the service.
// Returns nil if parameter is invalid.
//
// A parameter could be a string or []string.
// If no parameter is given, returns an empty rule.
// In that case, set the urls later.
func NewServiceDestination(params ...interface{}) *Rule {
	unit := &Rule{
		Urls:             make([]string, 0, len(params)),
		Categories:       []string{},
		Commands:         []string{},
		ExcludedCommands: []string{},
	}

	if len(params) == 0 {
		return unit
	}

	if len(params) != 1 {
		return nil
	}

	urls := anyToStringSlice(params[0])
	if urls == nil {
		return nil
	}
	urls = slices.Compact(urls)
	unit.Urls = append(unit.Urls, urls...)

	return unit
}

// IsService returns true for the service destination.
// The rule is a service destination if Urls are given, but not Categories and Commands.
func (unit *Rule) IsService() bool {
	if unit == nil {
		return false
	}
	return len(unit.Categories) == 0 && len(unit.Commands) == 0 && len(unit.Urls) > 0
}

// IsHandler returns true for the handler destination.
// The rule is a handler destination if Urls and Categories are given, but not Commands.
func (unit *Rule) IsHandler() bool {
	if unit == nil {
		return false
	}
	return len(unit.Urls) > 0 && len(unit.Categories) > 0 && len(unit.Commands) == 0
}

// IsRoute returns true if for the route destination.
// The rule is a route destination if Urls, Categories and Commands are given
func (unit *Rule) IsRoute() bool {
	if unit == nil {
		return false
	}
	return len(unit.Urls) > 0 && len(unit.Categories) > 0 && len(unit.Commands) > 0
}

// IsValid returns true for a valid destination.
// The rule is considered valid if it's for route or handler or service.
//
// The empty rule is not a valid rule.
func (unit *Rule) IsValid() bool {
	return unit != nil && !unit.IsEmpty() && !unit.IsEmptyCommands() &&
		(unit.IsService() || unit.IsHandler() || unit.IsRoute())
}

// IsEmpty returns true if no fields are given.
//
// One way to create an empty parameter is to call NewServiceDestination() without any argument.
func (unit *Rule) IsEmpty() bool {
	return unit != nil && (len(unit.Urls)+
		len(unit.Categories)+
		len(unit.Commands) == 0)
}

// IsEmptyCommands returns true if all Commands are in the ExcludedCommands
func (unit *Rule) IsEmptyCommands() bool {
	if unit == nil {
		return false
	}

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
	if unit == nil {
		return unit
	}

	for _, command := range commands {
		if slices.Contains(unit.ExcludedCommands, command) {
			continue
		}
		unit.ExcludedCommands = append(unit.ExcludedCommands, command)
	}
	return unit
}

// IsEqualRule returns true if the fields of both structs match.
func IsEqualRule(first *Rule, second *Rule) bool {
	if first == nil || second == nil {
		return false
	}

	if len(first.Urls) != len(second.Urls) ||
		len(first.Categories) != len(second.Categories) ||
		len(first.Commands) != len(second.Commands) ||
		len(first.ExcludedCommands) != len(second.ExcludedCommands) {
		return false
	}

	for _, url := range first.Urls {
		if !slices.Contains(second.Urls, url) {
			return false
		}
	}

	for _, cmd := range first.Commands {
		if !slices.Contains(second.Commands, cmd) {
			return false
		}
	}

	for _, cat := range first.Categories {
		if !slices.Contains(second.Categories, cat) {
			return false
		}
	}

	for _, cmd := range first.ExcludedCommands {
		if !slices.Contains(second.ExcludedCommands, cmd) {
			return false
		}
	}

	return true
}

// IsEqualProxy returns true if the proxies match.
func IsEqualProxy(first *Proxy, second *Proxy) bool {
	if first == nil || second == nil {
		return false
	}

	return first.Id == second.Id && first.Url == second.Url && first.Category == second.Category
}

//
// Proxy functions and methods
//

// IsValid returns true if all fields are set
func (proxy *Proxy) IsValid() bool {
	if proxy == nil {
		return false
	}
	return len(proxy.Url) > 0 && len(proxy.Id) > 0 && len(proxy.Category) > 0
}

//
// ProxyChain functions and methods
//

func IsProxyExist(proxies []*Proxy, id string) bool {
	return slices.ContainsFunc(proxies, func(el *Proxy) bool {
		return el.Id == id
	})
}

// IsStringSliceValid returns true if all elements are unique.
// Valid string slice must not be nil and elements must not be empty.
// The empty slice is considered as valid.
func IsStringSliceValid(haystack []string) bool {
	if haystack == nil {
		return false
	} else if len(haystack) == 0 {
		return true
	}

	if len(haystack) > 0 {
		for i, needle := range haystack {
			if len(needle) == 0 {
				return false
			}

			for j, element := range haystack {
				if i == j {
					continue
				}

				// duplicate
				if needle == element {
					return false
				}
			}
		}
	}

	return true
}

// IsProxiesValid returns true if the Proxies field has no duplicate elements.
// Proxies are compared against their ids.
func (proxyChain *ProxyChain) IsProxiesValid() bool {
	if proxyChain.Proxies == nil || len(proxyChain.Proxies) == 0 {
		return false
	}

	for i, needle := range proxyChain.Proxies {
		if !needle.IsValid() {
			return false
		}

		for j, proxy := range proxyChain.Proxies {
			if j == i {
				continue
			}

			if needle.Id == proxy.Id {
				return false
			}
		}
	}

	return true
}

// IsValid returns true if the proxy chain is valid.
// It's counted as valid if it doesn't have a duplicate values.
// Any nil field makes the proxy chain as invalid.
func (proxyChain *ProxyChain) IsValid() bool {
	return proxyChain != nil && proxyChain.Destination.IsValid() &&
		proxyChain.IsProxiesValid() &&
		IsStringSliceValid(proxyChain.Sources)
}

// IsRuleExist returns true if there is a proxy chain that matches to the rule
func IsRuleExist(proxyChains []*ProxyChain, rule *Rule) bool {
	return slices.ContainsFunc(proxyChains, func(proxyChain *ProxyChain) bool {
		return IsEqualRule(proxyChain.Destination, rule)
	})
}

// ProxyChainByRule returns a proxy chain that has the rule
func ProxyChainByRule(proxyChains []*ProxyChain, rule *Rule) *ProxyChain {
	if !rule.IsValid() {
		return nil
	}

	for _, proxyChain := range proxyChains {
		if IsEqualRule(proxyChain.Destination, rule) {
			return proxyChain
		}
	}

	return nil
}

// ProxyChainsByRuleUrl returns a list of proxy chains, where rule has the url.
// Returns empty list if no ulr was found.
//
// todo, it must be only one
func ProxyChainsByRuleUrl(proxyChains []*ProxyChain, url string) []*ProxyChain {
	foundProxyChains := make([]*ProxyChain, 0, len(proxyChains))
	if len(url) == 0 {
		return foundProxyChains
	}

	for _, proxyChain := range proxyChains {
		if slices.Contains(proxyChain.Destination.Urls, url) {
			foundProxyChains = append(foundProxyChains, proxyChain)
		}
	}

	return foundProxyChains
}

// LastProxies returns the list of the proxies from all proxy chains.
// The identical proxies are compacted
func LastProxies(proxyChains []*ProxyChain) []*Proxy {
	proxies := make([]*Proxy, 0, len(proxyChains))

	if len(proxyChains) == 0 {
		return proxies
	}

	for i := range proxyChains {
		if len(proxyChains[i].Proxies) == 0 {
			continue
		}

		lastProxy := len(proxyChains[i].Proxies) - 1

		last := proxyChains[i].Proxies[lastProxy]

		if IsProxyExist(proxies, last.Id) {
			continue
		}

		proxies = append(proxies, last)
	}

	return proxies
}
