// Keep the credentials in a vault
package vault

import (
	"errors"
	"fmt"

	"github.com/blocklords/sds/app/command"
	"github.com/blocklords/sds/app/remote/message"
	"github.com/blocklords/sds/security/credentials"

	zmq "github.com/pebbe/zmq4"
)

const GET_STRING command.CommandName = "get-string"

func VaultEndpoint() string {
	return "inproc://sds_vault"
}

// Fetch the given credentials from the Vault.
// It fetches the private key from the vault.
// Then gets the public key from it
func GetCredentials(bucket string, key string) (*credentials.Credentials, error) {
	private_key, err := GetStringFromVault(bucket, key)
	if err != nil {
		return nil, fmt.Errorf("vault: %w", err)
	}

	pub_key, err := zmq.AuthCurvePublic(private_key)
	if err != nil {
		return nil, fmt.Errorf("zmq.Convert Secret to Pub: %w", err)
	}

	return credentials.NewPrivateKey(private_key, pub_key), nil
}

// Get the string value from the vault
func GetStringFromVault(bucket string, key string) (string, error) {
	// Socket to talk to clients
	socket, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		return "", err
	}

	if err := socket.Connect(VaultEndpoint()); err != nil {
		return "", fmt.Errorf("error to bind socket for: " + err.Error())
	}

	request := message.Request{
		Command: GET_STRING.String(),
		Parameters: map[string]interface{}{
			"bucket": bucket,
			"key":    key,
		},
	}

	request_string, _ := request.ToString()

	//  We send a request, then we work to get a reply
	socket.SendMessage(request_string)

	// Wait for reply.
	r, _ := socket.RecvMessage(0)

	reply, _ := message.ParseReply(r)
	if !reply.IsOK() {
		return "", errors.New(reply.Message)
	}

	value, _ := reply.Parameters.GetString("value")

	return value, nil
}
