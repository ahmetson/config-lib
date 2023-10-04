package service

import (
	clientConfig "github.com/ahmetson/client-lib/config"
	handlerConfig "github.com/ahmetson/handler-lib/config"
	"github.com/pebbe/zmq4"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing orchestra
type TestServiceSuite struct {
	suite.Suite

	urls       []string
	categories []string
	commands   []string
	extIds     []string

	validProxy   *Proxy
	validProxy2  *Proxy
	invalidProxy *Proxy

	serviceId          string
	managerId          string
	service            *Service
	handlerOfCategory  *handlerConfig.Handler
	handler2OfCategory *handlerConfig.Handler
	handlerOfCategory2 *handlerConfig.Handler

	extensions []*clientConfig.Client
}

// Make sure that Account is set to five
// before each test
func (test *TestServiceSuite) SetupTest() {
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

	test.serviceId = "service_id"
	test.managerId = "service_id_manager"
	test.service = New(test.serviceId, test.urls[0], IndependentType, nil)
	test.handlerOfCategory = handlerConfig.NewInternalHandler(handlerConfig.ReplierType, test.categories[0])
	test.handler2OfCategory = handlerConfig.NewInternalHandler(handlerConfig.ReplierType, test.categories[0])
	test.handler2OfCategory.Id = "second_id"
	test.handlerOfCategory2 = handlerConfig.NewInternalHandler(handlerConfig.ReplierType, test.categories[1])

	test.extIds = []string{"ext_1", "ext_2"}
	extConfig := clientConfig.New(test.urls[0], test.extIds[0], 0, zmq4.REP)
	extConfig.UrlFunc(clientConfig.Url)
	extConfig2 := clientConfig.New(test.urls[1], test.extIds[1], 0, zmq4.REP)
	extConfig2.UrlFunc(clientConfig.Url)
	test.extensions = []*clientConfig.Client{extConfig, extConfig2}
}

// Test_10_ManagerId tests generation of the manager id with ManagerId function
func (test *TestServiceSuite) Test_10_ManagerId() {
	s := test.Require

	managerId := ManagerId("")
	s().Empty(managerId)

	managerId = ManagerId(test.serviceId)
	s().Equal(test.managerId, managerId)
}

// Test_11_NewManager tests NewManager
func (test *TestServiceSuite) Test_11_NewManager() {
	s := test.Require

	// empty arguments are not allowed
	_, err := NewManager("", test.urls[0])
	s().Error(err)
	_, err = NewManager(test.serviceId, "")
	s().Error(err)
	manager, err := NewManager(test.serviceId, test.urls[0])
	s().NoError(err)
	s().NotNil(manager)
	s().Equal(zmq4.REP, manager.TargetType)
	s().Equal(test.urls[0], manager.ServiceUrl)
	s().Equal(test.managerId, manager.Id)
	s().NotZero(manager.Port)
	s().NotEmpty(manager.Url())
}

// Test_12_Service_ValidateTypes tests validating types
func (test *TestServiceSuite) Test_12_Service_ValidateTypes() {
	s := test.Require

	invalidHandler := handlerConfig.NewInternalHandler("invalid_handler_type", "category")

	// test the service type
	generatedService := &Service{
		Type:     "the_invalid_type",
		Handlers: []*handlerConfig.Handler{test.handlerOfCategory},
	}

	err := generatedService.ValidateTypes()
	s().Error(err)

	// valid both handler and service types
	generatedService.Type = IndependentType
	err = generatedService.ValidateTypes()
	s().NoError(err)

	// validate the handler type
	generatedService.Handlers = []*handlerConfig.Handler{invalidHandler}
	err = generatedService.ValidateTypes()
	s().Error(err)
}

// Test_13_Service_Handler tests Service.HandlerByCategory and Service.HandlersByCategory
func (test *TestServiceSuite) Test_13_Service_Handler() {
	s := test.Require

	nonExistCategory := "not_found_category"
	handlers := []*handlerConfig.Handler{test.handlerOfCategory, test.handler2OfCategory, test.handlerOfCategory2}
	test.service.Handlers = handlers

	// empty category must fail
	_, err := test.service.HandlerByCategory("")
	s().Error(err)
	_, err = test.service.HandlersByCategory("")
	s().Error(err)

	// if handler of category not found, service must return error
	_, err = test.service.HandlerByCategory(nonExistCategory)
	s().Error(err)
	_, err = test.service.HandlersByCategory(nonExistCategory)
	s().Error(err)

	// find the handler
	foundHandler, err := test.service.HandlerByCategory(test.categories[0])
	s().NoError(err)
	s().NotNil(foundHandler)
	s().Equal(test.handlerOfCategory.Id, foundHandler.Id)

	foundHandler, err = test.service.HandlerByCategory(test.categories[1])
	s().NoError(err)
	s().NotNil(foundHandler)
	s().Equal(test.handlerOfCategory2.Id, foundHandler.Id)

	// find the handlers
	foundHandlers, err := test.service.HandlersByCategory(test.categories[0])
	s().NoError(err)
	s().Len(foundHandlers, 2)
	s().Equal(test.handlerOfCategory.Id, foundHandlers[0].Id)
	s().Equal(test.handler2OfCategory.Id, foundHandlers[1].Id)

	foundHandlers, err = test.service.HandlersByCategory(test.categories[1])
	s().NoError(err)
	s().Len(foundHandlers, 1)
	s().Equal(test.handlerOfCategory2.Id, foundHandlers[0].Id)
}

// Test_14_Service_SetHandler tests Service.SetHandler
func (test *TestServiceSuite) Test_14_Service_SetHandler() {
	s := test.Require

	categoryName := "new_category"

	// in the beginning, the handlers are empty
	s().Len(test.service.Handlers, 0)

	// Setting a handler to a nil must be skipped
	var serviceConfig *Service
	serviceConfig.SetHandler(test.handlerOfCategory)

	// setting the handler must be valid
	s().Len(test.service.Handlers, 0)
	test.service.SetHandler(test.handlerOfCategory)
	s().Len(test.service.Handlers, 1)
	s().Equal(test.categories[0], test.service.Handlers[0].Category)

	test.service.SetHandler(test.handlerOfCategory2)
	s().Len(test.service.Handlers, 2)
	s().Equal(test.categories[0], test.service.Handlers[0].Category) // no change in the handler order
	s().Equal(test.categories[1], test.service.Handlers[1].Category)

	// duplicate handler must be over-writing the handler
	test.handlerOfCategory.Category = categoryName
	s().Len(test.service.Handlers, 2)
	s().Equal(categoryName, test.service.Handlers[0].Category)
	s().Equal(test.categories[1], test.service.Handlers[1].Category)
}

// Test_15_Service_Extension test Service.SetExtension and Service.ExtensionByUrl
func (test *TestServiceSuite) Test_15_Service_Extension() {
	s := test.Require

	newId := "new_ext_id"

	// extensions must be not set
	s().Len(test.service.Extensions, 0)

	// no extension found
	ext := test.service.ExtensionByUrl(test.urls[0])
	s().Nil(ext)

	// set the extension
	test.service.SetExtension(test.extensions[0])
	s().Len(test.service.Extensions, 1)

	// check the set extension
	ext = test.service.ExtensionByUrl(test.urls[0])
	s().NotNil(ext)
	s().Equal(test.urls[0], ext.ServiceUrl)
	s().Equal(test.extIds[0], ext.Id)

	// set a new extension
	test.service.SetExtension(test.extensions[1])
	s().Len(test.service.Extensions, 2)

	//
	// both extensions must be returned
	//

	ext = test.service.ExtensionByUrl(test.urls[0])
	s().NotNil(ext)
	s().Equal(test.urls[0], ext.ServiceUrl)
	s().Equal(test.extIds[0], ext.Id)

	ext = test.service.ExtensionByUrl(test.urls[1])
	s().NotNil(ext)
	s().Equal(test.urls[1], ext.ServiceUrl)
	s().Equal(test.extIds[1], ext.Id)

	// setting a duplicate extension must over-write the previous extension
	test.extensions[0].Id = newId
	test.service.SetExtension(test.extensions[0])
	s().Len(test.service.Extensions, 2)

	// validate over-writing
	ext = test.service.ExtensionByUrl(test.urls[0])
	s().NotNil(ext)
	s().Equal(test.urls[0], ext.ServiceUrl)
	s().Equal(newId, ext.Id)
}

// Test_16_IsSourceExist tests the IsSourceExist
func (test *TestServiceSuite) Test_16_Service_SetSource() {
	s := test.Require

	rule := NewServiceDestination(test.urls[0])
	proxy1Url := "u"
	proxy1 := &Proxy{
		Id:       "i",
		Url:      proxy1Url,
		Category: "c",
	}
	proxy2 := &Proxy{
		Id:       "i2",
		Url:      "u2",
		Category: "c2",
	}
	managerClient := clientConfig.New(proxy1Url, proxy1.Id, 0, zmq4.REP)
	client1 := clientConfig.New(proxy1Url, proxy1.Id, 0, zmq4.REP)
	client2 := clientConfig.New(proxy1Url, proxy1.Id, 0, zmq4.REP)

	sourceService1 := &SourceService{
		Proxy:   proxy1,
		Manager: managerClient,
		Clients: []*clientConfig.Client{client1, client2},
	}
	sourceService2 := &SourceService{
		Proxy:   proxy2,
		Manager: managerClient,
		Clients: []*clientConfig.Client{client1, client2},
	}

	// by default no sources
	s().Len(test.service.Sources, 0)

	// trying to set the defaults must fail
	updated := test.service.SetServiceSource(nil, nil)
	s().False(updated)

	updated = test.service.SetServiceSource(rule, nil)
	s().False(updated)

	updated = test.service.SetServiceSource(nil, sourceService1)
	s().False(updated)

	// setting a fresh source
	updated = test.service.SetServiceSource(rule, sourceService1)
	s().True(updated)
	s().Len(test.service.Sources, 1)
	s().Len(test.service.Sources[0].Proxies, 1)

	// adding a duplicate must fail
	updated = test.service.SetServiceSource(rule, sourceService1)
	s().False(updated)

	// setting the second rule must succeed
	updated = test.service.SetServiceSource(rule, sourceService2)
	s().True(updated)
	s().Len(test.service.Sources, 1)
	s().Len(test.service.Sources[0].Proxies, 2)

	// non equal source services must be set
	proxy1Updated := &Proxy{
		Id:       "i",
		Url:      "u2",
		Category: "c2",
	}
	sourceService1Updated := &SourceService{
		Proxy:   proxy1Updated,
		Manager: managerClient,
		Clients: []*clientConfig.Client{client1, client2},
	}
	updated = test.service.SetServiceSource(rule, sourceService1Updated)
	s().True(updated)
	s().Len(test.service.Sources, 1)
	s().Len(test.service.Sources[0].Proxies, 2)

	// adding a new rule must succeed
	secondRule := NewHandlerDestination(test.urls[0], "category")
	updated = test.service.SetServiceSource(secondRule, sourceService1Updated)
	s().True(updated)
	s().Len(test.service.Sources, 2)
	s().Len(test.service.Sources[0].Proxies, 2)
	s().Len(test.service.Sources[1].Proxies, 1)
}

func TestService(t *testing.T) {
	suite.Run(t, new(TestServiceSuite))
}
