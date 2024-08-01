package gapi

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/steve-mir/bukka_backend/menu/pb"
)

type Server struct {
	pb.UnimplementedMenuServer
	// DB     *sql.DB
	Rabbit *amqp.Connection
}

func NewServer(rabbitConn *amqp.Connection) (*Server, error) {
	return &Server{
		// DB:     db,
		Rabbit: rabbitConn,
	}, nil
}
