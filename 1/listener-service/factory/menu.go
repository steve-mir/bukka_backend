package factory

import (
	"encoding/json"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MenuPayload struct {
	Action string `json:"menu_action"`
	ID     string `json:"id,omitempty"`
}

type MenuCommand struct {
	Payload MenuPayload
	conn    *grpc.ClientConn
}

func (c MenuCommand) Execute() ([]byte, error) {
	// Implement gRPC execution here
	return nil, nil
}

type MenuCommandFactory struct {
	conn *grpc.ClientConn
}

func NewMenuCommandFactory() *MenuCommandFactory {
	return &MenuCommandFactory{}
}

func (f *MenuCommandFactory) CreateCommand(body []byte) (Command, error) {
	var payload MenuPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("error unmarshaling menu payload: %w", err)
	}

	conn, err := grpc.Dial("menu-service:5001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	return MenuCommand{Payload: payload, conn: conn}, nil
}

func (f *MenuCommandFactory) Close() error {
	if f.conn != nil {
		return f.conn.Close()
	}
	return nil
}
