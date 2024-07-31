package gapi

import (
	"database/sql"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/steve-mir/bukka_backend/authentication/pb"
)

type Server struct {
	pb.UnimplementedUserAuthServer
	DB     *sql.DB
	Rabbit *amqp.Connection
}

func NewServer(db *sql.DB, rabbitConn *amqp.Connection) (*Server, error) {
	return &Server{
		DB:     db,
		Rabbit: rabbitConn,
	}, nil
}
