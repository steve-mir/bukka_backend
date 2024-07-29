package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *Config) routes() http.Handler {
	router := gin.Default()

	// specify who is allowed to connect
	// router.Use(cors.Handler(cors.Options{
	// 	AllowedOrigins: []string{"https://*", "http://*"},
	// 	AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	// 	ExposedHeaders: []string{"Link"},
	// 	AllowCredentials: true,
	// 	MaxAge: 300,
	// }))

	// mux.Use(middleware.Heartbeat("/ping"))

	router.Handle("POST", "/", app.Broker())

	// router.Handle("POST", "/log-grpc", app.LogViaGRPC)

	router.Handle("POST", "/handle", app.HandleSubmission())

	return router
}
