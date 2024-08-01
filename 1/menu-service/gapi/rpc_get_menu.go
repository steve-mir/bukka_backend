package gapi

import (
	"context"

	"github.com/steve-mir/bukka_backend/menu/pb"
)

func (s *Server) GetMenu(ctx context.Context, req *pb.GetMenuRequest) (*pb.GetMenuResponse, error) {
	return &pb.GetMenuResponse{MenuId: "menu_id"}, nil
}
