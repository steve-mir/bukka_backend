package main

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
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

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	redisOpt := asynq.RedisClientOpt{
		Addr:     config.RedisAddress,
		Username: config.RedisUsername,
		Password: config.RedisPwd,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	store := sqlc.NewStore(connPool)
	go runTaskProcessor(redisOpt, store, connPool)
	runGinServer(store, connPool, config, taskDistributor, err)

}

func runGinServer(store sqlc.Store, connPool *pgxpool.Pool, config utils.Config, taskDistributor worker.TaskDistributor, err error) {
	server := api.NewServer(store, connPool, config, taskDistributor)

	err = server.Start(config.HTTPAuthServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}

func runTaskProcessor(redisOpt asynq.RedisClientOpt, store sqlc.Store, connPool *pgxpool.Pool) {
	mailer := worker.NewRedisTaskProcessor(redisOpt, store, connPool)
	// log.Info().Msg("Starting task processor")
	err := mailer.Start()
	if err != nil {
		log.Fatal().Msg("cannot start task processor")
	}

}
