package sqlc

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/steve-mir/bukka_backend/utils"
)

var testQueries *Queries

func TestMain(m *testing.M) {
	config, err := utils.LoadConfig("../../")
	if err != nil {
		log.Fatal().Msg("cannot load config " + err.Error())
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	testQueries = New(conn)

	os.Exit(m.Run())
}
