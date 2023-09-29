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
}

// Make sure that Account is set to five
// before each test
func (test *TestProxySuite) SetupTest() {
	test.urls = []string{"url_1", "url_2"}
	test.categories = []string{"category_1", "category_2"}
	test.commands = []string{"command_1", "command_2"}
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

	// creating a destination with two parameters but invalid data must fail
	destinations = NewDestination(1, "")
	s().Nil(destinations)

	// creating a destination with three parameters but an invalid data type must fail
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

func TestProxy(t *testing.T) {
	suite.Run(t, new(TestProxySuite))
}
