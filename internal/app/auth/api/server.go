package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/utils"
	"github.com/steve-mir/bukka_backend/worker"
)

const (
	baseUrl = "v1/auth/"
)

type Server struct {
	store           db.Store
	router          *gin.Engine
	connPool        *pgxpool.Pool
	config          utils.Config
	taskDistributor worker.TaskDistributor
}

func NewServer(store db.Store, connPool *pgxpool.Pool, config utils.Config, td worker.TaskDistributor) *Server {
	server := &Server{store: store, connPool: connPool, config: config, taskDistributor: td}
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("emailValidator", utils.ValidEmail)
		v.RegisterValidation("phoneValidator", utils.ValidPhone)
		v.RegisterValidation("passwordValidator", utils.ValidPassword)
		v.RegisterValidation("usernameValidator", utils.ValidUsername)
	}

	router.POST(baseUrl+"register", server.register)
	router.POST(baseUrl+"login", server.login)
	router.POST(baseUrl+"rotate_token", server.rotateToken)
	router.POST(baseUrl+"verify_email", server.verifyEmail)
	// router.POST(baseUrl+"resend_verification", server.register)
	// router.POST(baseUrl+"change_password", server.register)
	// router.POST(baseUrl+"request_reset_password", server.register)
	// router.POST(baseUrl+"reset_password", server.register)
	// router.POST(baseUrl+"initiate_change_email", server.register)
	// router.POST(baseUrl+"confirm_change_email", server.register)
	// router.POST(baseUrl+"initiate_change_phone", server.register)
	// router.POST(baseUrl+"confirm_change_phone", server.register)
	// router.POST(baseUrl+"change_username", server.register)
	// router.PATCH(baseUrl+"update_user", server.register)
	// router.POST(baseUrl+"delete_account", server.register)
	// router.POST(baseUrl+"request_account_recovery", server.register)
	// router.POST(baseUrl+"recover_account", server.register)
	// router.POST(baseUrl+"register_sso", server.register)
	// router.POST(baseUrl+"login_sso", server.register)
	// router.POST(baseUrl+"register_mfa", server.register)
	// router.POST(baseUrl+"verify_mfa_works", server.register)
	// router.POST(baseUrl+"verify_mfa", server.register)
	// router.POST(baseUrl+"bypass_mfa", server.register)

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
