package main

import (
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/steve-mir/bukka_backend/menu/gapi"
	"github.com/steve-mir/bukka_backend/menu/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	webPort  = "80"
	gRPCPort = "0.0.0.0:5001"
)

var counts int64

type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	log.Info().Msg("Starting menu service")

	// connect to DB
	// conn := connectToDB()
	// if conn == nil {
	// 	log.Panic().Msg("Can't connect to Postgres!")
	// }

	// try to connect to rabbitmq
	rabbitConn, err := connect()
	if err != nil {
		log.Error().Msg(err.Error())
		os.Exit(1)
	}
	defer rabbitConn.Close()

	// set up config
	// app := Config{
	// 	Rabbit: rabbitConn,
	// }

	// srv := &http.Server{
	// 	Addr:    fmt.Sprintf(":%s", webPort),
	// 	Handler: app.routes(),
	// }

	// err = srv.ListenAndServe()
	// if err != nil {
	// 	log.Panic().Msg(err.Error())
	// }
	runGrpcServer(rabbitConn)
}

/*
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Info().Msg("Postgres not yet ready ...")
			counts++
		} else {
			log.Info().Msg("Connected to Postgres!")
			return connection
		}

		if counts > 10 {
			log.Info().Msgf("%v", err)
			return nil
		}

		log.Info().Msg("Backing off for two seconds....")
		time.Sleep(2 * time.Second)
		continue
	}
}
*/

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
			log.Info().Msg("Connected to RabbitMQ!")
			connection = c
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Info().Msg("backing off...")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}

func newGRPCServer() *grpc.Server {
	return grpc.NewServer()
}

func runGrpcServer(rabbitConn *amqp.Connection) {
	// Create a context that listens for termination signals
	ctx, cancel := context.WithCancel(context.Background())

	// Handle graceful shutdown on receiving termination signals
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		<-signals
		log.Info().Msg("Received termination signal. Shutting down gracefully...")
		cancel()
	}()

	// Auth server
	server, err := gapi.NewServer(rabbitConn)
	if err != nil {
		log.Fatal().Msg("cannot create a server:")
	}

	grpcServer := newGRPCServer()

	pb.RegisterMenuServer(grpcServer, server)

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", gRPCPort)
	if err != nil {
		log.Fatal().Msgf("cannot create listener: %v", err)
	}

	log.Info().Msgf("start grpc server at %s", listener.Addr().String())

	// Start the gRPC server in a goroutine
	go func() {
		err := grpcServer.Serve(listener)
		if err != nil {
			log.Fatal().Msgf("cannot start grpc server %v", err)
		}
	}()

	// Wait for the context to be canceled (either by the termination signal or an error)
	<-ctx.Done()

	// Stop the gRPC server
	grpcServer.GracefulStop()

	// Log a message indicating a graceful shutdown
	log.Info().Msg("gRPC server stopped gracefully")
}
