package api

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/steve-mir/bukka_backend/constants"
	db "github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/internal/app/auth/middlewares"
	"github.com/steve-mir/bukka_backend/internal/app/auth/services"
	"github.com/steve-mir/bukka_backend/internal/cache"
	"github.com/steve-mir/bukka_backend/token"
	"github.com/steve-mir/bukka_backend/utils"
	"github.com/steve-mir/bukka_backend/worker"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/time/rate"
)

const (
	baseURL = "v1/auth"
)

type RouteConfig struct {
	Method      string
	Path        string
	Handler     gin.HandlerFunc
	Middlewares []gin.HandlerFunc
	Roles       []int8
	RateLimit   middlewares.RateLimitConfig
}

type Server struct {
	store           db.Store
	Router          *gin.Engine
	db              *sql.DB
	config          utils.Config
	taskDistributor worker.TaskDistributor
	tokenService    *services.TokenService
	oauthConfig     *oauth2.Config
	tokenMaker      token.Maker
	cache           *cache.Cache
}

func NewServer(store db.Store, db *sql.DB, config utils.Config, td worker.TaskDistributor) *Server {
	tokenMaker, err := token.NewPasetoMaker(config.AccessTokenSymmetricKey, config.RefreshTokenSymmetricKey)
	if err != nil {
		panic(err)
	}

	cache := cache.NewCache(config.RedisAddress, config.RedisUsername, config.RedisPwd, 0)
	tokenService := services.NewTokenService(config, cache, tokenMaker)

	oauthConfig := &oauth2.Config{
		ClientID:     config.GoogleOauthClientId,
		ClientSecret: config.GoogleOauthClientSecret,
		RedirectURL:  config.GoogleOauthClientRedirect,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	server := &Server{
		store:           store,
		db:              db,
		config:          config,
		taskDistributor: td,
		tokenService:    tokenService,
		oauthConfig:     oauthConfig,
		tokenMaker:      tokenMaker,
		cache:           cache,
	}

	server.setupValidator()
	server.setupRouter()

	return server
}

func (server *Server) setupValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("emailValidator", utils.ValidEmail)
		v.RegisterValidation("phoneValidator", utils.ValidPhone)
		v.RegisterValidation("passwordValidator", utils.ValidPassword)
		v.RegisterValidation("usernameValidator", utils.ValidUsername)
	}
}

func (server *Server) setupRouter() {
	router := gin.Default()
	rl := setupRateLimiter()
	router.Use(middlewares.RateLimit(rl))

	routes := []RouteConfig{
		{Method: "POST", Path: "register", Handler: server.register, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(10 * time.Second), Burst: 2}},
		{Method: "POST", Path: "login", Handler: server.login, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(5 * time.Second), Burst: 3}},
		{Method: "GET", Path: "google/login", Handler: server.googleLogin, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(5 * time.Second), Burst: 3}},
		{Method: "GET", Path: "google/callback", Handler: server.googleCallback},
		{Method: "POST", Path: "rotate_token", Handler: server.rotateToken, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1}},
		{Method: "POST", Path: "verify_email", Handler: server.verifyEmail, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1}},
		{Method: "GET", Path: "resend_verification", Handler: server.resendVerificationEmail, Roles: []int8{constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers}, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1}},
		{Method: "DELETE", Path: "delete_account", Handler: server.deleteAccount, Roles: []int8{constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers}, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1}},
		{Method: "POST", Path: "request_account_recovery", Handler: server.requestAccountRecovery, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1}},
		{Method: "GET", Path: "recover_account", Handler: server.completeAccountRecovery, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1}},
		{Method: "POST", Path: "change_password", Handler: server.changePwd, Roles: []int8{constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers}, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1}},
		{Method: "POST", Path: "forgot_password", Handler: server.forgotPwd, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1}},
		{Method: "POST", Path: "reset_password", Handler: server.resetPwd, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1}},
		{Method: "GET", Path: "profile", Handler: server.viewProfile, Roles: []int8{constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers}},
		{Method: "GET", Path: "logout", Handler: server.logout, Roles: []int8{constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers}, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(5 * time.Second), Burst: 3}},
		{Method: "GET", Path: "home", Handler: server.home},
	}

	accessibleRoles := make(map[string][]int8)

	for _, route := range routes {
		fullPath := fmt.Sprintf("%s /%s/%s", route.Method, baseURL, route.Path)
		handlers := []gin.HandlerFunc{route.Handler}

		if len(route.Roles) > 0 {
			accessibleRoles[fullPath] = route.Roles
			router.Use(middlewares.AuthMiddleWare(server.config, server.tokenMaker, *server.cache, accessibleRoles))
		}

		if route.RateLimit.Rate != 0 {
			rl.SetRateLimitConfig("/"+route.Path, route.RateLimit)
		}

		router.Handle(route.Method, "/"+baseURL+"/"+route.Path, handlers...)
	}

	// Apply the auth middleware to all routes
	// router.Use(middlewares.AuthMiddleWare(server.config, server.tokenMaker, *server.cache, accessibleRoles))

	server.Router = router
}

func setupRateLimiter() *middlewares.RateLimiter {
	return middlewares.NewRateLimiter()
}

func (server *Server) Start(address string) error {
	return server.Router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
