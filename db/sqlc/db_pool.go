package sqlc

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"

	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/steve-mir/bukka_backend/utils"
)

func CreatePgxPool(config utils.Config) (*pgxpool.Pool, error) {
	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
		return nil, err
	}

	return connPool, err
}

func CreateDbPool(config utils.Config) (*sql.DB, error) {
	log.Info().Msg("CreateDbPool Initiated: ")
	db, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot connect to db")
	}
	// defer db.Close()

	DB_MAX_IDLE_CONN := config.DBMaxIdleConn
	DB_MAX_OPEN_CONN := config.DBMaxOpenConn
	DB_MAX_IDLE_TIME := config.DBMaxIdleTime
	DB_MAX_LIFE_TIME := config.DBMaxLifeTime

	db.SetMaxIdleConns(DB_MAX_IDLE_CONN)
	db.SetMaxOpenConns(DB_MAX_OPEN_CONN)
	db.SetConnMaxIdleTime(time.Duration(DB_MAX_IDLE_TIME) * time.Second)
	db.SetConnMaxLifetime(time.Duration(DB_MAX_LIFE_TIME) * time.Second)

	log.Info().Msgf("@CreateDbPool POSTGRES_SQL MAX Open Connections: %v", db.Stats().MaxOpenConnections)

	// This is for analyzing the stats after setting a connection

	log.Info().Msgf("@CreateDbPool MYSQL Open Connections: %v", db.Stats().OpenConnections)
	log.Info().Msgf("@CreateDbPool MYSQL InUse Connections: %v", db.Stats().InUse)
	log.Info().Msgf("@CreateDbPool MYSQL Idle Connections: %v", db.Stats().Idle)

	if err != nil {
		log.Fatal().Err(err).Msg("CreateDbPool Error: ")
		return nil, err
	}

	return db, err
}
