package factory

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/steve-mir/bukka_backend/listener/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MenuPayload struct {
	Action string `json:"menu_action"`
	ID     string `json:"id,omitempty"`
}

type MenuCommand struct {
	Payload MenuPayload
	client  pb.MenuClient
}

func (c MenuCommand) Execute() ([]byte, error) {
	// Implement gRPC execution here
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var response interface{}
	var err error

	switch c.Payload.Action {
	case "home":
		response, err = c.client.GetMenu(ctx, &pb.GetMenuRequest{})
	// ! Implement other cases (login, forgot) as needed
	default:
		return nil, fmt.Errorf("invalid menu action: %s", c.Payload.Action)
	}

	if err != nil {
		return nil, fmt.Errorf("error executing gRPC call: %w", err)
	}

	return json.Marshal(response)
}

type MenuCommandFactory struct {
	conn      *grpc.ClientConn
	client    pb.MenuClient
	once      sync.Once
	clientErr error
}

func NewMenuCommandFactory() *MenuCommandFactory {
	return &MenuCommandFactory{}
}

func (f *MenuCommandFactory) initClient() {
	f.once.Do(func() {
		conn, err := grpc.Dial("menu-service:5001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			f.clientErr = fmt.Errorf("failed to connect to gRPC server: %w", err)
			return
		}
		f.conn = conn
		f.client = pb.NewMenuClient(conn)
	})
}

func (f *MenuCommandFactory) CreateCommand(body []byte) (Command, error) {
	f.initClient()
	if f.clientErr != nil {
		return nil, f.clientErr
	}

	var payload MenuPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("error unmarshaling menu payload: %w", err)
	}

	return MenuCommand{Payload: payload, client: f.client}, nil
}

func (f *MenuCommandFactory) Close() error {
	if f.conn != nil {
		return f.conn.Close()
	}
	return nil
}
