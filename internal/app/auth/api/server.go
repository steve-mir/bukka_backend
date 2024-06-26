package api

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/internal/app/auth/middlewares"
	"github.com/steve-mir/bukka_backend/internal/cache"
	"github.com/steve-mir/bukka_backend/utils"
	"github.com/steve-mir/bukka_backend/worker"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/time/rate"
)

const (
	baseUrl = "v1/auth/"
)

type Server struct {
	store           db.Store
	Router          *gin.Engine
	db              *sql.DB
	config          utils.Config
	taskDistributor worker.TaskDistributor
	cache           *cache.Cache
	oauthConfig     *oauth2.Config
}

func NewServer(store db.Store, db *sql.DB, config utils.Config, td worker.TaskDistributor) *Server {
	cache := cache.NewCache(config.RedisAddress, config.RedisUsername, config.RedisPwd, 0) // ! Remove
	oauthConfig := &oauth2.Config{
		ClientID:     config.GoogleOauthClientId,
		ClientSecret: config.GoogleOauthClientSecret,
		RedirectURL:  config.GoogleOauthClientRedirect,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	server := &Server{store: store, db: db, config: config, taskDistributor: td, cache: cache, oauthConfig: oauthConfig}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("emailValidator", utils.ValidEmail)
		v.RegisterValidation("phoneValidator", utils.ValidPhone)
		v.RegisterValidation("passwordValidator", utils.ValidPassword)
		v.RegisterValidation("usernameValidator", utils.ValidUsername)
	}

	server.setupRouter()
	return server
}

func (server *Server) setupRouter() {
	router := gin.Default()

	rl := setupRateLimiter()
	router.Use(middlewares.RateLimit(rl))

	router.POST(baseUrl+"register", server.register)
	router.POST(baseUrl+"login", server.login)
	// Add Google Sign-In routes
	router.GET(baseUrl+"google/login", server.googleLogin)
	router.GET(baseUrl+"google/callback", server.googleCallback)

	router.POST(baseUrl+"rotate_token", server.rotateToken)
	router.POST(baseUrl+"verify_email", server.verifyEmail)

	authRoutes := router.Group("/").Use(middlewares.AuthMiddlerWare(server.config))
	authRoutes.GET(baseUrl+"resend_verification", server.resendVerificationEmail)

	authRoutes.DELETE(baseUrl+"delete_account", server.deleteAccount)
	router.POST(baseUrl+"request_account_recovery", server.requestAccountRecovery)
	router.GET(baseUrl+"recover_account", server.completeAccountRecovery)
	authRoutes.POST(baseUrl+"change_password", server.changePwd)
	router.POST(baseUrl+"forgot_password", server.forgotPwd)
	router.POST(baseUrl+"reset_password", server.resetPwd)
	// authRoutes.GET(baseUrl+"profile", server.viewProfile)
	// authRoutes.PATCH(baseUrl+"profile", server.updateProfile)
	authRoutes.GET(baseUrl+"logout", server.logout)
	router.GET(baseUrl+"home", server.home)

	// router.POST(baseUrl+"initiate_change_email", server.register)
	// router.POST(baseUrl+"confirm_change_email", server.register)
	// router.POST(baseUrl+"initiate_change_phone", server.register)
	// router.POST(baseUrl+"confirm_change_phone", server.register)
	// router.POST(baseUrl+"change_username", server.register)
	// router.PATCH(baseUrl+"update_user", server.register)
	// router.POST(baseUrl+"register_sso", server.register)
	// router.POST(baseUrl+"login_sso", server.register)
	// router.POST(baseUrl+"register_mfa", server.register)
	// router.POST(baseUrl+"verify_mfa_works", server.register)
	// router.POST(baseUrl+"verify_mfa", server.register)
	// router.POST(baseUrl+"bypass_mfa", server.register)

	server.Router = router
}

func setupRateLimiter() *middlewares.RateLimiter {
	rl := middlewares.NewRateLimiter()

	rl.SetRateLimitConfig("/register", middlewares.RateLimitConfig{Rate: rate.Every(10 * time.Second), Burst: 2})
	rl.SetRateLimitConfig("/login", middlewares.RateLimitConfig{Rate: rate.Every(5 * time.Second), Burst: 3})
	rl.SetRateLimitConfig("/google/login", middlewares.RateLimitConfig{Rate: rate.Every(5 * time.Second), Burst: 3})
	rl.SetRateLimitConfig("/logout", middlewares.RateLimitConfig{Rate: rate.Every(5 * time.Second), Burst: 3})
	rl.SetRateLimitConfig("/rotate_token", middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
	rl.SetRateLimitConfig("/verify_email", middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
	rl.SetRateLimitConfig("/resend_verification", middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
	rl.SetRateLimitConfig("/delete_account", middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
	rl.SetRateLimitConfig("/request_account_recovery", middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
	rl.SetRateLimitConfig("/recover_account", middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
	rl.SetRateLimitConfig("/change_password", middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
	rl.SetRateLimitConfig("/forgot_password", middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
	rl.SetRateLimitConfig("/reset_password", middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})

	return rl
}

// Start runs the HTTP server on a specifix address
func (server *Server) Start(address string) error {
	return server.Router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
