package service

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestProxySuite struct {
	suite.Suite

	urls       []string
	categories []string
	commands   []string

	validProxy   *Proxy
	validProxy2  *Proxy
	invalidProxy *Proxy
}

// Make sure that Account is set to five
// before each test
func (test *TestProxySuite) SetupTest() {
	test.urls = []string{"url_1", "url_2"}
	test.categories = []string{"category_1", "category_2"}
	test.commands = []string{"command_1", "command_2"}

	test.validProxy = &Proxy{
		Local:    &Local{},
		Id:       "i",
		Url:      "u",
		Category: "cat",
	}
	test.validProxy2 = &Proxy{
		Local:    &Local{},
		Id:       test.validProxy.Id + "2",
		Url:      test.validProxy.Url + "2",
		Category: test.validProxy.Category + "2",
	}
	test.invalidProxy = &Proxy{
		Id:       "",
		Url:      "u",
		Category: "c",
	}
}

// Test_10_NewDestination tests NewDestination
func (test *TestProxySuite) Test_10_NewDestination() {
	s := test.Require

	// creating a destination without any parameter must fail
	destinations := NewDestination()
	s().Nil(destinations)

	// creating a destination with more than 3 parameters must fail
	destinations = NewDestination("", "", "", "")
	s().Nil(destinations)

	//
	// creating a destination with two parameters but invalid data must fail
	//

	destinations = NewDestination(1, "")
	s().Nil(destinations)
	destinations = NewDestination("", 1)
	s().Nil(destinations)
	s().Nil(NewDestination(1, "", ""))
	s().Nil(NewDestination("", 1, ""))

	destinations = NewDestination("hello", []string{"yes", "data"}, 2)
	s().Nil(destinations)

	//
	// creating a destination with two parameters
	//

	// all are scalar type
	destinations = NewDestination(test.categories[0], test.commands[0])
	s().NotNil(destinations)
	s().Empty(destinations.Urls)
	s().Len(destinations.Categories, 1)
	s().Len(destinations.Commands, 1)
	s().Equal(test.categories[0], destinations.Categories[0])
	s().Equal(test.commands[0], destinations.Commands[0])

	destinations = NewDestination(test.categories, test.commands[0])
	s().NotNil(destinations)
	s().Empty(destinations.Urls)
	s().Len(destinations.Categories, 2)
	s().Len(destinations.Commands, 1)
	s().EqualValues(test.categories, destinations.Categories)
	s().Equal(test.commands[0], destinations.Commands[0])

	destinations = NewDestination(test.categories[0], test.commands)
	s().NotNil(destinations)
	s().Empty(destinations.Urls)
	s().Len(destinations.Categories, 1)
	s().Len(destinations.Commands, 2)
	s().Equal(test.categories[0], destinations.Categories[0])
	s().EqualValues(test.commands, destinations.Commands)

	destinations = NewDestination(test.categories, test.commands)
	s().NotNil(destinations)
	s().Empty(destinations.Urls)
	s().Len(destinations.Categories, 2)
	s().Len(destinations.Commands, 2)
	s().EqualValues(test.categories, destinations.Categories)
	s().EqualValues(test.commands, destinations.Commands)

	//
	// Testing with three parameters
	//
	destinations = NewDestination(test.urls[0], test.categories[0], test.commands[0])
	s().NotNil(destinations)
	s().Len(destinations.Urls, 1)
	s().Len(destinations.Categories, 1)
	s().Len(destinations.Commands, 1)
	s().Equal(test.urls[0], destinations.Urls[0])
	s().Equal(test.categories[0], destinations.Categories[0])
	s().Equal(test.commands[0], destinations.Commands[0])

	destinations = NewDestination(test.urls, test.categories, test.commands[0])
	s().NotNil(destinations)
	s().Len(destinations.Urls, 2)
	s().Len(destinations.Categories, 2)
	s().Len(destinations.Commands, 1)
	s().EqualValues(test.urls, destinations.Urls)
	s().EqualValues(test.categories, destinations.Categories)
	s().Equal(test.commands[0], destinations.Commands[0])
}

// Test_11_NewHandlerDestination tests NewHandlerDestination
func (test *TestProxySuite) Test_11_NewHandlerDestination() {
	s := test.Require

	// creating a destination without any parameter must fail
	destinations := NewHandlerDestination()
	s().Nil(destinations)

	// creating a destination with more than 2 parameters must fail
	destinations = NewHandlerDestination("", "", "")
	s().Nil(destinations)

	// creating a destination with invalid data type must fail
	destinations = NewHandlerDestination(1)
	s().Nil(destinations)
	s().Nil(NewHandlerDestination(1, "category"))

	// creating a destination with two parameters but an invalid data type must fail
	destinations = NewHandlerDestination([]string{"bla", "bla"}, 2)
	s().Nil(destinations)

	//
	// creating a destination with one parameter
	//

	// all are scalar type
	destinations = NewHandlerDestination(test.categories[0])
	s().NotNil(destinations)
	s().Empty(destinations.Urls)
	s().Len(destinations.Categories, 1)
	s().Empty(destinations.Commands)
	s().Equal(test.categories[0], destinations.Categories[0])

	destinations = NewHandlerDestination(test.categories)
	s().NotNil(destinations)
	s().Empty(destinations.Urls)
	s().Len(destinations.Categories, 2)
	s().Empty(destinations.Commands)
	s().EqualValues(test.categories, destinations.Categories)

	//
	// Testing with two parameters
	//
	destinations = NewHandlerDestination(test.urls[0], test.categories[0])
	s().NotNil(destinations)
	s().Len(destinations.Urls, 1)
	s().Len(destinations.Categories, 1)
	s().Empty(destinations.Commands)
	s().Equal(test.urls[0], destinations.Urls[0])
	s().Equal(test.categories[0], destinations.Categories[0])

	destinations = NewHandlerDestination(test.urls, test.categories)
	s().NotNil(destinations)
	s().Len(destinations.Urls, 2)
	s().Len(destinations.Categories, 2)
	s().Empty(destinations.Commands)
	s().EqualValues(test.urls, destinations.Urls)
	s().EqualValues(test.categories, destinations.Categories)
}

// Test_12_NewServiceDestination tests NewServiceDestination
func (test *TestProxySuite) Test_12_NewServiceDestination() {
	s := test.Require

	// creating a destination with invalid data must fail
	destinations := NewServiceDestination()
	s().NotNil(destinations)

	// creating a destination with invalid data type must fail
	destinations = NewServiceDestination(1)
	s().Nil(destinations)

	// creating a destination with more than 1 parameter must fail
	destinations = NewServiceDestination("", "")

	// all are scalar type
	destinations = NewServiceDestination(test.urls[0])
	s().NotNil(destinations)
	s().Len(destinations.Urls, 1)
	s().Empty(destinations.Categories)
	s().Empty(destinations.Commands)
	s().Equal(test.urls[0], destinations.Urls[0])

	destinations = NewServiceDestination(test.urls)
	s().NotNil(destinations)
	s().Len(destinations.Urls, 2)
	s().Empty(destinations.Categories)
	s().Empty(destinations.Commands)
	s().EqualValues(test.urls, destinations.Urls)
}

// Test_13_Is tests IsService, IsHandler and IsRoute and IsValid.
// Tests with a Rule derived from NewDestination.
// Tests with a Rule derived from NewHandlerDestination.
// Tests with a Rule derived from NewServiceDestination.
//
// Tests IsValid and IsEmpty for different combinations.
// Tests Is* on the fly.
func (test *TestProxySuite) Test_13_Is() {
	s := test.Require

	var rule *Rule = nil
	s().False(rule.IsService())
	s().False(rule.IsHandler())
	s().False(rule.IsRoute())
	s().False(rule.IsEmptyCommands())
	s().Nil(rule.ExcludeCommands("items"))

	// handler destination
	destinations := NewHandlerDestination(test.categories)
	s().False(destinations.IsService()) // categories given
	s().False(destinations.IsValid())   // missing urls
	s().False(destinations.IsEmpty())
	s().False(destinations.IsHandler()) // missing urls
	s().False(destinations.IsRoute())   // missing commands

	destinations = NewHandlerDestination(test.urls, test.categories)
	s().False(destinations.IsService()) // categories exist
	s().True(destinations.IsHandler())  // valid as urls and categories are given
	s().True(destinations.IsValid())    // valid since IsHandler is valid
	s().False(destinations.IsRoute())   // missing commands
	s().False(destinations.IsEmpty())

	// route destination
	destinations = NewDestination(test.categories, test.commands)
	s().False(destinations.IsService()) // categories and commands are given
	s().False(destinations.IsHandler()) // commands given
	s().False(destinations.IsRoute())   // missing urls
	s().False(destinations.IsValid())   // missing urls
	s().False(destinations.IsEmpty())

	destinations = NewDestination(test.urls, test.categories, test.commands)
	s().False(destinations.IsService()) // categories and commands given
	s().False(destinations.IsHandler()) // commands given
	s().True(destinations.IsRoute())    // valid as urls, categories and commands given
	s().True(destinations.IsValid())    // valid since IsRoute is valid
	s().False(destinations.IsEmpty())

	// service destination
	destinations = NewServiceDestination()
	s().False(destinations.IsService()) // missing urls
	s().False(destinations.IsHandler()) // missing categories
	s().False(destinations.IsRoute())   // missing categories and commands
	s().False(destinations.IsValid())   // empty route is not valid
	s().True(destinations.IsEmpty())    // the `NewServiceDestination()` only way to create an empty route by functions

	destinations = NewServiceDestination(test.urls)
	s().True(destinations.IsService())  // urls given
	s().False(destinations.IsHandler()) // missing categories
	s().False(destinations.IsRoute())   // missing categories and commands
	s().True(destinations.IsValid())    // valid since IsService is valid
	s().False(destinations.IsEmpty())   // urls given
}

// Test_14_ExcludeCommands tests excluding commands
func (test *TestProxySuite) Test_14_ExcludeCommands() {
	s := test.Require

	destinations := NewDestination(test.urls, test.categories, test.commands)
	s().True(destinations.IsValid())
	s().False(destinations.IsEmptyCommands())

	// Excluding already existing commands must fail
	destinations.ExcludeCommands(test.commands...)
	s().Equal(test.commands, destinations.ExcludedCommands)
	s().False(destinations.IsValid()) // IsEmptyCommands failed
	s().True(destinations.IsEmptyCommands())

	// Only one command is excluded, it must return empty
	destinations.ExcludedCommands = []string{test.commands[0]}
	s().True(destinations.IsValid())
	s().False(destinations.IsEmptyCommands())

	// We set manually test.commands[0] as the excluded command
	// Trying to exclude already existing command must be skipped
	destinations.ExcludeCommands(test.commands[0])
	s().Len(destinations.ExcludedCommands, 1)
}

// Test_15_Proxy_IsValid tests Proxy.IsValid method
func (test *TestProxySuite) Test_15_Proxy_IsValid() {
	s := test.Require

	id := "proxy_id"
	url := "github.com/ahmetson/proxy"
	category := "category"

	var proxy *Proxy = nil
	s().False(proxy.IsValid())

	proxy = &Proxy{}
	s().False(proxy.IsValid())

	proxy.Id = id
	s().False(proxy.IsValid())

	proxy.Url = url
	s().False(proxy.IsValid())

	proxy.Category = category
	s().False(proxy.IsValid())

	proxy.Local = &Local{}
	s().True(proxy.IsValid())
}

// Test_16_IsProxiesValid tests ProxyChain.IsProxiesValid
func (test *TestProxySuite) Test_16_IsProxiesValid() {
	s := test.Require

	proxyChain := &ProxyChain{}
	// no proxies must be unique
	s().False(proxyChain.IsProxiesValid())

	// empty proxies must be unique
	proxyChain.Proxies = make([]*Proxy, 0)
	s().False(proxyChain.IsProxiesValid())

	// duplicate proxies must fail
	proxyChain.Proxies = []*Proxy{test.validProxy, test.validProxy}
	s().False(proxyChain.IsProxiesValid())

	// proxy is invalid
	proxyChain.Proxies = []*Proxy{test.invalidProxy}
	s().False(proxyChain.IsProxiesValid())

	// list of unique, valid proxies must be valid
	proxyChain.Proxies = []*Proxy{test.validProxy, test.validProxy2}
	s().True(proxyChain.IsProxiesValid())
}

// Test_17_ProxyChain_IsValid tests ProxyChain.IsValid and IsStringSliceValid
func (test *TestProxySuite) Test_17_ProxyChain_IsValid() {
	s := test.Require

	url := "url"
	url2 := "url_2"
	validProxies := []*Proxy{test.validProxy, test.validProxy2}
	validSources := []string{url, url2}

	//
	// First testing against nil values
	//
	var proxyChain *ProxyChain
	s().False(proxyChain.IsValid()) // missing Sources, Proxies, and Destination

	proxyChain = &ProxyChain{}
	s().False(proxyChain.IsValid()) // missing Sources, Proxies, and Destination

	proxyChain.Proxies = validProxies
	s().False(proxyChain.IsValid()) // missing Sources and Destination

	proxyChain.Sources = validSources
	s().False(proxyChain.IsValid()) // missing Destination

	proxyChain.Destination = NewServiceDestination(url)
	s().True(proxyChain.IsValid())

	//
	// testing for invalid fields
	//

	proxyChain.Proxies = []*Proxy{test.validProxy, test.invalidProxy}
	s().False(proxyChain.IsValid()) // The Proxies field invalid
	proxyChain.Proxies = validProxies

	proxyChain.Sources = []string{url, url} // duplicate values
	s().False(proxyChain.IsValid())         // The Sources field invalid
	proxyChain.Sources = validSources

	proxyChain.Destination = NewServiceDestination()
	s().False(proxyChain.IsValid()) // Destination field invalid
	proxyChain.Destination = NewServiceDestination(url)
	s().True(proxyChain.IsValid())

	//
	// Testing IsStringSliceValid
	//
	proxyChain.Sources = nil
	s().False(proxyChain.IsValid())

	// empty sources must return true
	proxyChain.Sources = []string{}
	s().True(proxyChain.IsValid())

	// empty source element must be not valid
	proxyChain.Sources = []string{url, ""}
	s().False(proxyChain.IsValid())
}

// Test_18_IsEqualRule test IsEqualRule
func (test *TestProxySuite) Test_18_IsEqualRule() {
	s := test.Require

	first := NewServiceDestination(test.urls[0])
	second := NewServiceDestination(test.urls[1])

	// The nil values must return false
	s().False(IsEqualRule(nil, second))
	s().False(IsEqualRule(first, nil))
	s().False(IsEqualRule(nil, nil))

	// The length of categories mismatch
	second = NewHandlerDestination(test.urls, test.categories)
	s().False(IsEqualRule(first, second))

	// The both are valid
	second = NewServiceDestination(test.urls[0])
	s().True(IsEqualRule(first, second))

	// Testing with invalid service rule
	second = NewServiceDestination(test.urls[1])
	s().False(IsEqualRule(first, second))

	// Testing the handler rule
	first = NewHandlerDestination(test.urls, test.categories)
	second = NewHandlerDestination(test.urls, test.categories)
	s().True(IsEqualRule(first, second))

	// Testing the invalid handler rule
	second.Categories[0] = "non_existing"
	s().False(IsEqualRule(first, second))

	// Testing the route rule
	first = NewDestination(test.urls, test.categories, test.commands)
	second = NewDestination(test.urls, test.categories, test.commands)
	s().True(IsEqualRule(first, second))

	// Testing the invalid command
	second.Commands[0] = "non_existing"
	s().False(IsEqualRule(first, second))

	// Testing with excluded commands
	first = NewHandlerDestination(test.urls, test.categories)
	second = NewHandlerDestination(test.urls, test.categories)
	first.ExcludeCommands(test.commands[0])
	second.ExcludeCommands(test.commands[0])
	s().True(IsEqualRule(first, second))

	// Testing with different excluding commands
	second.ExcludedCommands[0] = "non_existing"
	s().False(IsEqualRule(first, second))
}

// Test_19_IsEqualProxy test IsEqualProxy
func (test *TestProxySuite) Test_19_IsEqualProxy() {
	s := test.Require

	id := "id"
	url := "url"
	category := "category"

	local1 := &Local{}
	local2 := &Local{}
	first := &Proxy{local1, id, url, category}
	second := &Proxy{local2, id, url, category}

	// identical
	s().True(IsEqualProxy(first, second))

	// The nil values must return false
	s().False(IsEqualProxy(nil, second))
	s().False(IsEqualProxy(first, nil))
	s().False(IsEqualProxy(nil, nil))

	// Invalid id
	second.Id = ""
	s().False(IsEqualProxy(first, second))
	second.Id = id

	// Url
	second.Url = "url_2"
	s().False(IsEqualProxy(first, second))
	second.Url = url

	// Category
	second.Category = "2"
	s().False(IsEqualProxy(first, second))
}

// Test_20_IsProxyExist test IsProxyExist
func (test *TestProxySuite) Test_20_IsProxyExist() {
	s := test.Require

	invalidId := "non_existing"
	proxies := []*Proxy{test.validProxy, test.validProxy2}

	// not exists
	s().False(IsProxyExist(nil, invalidId))
	s().False(IsProxyExist(proxies, invalidId))
	s().True(IsProxyExist(proxies, test.validProxy.Id))
}

// Test_21_ProxyChainByRule test ProxyChainByRule and ProxyChainsByRuleUrl
func (test *TestProxySuite) Test_21_ProxyChainByRule() {
	s := test.Require

	invalidUrl := "non_existing"

	invalidRule := NewServiceDestination(invalidUrl)
	validRule := NewServiceDestination(test.urls[0])

	proxies := []*Proxy{test.validProxy, test.validProxy2}
	proxyChain1 := &ProxyChain{
		Sources:     []string{},
		Proxies:     proxies,
		Destination: NewServiceDestination(test.urls[0]),
	}
	proxyChain2 := &ProxyChain{
		Sources:     []string{},
		Proxies:     proxies,
		Destination: NewServiceDestination(test.urls[1]),
	}
	proxyChains := []*ProxyChain{proxyChain1, proxyChain2}

	// the rule must be valid
	invalidRule = NewServiceDestination() // empty rule is not valid
	s().Nil(ProxyChainByRule(proxyChains, invalidRule))

	// the rule that is not in the proxy chains
	invalidRule = NewServiceDestination(invalidUrl)
	s().Nil(ProxyChainByRule(proxyChains, invalidRule))

	// the proxy chain exists
	s().NotNil(ProxyChainByRule(proxyChains, validRule))

	// the nil proxy chain must be false
	s().Nil(ProxyChainByRule(nil, validRule))
}

// Test_22_LastProxies test LastProxies
func (test *TestProxySuite) Test_22_LastProxies() {
	s := test.Require

	proxies := []*Proxy{test.validProxy, test.validProxy2}
	proxyChain1 := &ProxyChain{
		Sources:     []string{},
		Proxies:     proxies,
		Destination: NewServiceDestination(test.urls[0]),
	}
	proxyChain2 := &ProxyChain{
		Sources:     []string{},
		Proxies:     proxies,
		Destination: NewServiceDestination(test.urls[1]),
	}
	proxyChains := []*ProxyChain{proxyChain1, proxyChain2}

	// last proxies for a nil must return empty result
	lastProxies := LastProxies(nil)
	s().Len(lastProxies, 0)
	lastProxies = LastProxies([]*ProxyChain{})
	s().Len(lastProxies, 0)

	// must return one proxy since they are identical
	lastProxies = LastProxies(proxyChains)
	s().Len(lastProxies, 1)

	// if the proxies are empty, then this proxy chain is skipped
	proxyChains[0].Proxies = nil
	proxyChains[1].Proxies = nil
	lastProxies = LastProxies(proxyChains)
	s().Len(lastProxies, 0)

	// each proxy chain has a unique proxy
	proxyChains[0].Proxies = proxies
	proxyChains[1].Proxies = []*Proxy{test.validProxy2, test.validProxy}
	lastProxies = LastProxies(proxyChains)
	s().Len(lastProxies, 2)
}

// Test_23_IsRuleExist test IsRuleExist
func (test *TestProxySuite) Test_23_IsRuleExist() {
	s := test.Require

	invalidUrl := "non_existing"

	invalidRule := NewServiceDestination(invalidUrl)
	validRule := NewServiceDestination(test.urls[0])

	proxies := []*Proxy{test.validProxy, test.validProxy2}
	proxyChain1 := &ProxyChain{
		Sources:     []string{},
		Proxies:     proxies,
		Destination: NewServiceDestination(test.urls[0]),
	}
	proxyChain2 := &ProxyChain{
		Sources:     []string{},
		Proxies:     proxies,
		Destination: NewServiceDestination(test.urls[1]),
	}
	proxyChains := []*ProxyChain{proxyChain1, proxyChain2}

	// valid
	exist := IsRuleExist(proxyChains, validRule)
	s().True(exist)

	// the rule doesn't exist
	exist = IsRuleExist(proxyChains, invalidRule)
	s().False(exist)

	// nil values must return false
	exist = IsRuleExist(nil, invalidRule)
	s().False(exist)
	exist = IsRuleExist(proxyChains, nil)
	s().False(exist)
	exist = IsRuleExist(nil, nil)
	s().False(exist)

}

func TestProxy(t *testing.T) {
	suite.Run(t, new(TestProxySuite))
}
