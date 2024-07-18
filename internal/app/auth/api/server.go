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

// const (
// 	baseUrl                = "v1/auth/"
// 	home                   = "home"
// 	register               = "register"
// 	login                  = "login"
// 	gLogin                 = "google/login"
// 	gCallback              = "google/callback"
// 	rotateToken            = "rotate_token"
// 	verifyEmail            = "verify_email"
// 	resendVerification     = "resend_verification"
// 	deleteAccount          = "delete_account"
// 	requestAccountRecovery = "request_account_recovery"
// 	recoverAccount         = "recover_account"
// 	changePassword         = "change_password"
// 	forgotPassword         = "forgot_password"
// 	resetPassword          = "reset_password"
// 	profile                = "profile"
// 	logout                 = "logout"
// )

// type Server struct {
// 	store           db.Store
// 	Router          *gin.Engine
// 	db              *sql.DB
// 	config          utils.Config
// 	taskDistributor worker.TaskDistributor
// 	tokenService    *services.TokenService
// 	oauthConfig     *oauth2.Config
// 	tokenMaker      token.Maker
// 	cache           *cache.Cache
// }

// func NewServer(store db.Store, db *sql.DB, config utils.Config, td worker.TaskDistributor) *Server {
// 	// TODO: Handle error making token.
// 	tokenMaker, _ := token.NewPasetoMaker(config.AccessTokenSymmetricKey, config.RefreshTokenSymmetricKey)
// 	cache := cache.NewCache(config.RedisAddress, config.RedisUsername, config.RedisPwd, 0) // ! Remove
// 	tokenService := services.NewTokenService(config, cache, tokenMaker)

// 	oauthConfig := &oauth2.Config{
// 		ClientID:     config.GoogleOauthClientId,
// 		ClientSecret: config.GoogleOauthClientSecret,
// 		RedirectURL:  config.GoogleOauthClientRedirect,
// 		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
// 		Endpoint:     google.Endpoint,
// 	}

// 	server := &Server{
// 		store:           store,
// 		db:              db,
// 		config:          config,
// 		taskDistributor: td,
// 		tokenService:    tokenService,
// 		oauthConfig:     oauthConfig,
// 		tokenMaker:      tokenMaker,
// 		cache:           cache,
// 	}

// 	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
// 		v.RegisterValidation("emailValidator", utils.ValidEmail)
// 		v.RegisterValidation("phoneValidator", utils.ValidPhone)
// 		v.RegisterValidation("passwordValidator", utils.ValidPassword)
// 		v.RegisterValidation("usernameValidator", utils.ValidUsername)
// 	}

// 	server.setupRouter()
// 	return server
// }

// func accessibleRoles() map[string][]int8 {
// 	return map[string][]int8{

// 		// For All except guests
// 		"/" + baseUrl + resendVerification: {constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers},
// 		"/" + baseUrl + deleteAccount:      {constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers},
// 		"/" + baseUrl + changePassword:     {constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers},
// 		"/" + baseUrl + profile:            {constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers},
// 		"/" + baseUrl + logout:             {constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers},
// 	}
// }
// func (server *Server) setupRouter() {
// 	router := gin.Default()

// 	rl := setupRateLimiter()
// 	router.Use(middlewares.RateLimit(rl))

// 	router.POST(baseUrl+register, server.register)
// 	router.POST(baseUrl+login, server.login)
// 	// Add Google Sign-In routes
// 	router.GET(baseUrl+gLogin, server.googleLogin)
// 	router.GET(baseUrl+gCallback, server.googleCallback)

// 	router.POST(baseUrl+rotateToken, server.rotateToken)
// 	router.POST(baseUrl+verifyEmail, server.verifyEmail)

// 	authRoutes := router.Group("/").Use(middlewares.AuthMiddleWare(server.config, server.tokenMaker, *server.cache, accessibleRoles()))
// 	authRoutes.GET(baseUrl+resendVerification, server.resendVerificationEmail)

// 	authRoutes.DELETE(baseUrl+deleteAccount, server.deleteAccount)
// 	router.POST(baseUrl+requestAccountRecovery, server.requestAccountRecovery)
// 	router.GET(baseUrl+recoverAccount, server.completeAccountRecovery)
// 	authRoutes.POST(baseUrl+changePassword, server.changePwd)
// 	router.POST(baseUrl+forgotPassword, server.forgotPwd)
// 	router.POST(baseUrl+resetPassword, server.resetPwd)
// 	authRoutes.GET(baseUrl+profile, server.viewProfile)
// 	// authRoutes.PATCH(baseUrl+"profile", server.updateProfile)
// 	authRoutes.GET(baseUrl+logout, server.logout)
// 	router.GET(baseUrl+home, server.home)

// 	server.Router = router
// }

// func setupRateLimiter() *middlewares.RateLimiter {
// 	rl := middlewares.NewRateLimiter()

// 	rl.SetRateLimitConfig("/"+register, middlewares.RateLimitConfig{Rate: rate.Every(10 * time.Second), Burst: 2})
// 	rl.SetRateLimitConfig("/"+login, middlewares.RateLimitConfig{Rate: rate.Every(5 * time.Second), Burst: 3})
// 	rl.SetRateLimitConfig("/"+gLogin, middlewares.RateLimitConfig{Rate: rate.Every(5 * time.Second), Burst: 3})
// 	rl.SetRateLimitConfig("/"+logout, middlewares.RateLimitConfig{Rate: rate.Every(5 * time.Second), Burst: 3})
// 	rl.SetRateLimitConfig("/"+rotateToken, middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
// 	rl.SetRateLimitConfig("/"+verifyEmail, middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
// 	rl.SetRateLimitConfig("/"+resendVerification, middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
// 	rl.SetRateLimitConfig("/"+deleteAccount, middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
// 	rl.SetRateLimitConfig("/"+requestAccountRecovery, middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
// 	rl.SetRateLimitConfig("/"+recoverAccount, middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
// 	rl.SetRateLimitConfig("/"+changePassword, middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
// 	rl.SetRateLimitConfig("/"+forgotPassword, middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})
// 	rl.SetRateLimitConfig("/"+resetPassword, middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1})

// 	return rl
// }

// // Start runs the HTTP server on a specifix address
// func (server *Server) Start(address string) error {
// 	return server.Router.Run(address)
// }

// func errorResponse(err error) gin.H {
// 	return gin.H{"error": err.Error()}
// }
