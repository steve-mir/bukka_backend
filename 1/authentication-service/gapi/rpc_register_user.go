package gapi

import (
	"context"

	"github.com/steve-mir/bukka_backend/authentication/pb"
)

func (s *Server) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	return &pb.RegisterUserResponse{User: &pb.User{Username: req.User.Email, Phone: &req.User.Email}}, nil
}
