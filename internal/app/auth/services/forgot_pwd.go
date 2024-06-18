package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/utils"
	"github.com/steve-mir/bukka_backend/worker"
	"golang.org/x/sync/errgroup"
)

func RequestPwdReset(ctx context.Context, email string, store sqlc.Store, td worker.TaskDistributor) error {

	usr, pwdResetCodeStr, err := initChangeRequest(ctx, store, email)
	if err != nil {
		return errors.New(ResetMsg)
	}
	msg := fmt.Sprintf("Below is code to reset your password: %s.\nPlease do not share this with anyone", pwdResetCodeStr)
	var eg errgroup.Group

	// Add link to db
	eg.Go(func() error {
		if err := store.CreatePasswordResetRequest(ctx, sqlc.CreatePasswordResetRequestParams{
			UserID:    usr.ID,
			Email:     usr.Email,
			Token:     pwdResetCodeStr,
			ExpiresAt: time.Now().Add(time.Minute * 15),
		}); err != nil {
			return err
		}
		return nil
	})

	// Send email
	eg.Go(func() error {
		if err := SendEmail(td, ctx, email, msg); err != nil {
			return err
		}
		return nil
	})

	// Wait for both goroutines to complete
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("an unexpected error occurred: %v", err)
	}

	return nil
}

func ResetPassword(ctx context.Context, qtx *sqlc.Queries, tx *sql.Tx, store sqlc.Store, code, pwd string) error {
	// Create a context with a timeout for the transaction
	ctx, cancel := context.WithTimeout(ctx, time.Second*10) // Adjust the timeout as needed
	defer cancel()

	if len(code) != length {
		return errors.New("invalid token")
	}

	tokenData, err := store.GetPasswordResetRequestByToken(ctx, code)
	if err != nil {
		return err
	}

	if tokenData.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("token expired")
	}

	if tokenData.Used.Bool {
		return fmt.Errorf("token already used")
	}

	// ! 1 Get User
	user, err := getUser(ctx, store, tokenData.Email)
	if err != nil {
		return err
	}

	err = checkAccountStatus(user)
	if err != nil {
		return err
	}

	// ! 2 Check old password
	if err = utils.CheckPassword(pwd, user.PasswordHash); err == nil {
		return errors.New("cannot use old password")
	}

	// ! 3 Hash password
	hashedPwd, err := utils.HashPassword(pwd)
	if err != nil {
		return err
	}

	var eg errgroup.Group

	// Goroutine for updating token status
	eg.Go(func() error {
		// Update password_request table
		err := updateResetPwdTokenStatus(ctx, qtx, code)
		if err != nil {
			return err
		}
		return nil
	})

	// Goroutine for updating user password
	eg.Go(func() error {

		// Update users table
		err := updateUserPassword(ctx, qtx, tokenData.Email, hashedPwd)
		if err != nil {
			return err
		}
		return nil
	})

	// Wait for both goroutines to complete
	if err := eg.Wait(); err != nil {
		tx.Rollback()
		return errors.New("error resetting password " + err.Error())
	}

	// Commit the transaction if all updates were successful
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil

}
func getUser(ctx context.Context, store sqlc.Store, email string) (sqlc.Authentication, error) {
	user, err := store.GetUserByIdentifier(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return sqlc.Authentication{}, errors.New("1. email not found")
		}
		return sqlc.Authentication{}, errors.New("email not found")
	}
	return user, nil
}

func updateResetPwdTokenStatus(ctx context.Context, qtx *sqlc.Queries, token string) error {
	err := qtx.UpdatePasswordResetRequestByToken(ctx, sqlc.UpdatePasswordResetRequestByTokenParams{
		Token: token,
		Used:  sql.NullBool{Bool: true, Valid: true},
	})
	return err
}

func updateUserPassword(ctx context.Context, qtx *sqlc.Queries, email, hashedPwd string) error {
	// Change password
	err := qtx.UpdateUserPasswordByEmail(ctx, sqlc.UpdateUserPasswordByEmailParams{
		Email:        email,
		PasswordHash: hashedPwd,
	})
	return err
}

func initChangeRequest(ctx context.Context, store sqlc.Store, id string) (sqlc.Authentication, string, error) {
	emailResetCode, err := utils.GenerateSecureRandomNumber(codeLength)
	if err != nil {
		return sqlc.Authentication{}, "", fmt.Errorf("failed to generate secure random number %v", err)
	}

	// Check if identifier exists
	usr, err := store.GetUserByIdentifier(ctx, id)
	if err != nil {
		return sqlc.Authentication{}, "", fmt.Errorf("failed to get user by identifier, %s. %v", ResetMsg, err)
	}

	// check account status
	err = checkAccountStatus(usr)
	if err != nil {
		return sqlc.Authentication{}, "", err
	}

	return usr, fmt.Sprintf("%06d", emailResetCode), nil
}
