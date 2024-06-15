package services

import (
	"errors"

	"github.com/steve-mir/bukka_backend/db/sqlc"
)

func checkAccountStatus(usr sqlc.Authentication) error {
	if usr.IsSuspended.Bool {
		return errors.New("account suspended")
	}

	if usr.IsDeleted.Bool {
		return errors.New("account deleted")
	}
	return nil
}
