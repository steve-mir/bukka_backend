package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/internal/app/auth/api"
	"github.com/steve-mir/bukka_backend/utils"
	"github.com/steve-mir/bukka_backend/worker"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal().Msg("cannot load config " + err.Error())
	}

	// Run db migrations
	// runDbMigration(config.MigrationUrl, config.DBSource)

	// connPool, err := pgxpool.New(context.Background(), config.DBSource)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("cannot connect to db")
	// }
	db, err := sqlc.CreateDbPool(config)
	if err != nil {
		log.Fatal().Msg("cannot create db pool")
		return
	}

	redisOpt := asynq.RedisClientOpt{
		Addr:     config.RedisAddress,
		Username: config.RedisUsername,
		Password: config.RedisPwd,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	store := sqlc.NewStore(db)
	go runTaskProcessor(redisOpt, store, db, config)
	// runGinServer(store, db, config, taskDistributor, err)
	server := setupGinServer(store, db, config, taskDistributor)
	startGinServer(server, config.HTTPAuthServerAddress)

}

func setupGinServer(store sqlc.Store, db *sql.DB, config utils.Config, taskDistributor worker.TaskDistributor) *http.Server {
	apiServer := api.NewServer(store, db, config, taskDistributor)
	return &http.Server{
		Addr:         config.HTTPAuthServerAddress,
		Handler:      apiServer.Router,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}
}

func startGinServer(server *http.Server, address string) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("could not listen on %s: %v\n", address, err)
		}
	}()

	<-ctx.Done()

	log.Info().Msg("shutting down gracefully, press Ctrl+C again to force")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutDown); err != nil {
		log.Info().Msgf("server forced to shutdown: %v", err)
	}

	log.Fatal().Msgf("server exiting")
}

func runDbMigration(migrationUrl string, dbSource string) {
	migration, err := migrate.New(migrationUrl, dbSource)
	if err != nil {
		log.Fatal().Msg("cannot create new migration instance:") //, err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msg("failed to run migrate up:") //, err)
	}
	log.Info().Msg("db migrated successfully")

}

// func runGinServer(store sqlc.Store, db *sql.DB, config utils.Config, taskDistributor worker.TaskDistributor, err error) {

// 	server := api.NewServer(store, db, config, taskDistributor)

// 	err = server.Start(config.HTTPAuthServerAddress)
// 	if err != nil {
// 		log.Fatal().Err(err).Msg("cannot start server")
// 	}
// }

func runTaskProcessor(redisOpt asynq.RedisClientOpt, store sqlc.Store, db *sql.DB, config utils.Config) {
	mailer := worker.NewRedisTaskProcessor(redisOpt, store, db, config)
	// log.Info().Msg("Starting task processor")
	err := mailer.Start()
	if err != nil {
		log.Fatal().Msg("cannot start task processor")
	}

}
