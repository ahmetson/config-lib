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
		Id:       "i",
		Url:      "u",
		Category: "cat",
	}
	test.validProxy2 = &Proxy{
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

// Test_17_ProxyChain_IsValid tests ProxyChain.IsValid
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
}

func TestProxy(t *testing.T) {
	suite.Run(t, new(TestProxySuite))
}
