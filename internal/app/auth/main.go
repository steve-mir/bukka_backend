package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/internal/app/auth/api"
	"github.com/steve-mir/bukka_backend/utils"
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

	store := sqlc.NewStore(connPool)
	server := api.NewServer(store)

	err = server.Start(config.HTTPAuthServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}
