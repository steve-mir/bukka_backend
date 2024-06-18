package main

import (
	"database/sql"

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
	go runTaskProcessor(redisOpt, store, db)
	runGinServer(store, db, config, taskDistributor, err)

}

func runGinServer(store sqlc.Store, db *sql.DB, config utils.Config, taskDistributor worker.TaskDistributor, err error) {
	server := api.NewServer(store, db, config, taskDistributor)

	err = server.Start(config.HTTPAuthServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}

func runTaskProcessor(redisOpt asynq.RedisClientOpt, store sqlc.Store, db *sql.DB) {
	mailer := worker.NewRedisTaskProcessor(redisOpt, store, db)
	// log.Info().Msg("Starting task processor")
	err := mailer.Start()
	if err != nil {
		log.Fatal().Msg("cannot start task processor")
	}

}
