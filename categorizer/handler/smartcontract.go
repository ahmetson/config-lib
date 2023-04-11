package handler

import (
	"github.com/blocklords/sds/app/command"
	"github.com/blocklords/sds/app/log"
	"github.com/blocklords/sds/app/remote"
	"github.com/blocklords/sds/app/remote/message"
	"github.com/blocklords/sds/app/service"
	blockchain_command "github.com/blocklords/sds/blockchain/handler"
	blockchain_inproc "github.com/blocklords/sds/blockchain/inproc"
	"github.com/blocklords/sds/blockchain/network"
	"github.com/blocklords/sds/categorizer/event"
	"github.com/blocklords/sds/categorizer/smartcontract"
	"github.com/blocklords/sds/common/data_type/key_value"
	"github.com/blocklords/sds/common/smartcontract_key"

	"github.com/blocklords/sds/db"
)

type GetSmartcontractRequest struct {
	Key smartcontract_key.Key
}
type GetSmartcontractReply struct {
	Smartcontract smartcontract.Smartcontract `json:"smartcontract"`
}

type SetSmartcontractRequest struct {
	Smartcontract smartcontract.Smartcontract `json:"smartcontract"`
}
type SetSmartcontractsReply struct{}

type GetSmartcontractsRequest struct{}
type GetSmartcontractsReply struct {
	Smartcontracts []smartcontract.Smartcontract `json:"smartcontracts"`
}
type PushCategorization struct {
	Smartcontracts []smartcontract.Smartcontract `json:"smartcontracts"`
	Logs           []event.Log                   `json:"logs"`
}
type CategorizationReply key_value.KeyValue

// return a categorized smartcontract parameters by network id and smartcontract address
func GetSmartcontract(request message.Request, _ log.Logger, parameters ...interface{}) message.Reply {
	db := parameters[0].(*db.Database)

	key, err := smartcontract_key.NewFromKeyValue(request.Parameters)
	if err != nil {
		return message.Fail("smartcontract_key.NewFromKeyValue: " + err.Error())
	}

	sm, err := smartcontract.Get(db, key)

	if err != nil {
		return message.Fail("smartcontract.Get: " + err.Error())
	}

	reply := GetSmartcontractReply{
		Smartcontract: *sm,
	}

	reply_message, err := command.Reply(reply)
	if err != nil {
		return message.Fail("parse reply: " + err.Error())
	}

	return reply_message

}

// returns all smartcontract categorized smartcontracts
func GetSmartcontracts(_ message.Request, _ log.Logger, parameters ...interface{}) message.Reply {
	db := parameters[0].(*db.Database)
	smartcontracts, err := smartcontract.GetAll(db)
	if err != nil {
		return message.Fail("the database error " + err.Error())
	}

	reply := GetSmartcontractsReply{
		Smartcontracts: smartcontracts,
	}

	reply_message, err := command.Reply(reply)
	if err != nil {
		return message.Fail("parse reply: " + err.Error())
	}

	return reply_message
}

// Register a new smartcontract to categorizer.
func SetSmartcontract(request message.Request, _ log.Logger, parameters ...interface{}) message.Reply {
	db_con := parameters[0].(*db.Database)

	var request_parameters SetSmartcontractRequest
	err := request.Parameters.ToInterface(&request_parameters)
	if err != nil {
		return message.Fail("parsing request parameters: " + err.Error())
	}

	if smartcontract.Exists(db_con, request_parameters.Smartcontract.SmartcontractKey) {
		return message.Fail("the smartcontract already in SDS Categorizer")
	}

	saveErr := smartcontract.Save(db_con, &request_parameters.Smartcontract)
	if saveErr != nil {
		return message.Fail("database: " + saveErr.Error())
	}

	networks, ok := parameters[2].(network.Networks)
	if !ok {
		return message.Fail("no networks were given")
	}
	if !networks.Exist(request_parameters.Smartcontract.SmartcontractKey.NetworkId) {
		return message.Fail("network data not found for network id: " + request_parameters.Smartcontract.SmartcontractKey.NetworkId)
	}
	network, err := networks.Get(request_parameters.Smartcontract.SmartcontractKey.NetworkId)
	if err != nil {
		return message.Fail("networks.Get: " + err.Error())
	}

	network_sockets, ok := parameters[1].(key_value.KeyValue)
	if !ok {
		return message.Fail("no network sockets in the app parameters")
	}

	client_socket, ok := network_sockets[network.Type.String()].(*remote.ClientSocket)
	if !ok {
		return message.Fail("no network client for " + network.Type.String())
	}

	url := blockchain_inproc.CategorizerEndpoint(network.Id)
	categorizer_service, err := service.InprocessFromUrl(url)
	if err != nil {
		return message.Fail("blockchain_inproc.CategorizerEndpoint(network.Id): " + err.Error())
	}

	new_sm_request := blockchain_command.PushNewSmartcontracts{
		Smartcontracts: []smartcontract.Smartcontract{request_parameters.Smartcontract},
	}
	var new_sm_reply key_value.KeyValue
	err = blockchain_command.NEW_CATEGORIZED_SMARTCONTRACTS.RequestRouter(client_socket, categorizer_service, new_sm_request, &new_sm_reply)
	if err != nil {
		return message.Fail("failed to send to blockchain package: " + err.Error())
	}

	reply := SetSmartcontractsReply{}
	reply_message, err := command.Reply(reply)
	if err != nil {
		return message.Fail("parse reply: " + err.Error())
	}

	return reply_message
}
