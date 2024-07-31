package factory

import (
	"encoding/json"
	"fmt"
)

type MenuPayload struct {
	Action string `json:"menu_action"`
	ID     string `json:"id,omitempty"`
}

type MenuCommand struct {
	Payload MenuPayload
}

func (c MenuCommand) Execute() ([]byte, error) {
	return sendRequest("http://menu-service/menu", c.Payload)
}

type MenuCommandFactory struct{}

func (f MenuCommandFactory) CreateCommand(body []byte) (Command, error) {
	var payload MenuPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("error unmarshaling menu payload: %w", err)
	}
	return MenuCommand{Payload: payload}, nil
}
