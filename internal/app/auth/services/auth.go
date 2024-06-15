package services

import (
	"context"
	"errors"

	"github.com/steve-mir/bukka_backend/db/sqlc"
)

func CheckIfUsernameExists(ctx context.Context, qtx *sqlc.Queries, username string) error {
	// Compares the username lowercase with lowercase of that in db
	c, err := qtx.CheckUsername(ctx, username)
	if err != nil {
		return err
	}

	// If count is greater than 0 it means a username variant was found irrespective of the case
	if c > 0 {
		return errors.New("username taken")
	}

	// If none of the conditions are met, it means no user exists with that username and there were no errors.
	return nil
}
