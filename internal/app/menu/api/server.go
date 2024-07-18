package api

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	db "github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/internal/app/auth/middlewares"
	"github.com/steve-mir/bukka_backend/internal/cache"
	"github.com/steve-mir/bukka_backend/token"
	"github.com/steve-mir/bukka_backend/utils"
	"github.com/steve-mir/bukka_backend/worker"
)

const (
	baseURL = "v1/menu"
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
	tokenMaker      token.Maker
	cache           *cache.Cache
}

func NewServer(store db.Store, db *sql.DB, config utils.Config, td worker.TaskDistributor) *Server {
	tokenMaker, err := token.NewPasetoMaker(config.AccessTokenSymmetricKey, config.RefreshTokenSymmetricKey)
	if err != nil {
		panic(err)
	}

	cache := cache.NewCache(config.RedisAddress, config.RedisUsername, config.RedisPwd, 0)

	server := &Server{
		store:           store,
		db:              db,
		config:          config,
		taskDistributor: td,
		tokenMaker:      tokenMaker,
		cache:           cache,
	}

	// server.setupValidator()
	server.setupRouter()

	return server
}

// func (server *Server) setupValidator() {
// 	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
// 		v.RegisterValidation("emailValidator", utils.ValidEmail)
// 		v.RegisterValidation("phoneValidator", utils.ValidPhone)
// 		v.RegisterValidation("passwordValidator", utils.ValidPassword)
// 		v.RegisterValidation("usernameValidator", utils.ValidUsername)
// 	}
// }

func (server *Server) setupRouter() {
	router := gin.Default()
	rl := setupRateLimiter()
	router.Use(middlewares.RateLimit(rl))

	routes := []RouteConfig{
		{Method: "GET", Path: "home", Handler: server.home},
		// {Method: "POST", Path: "categories", Handler: server.register, Roles: []int8{constants.SuperAdmin, constants.AppAdmin}, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(10 * time.Second), Burst: 2}},
		// {Method: "GET", Path: "categories", Handler: server.login, Roles: []int8{constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers}, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(5 * time.Second), Burst: 3}},
		// {Method: "GET", Path: "dishes", Handler: server.googleLogin, Roles: []int8{constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers}, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(5 * time.Second), Burst: 3}},
		// {Method: "POST", Path: "dishes", Handler: server.rotateToken, Roles: []int8{constants.SuperAdmin, constants.AppAdmin}, RateLimit: middlewares.RateLimitConfig{Rate: rate.Every(1 * time.Minute), Burst: 1}},
		// {Method: "GET", Path: "usersFavorites", Handler: server.googleCallback, Roles: []int8{constants.RegularUsers}},

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

// const (
// 	baseUrl        = "v1/menu/"
// 	home           = "home"
// 	categories     = "categories"
// 	dishes         = "dishes"
// 	usersFavorites = "users/favorites"
// )

// type Server struct {
// 	store           db.Store
// 	Router          *gin.Engine
// 	db              *sql.DB
// 	config          utils.Config
// 	taskDistributor worker.TaskDistributor
// }

// func NewServer(store db.Store, db *sql.DB, config utils.Config, td worker.TaskDistributor) *Server {

// 	server := &Server{
// 		store:           store,
// 		db:              db,
// 		config:          config,
// 		taskDistributor: td,
// 	}

// 	server.setupRouter()
// 	return server
// }

// func accessibleRoles() map[string][]int8 {
// 	return map[string][]int8{

// 		// For All except guests
// 		"/" + baseUrl + home:           {constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers},
// 		"/" + baseUrl + categories:     {constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers},
// 		"/" + baseUrl + dishes:         {constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers},
// 		"/" + baseUrl + usersFavorites: {constants.SuperAdmin, constants.AppAdmin, constants.RegularUsers},
// 		// "/" + baseUrl + categories:     {constants.SuperAdmin, constants.AppAdmin, constants.Restaurant}, add new category
// 	}
// }

// func (server *Server) setupRouter() {
// 	router := gin.Default()

// 	rl := setupRateLimiter()
// 	router.Use(middlewares.RateLimit(rl))

// 	// authRoutes := router.Group("/").Use(middlewares.AuthMiddlerWare(server.config, server.tokenMaker, *server.cache))

// 	// authRoutes.POST(baseUrl+"change_password", server.changePwd)

// 	server.Router = router
// }

// func setupRateLimiter() *middlewares.RateLimiter {
// 	rl := middlewares.NewRateLimiter()

// 	rl.SetRateLimitConfig("/register", middlewares.RateLimitConfig{Rate: rate.Every(10 * time.Second), Burst: 3})

// 	return rl
// }

// func errorResponse(err error) gin.H {
// 	return gin.H{"error": err.Error()}
// }
