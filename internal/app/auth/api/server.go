package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/steve-mir/bukka_backend/db/sqlc"
)

// github.com/stretchr/testify/require
type Server struct {
	store  db.Store
	router *gin.Engine
}

func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.POST("v1/auth/register", server.register)
	// router.POST("v1/auth/:uid", server.getUser)

	server.router = router
	return server
}

// Start runs the HTTP server on a specifix address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
