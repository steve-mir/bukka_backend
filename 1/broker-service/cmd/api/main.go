package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "80"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	// try to connect to rabbitmq
	rabbitConn, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitConn.Close()

	app := Config{
		Rabbit: rabbitConn,
	}

	log.Printf("Starting broker service on port %s\n", webPort)

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start the server
	// server := setupGinServer(store, db, config, taskDistributor)
	startGinServer(srv, webPort)
	// err = srv.ListenAndServe()
	// if err != nil {
	// 	log.Panic(err)
	// }
}

// func setupGinServer(store sqlc.Store, db *sql.DB, config utils.Config, taskDistributor worker.TaskDistributor) *http.Server {
// 	apiServer := api.NewServer(store, db, config, taskDistributor)
// 	return &http.Server{
// 		Addr:         config.HTTPAuthServerAddress,
// 		Handler:      apiServer.Router,
// 		IdleTimeout:  120 * time.Second,
// 		ReadTimeout:  1 * time.Second,
// 		WriteTimeout: 1 * time.Second,
// 	}
// }

func startGinServer(server *http.Server, address string) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("could not listen on %s: %v\n", address, err)
		}
	}()

	<-ctx.Done()

	log.Printf("shutting down gracefully, press Ctrl+C again to force")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutDown); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	log.Printf("server exiting")
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	// don't continue until rabbit is ready
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			log.Println("Connected to RabbitMQ!")
			connection = c
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off...")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}
