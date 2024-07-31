package factory

import (
	"encoding/json"
	"fmt"
)

type AuthPayload struct {
	Action   string `json:"auth_action"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type AuthCommand struct {
	Payload AuthPayload
}

func (c AuthCommand) Execute() ([]byte, error) {
	return sendRequest("http://authentication-service/authenticate", c.Payload)
}

type AuthCommandFactory struct{}

func (f AuthCommandFactory) CreateCommand(body []byte) (Command, error) {
	var payload AuthPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("error unmarshaling auth payload: %w", err)
	}
	return AuthCommand{Payload: payload}, nil
}
