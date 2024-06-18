package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/utils"
)

func ChangeUserPwd(ctx context.Context, oldPassword, newPassword string, store sqlc.Store, uid uuid.UUID) error {
	// TODO: Check if the current session (claims.SessionID) is active

	user, err := store.GetUserByID(ctx, uid)
	if err != nil {
		return fmt.Errorf("couldn't verify old password %s", err)
	}

	err = utils.CheckPassword(oldPassword, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("old password incorrect %s", err)
	}

	// Hash password
	pwdHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("couldn't verify new password %s", err)
	}

	// Check if new password is the same as the old password
	if oldPassword == newPassword {
		return errors.New("new password cannot be the same as the old password")
	}

	// Update password
	err = store.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: pwdHash,
	})
	if err != nil {
		return fmt.Errorf("couldn't update password password %s", err)
	}

	// TODO: Close all the users session after changing password

	return nil
}
