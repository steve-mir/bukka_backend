package sqlc

import (
	"os"
	"testing"

	// "github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/steve-mir/bukka_backend/utils"
)

var testStore Store

func TestMain(m *testing.M) {
	config, err := utils.LoadConfig("../../")
	if err != nil {
		log.Fatal().Msg("cannot load config " + err.Error())
	}

	// connPool, err := pgxpool.New(context.Background(), config.DBSource)
	db, err := CreateDbPool(config)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	testStore = NewStore(db)

	os.Exit(m.Run())
}
