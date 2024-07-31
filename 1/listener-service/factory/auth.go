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

type AuthPayload struct {
	Action   string `json:"auth_action"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	Username string `json:"username,omitempty"`
}

type AuthCommand struct {
	Payload AuthPayload
	client  pb.UserAuthClient
}

func (c AuthCommand) Execute() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var response interface{}
	var err error

	switch c.Payload.Action {
	case "register":
		response, err = c.client.RegisterUser(ctx, &pb.RegisterUserRequest{
			User:     &pb.User{Email: c.Payload.Email, Username: c.Payload.Username},
			Password: c.Payload.Password,
		})
	// ! Implement other cases (login, forgot) as needed
	default:
		return nil, fmt.Errorf("invalid auth action: %s", c.Payload.Action)
	}

	if err != nil {
		return nil, fmt.Errorf("error executing gRPC call: %w", err)
	}

	return json.Marshal(response)
}

type AuthCommandFactory struct {
	client    pb.UserAuthClient
	conn      *grpc.ClientConn
	once      sync.Once
	clientErr error
}

func NewAuthCommandFactory() *AuthCommandFactory {
	return &AuthCommandFactory{}
}

func (f *AuthCommandFactory) initClient() {
	f.once.Do(func() {
		conn, err := grpc.Dial("authentication-service:5001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			f.clientErr = fmt.Errorf("failed to connect to gRPC server: %w", err)
			return
		}
		f.conn = conn
		f.client = pb.NewUserAuthClient(conn)
	})
}

func (f *AuthCommandFactory) CreateCommand(body []byte) (Command, error) {
	f.initClient()
	if f.clientErr != nil {
		return nil, f.clientErr
	}

	var payload AuthPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("error unmarshaling auth payload: %w", err)
	}

	return AuthCommand{Payload: payload, client: f.client}, nil
}

func (f *AuthCommandFactory) Close() error {
	if f.conn != nil {
		return f.conn.Close()
	}
	return nil
}
