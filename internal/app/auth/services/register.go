package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/steve-mir/bukka_backend/constants"
	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/token"
	"github.com/steve-mir/bukka_backend/utils"
	"github.com/steve-mir/bukka_backend/worker"
	"golang.org/x/sync/errgroup"
)

type RegisterReq struct {
	FullName string `json:"full_name" binding:"required"`
	Username string `json:"username" binding:"required,usernameValidator"`
	Email    string `json:"email" binding:"required,emailValidator"`
	Phone    string `json:"phone" binding:"phoneValidator"`
	Password string `json:"password" binding:"required,passwordValidator"`
	FcmToken string `json:"fcm_token"`
}

type UserResult struct {
	User sqlc.Authentication
	Err  error
}

type AccessTokenResult struct {
	AccessToken string
	Payload     *token.Payload
	Err         error
}

func CheckUserExists(ctx context.Context, qtx *sqlc.Queries, email, username string) error {
	// Check db if email exists
	if err := CheckEmailExistsError(ctx, qtx, email); err != nil {
		fmt.Println("Error checking email", email)
		return err
	}

	// Check db if username exists
	if err := CheckIfUsernameExists(ctx, qtx, username); err != nil {
		fmt.Println("Error checking username", username)
		return err
	}
	return nil
}

func PrepareUserData(pwd string) (string, uuid.UUID, error) {
	hashedPwd, err := utils.HashPassword(pwd)
	if err != nil {
		return "", uuid.UUID{}, errors.New("error processing data")
	}

	// Generate UUID in advance
	uid, err := uuid.NewRandom()
	if err != nil {
		return "", uuid.UUID{}, errors.New("an unexpected error occurred")
	}

	return hashedPwd, uid, nil
}

func CreateUserConcurrent(ctx context.Context, qtx *sqlc.Queries /*tx *sql.Tx,*/, uid uuid.UUID, email, username, pwd string, isEmailVerified, isOauthUser bool) (sqlc.Authentication, error) {
	params := sqlc.CreateUserParams{
		ID:           uid,
		Email:        email,
		Username:     sql.NullString{String: username, Valid: true},
		PasswordHash: pwd,
		IsEmailVerified: sql.NullBool{
			Bool:  isEmailVerified,
			Valid: true,
		},
		IsOauthUser: sql.NullBool{
			Bool:  isOauthUser,
			Valid: true,
		},
	}

	userData, err := qtx.CreateUser(ctx, params)
	if err != nil {
		return sqlc.Authentication{}, errors.New("error creating user: " + err.Error())
	}

	return sqlc.Authentication{
		Email:           userData.Email,
		Username:        userData.Username,
		IsEmailVerified: userData.IsEmailVerified,
		CreatedAt:       userData.CreatedAt,
		ID:              userData.ID,
	}, nil
}

func RunConcurrentUserCreationTasks(ctx context.Context, tokenMaker token.Maker, qtx *sqlc.Queries, tx *sql.Tx, config utils.Config, td worker.TaskDistributor,
	req RegisterReq, uid uuid.UUID, clientIP string, agent string, isEmailVerified bool) (string, time.Time, error) {

	var accessToken string
	var accessPayload *token.Payload
	var err error

	var eg errgroup.Group

	// Create access token
	eg.Go(func() error {
		accessToken, accessPayload, err = tokenMaker.CreateToken(
			token.PayloadData{
				Role:          constants.RegularUsers,
				Subject:       uid,
				Username:      req.Username,
				Email:         req.Email,
				EmailVerified: false,
				Issuer:        config.AppName,
				Audience:      "website users",
				IP:            clientIP,
				UserAgent:     agent,
				MfaPassed:     true,
				TokenType:     token.TokenType(token.AccessToken),
			}, config.AccessTokenDuration, token.TokenType(token.AccessToken),
		)
		if err != nil {
			return err
		}

		return nil
	})

	// Create user profile
	eg.Go(func() error {
		err = qtx.CreateUserProfile(ctx, sqlc.CreateUserProfileParams{
			UserID:    uid,
			FirstName: sql.NullString{String: req.FullName, Valid: true},
			LastName:  sql.NullString{String: req.FullName, Valid: true},
		})
		if err != nil {
			return err
		}

		return nil
	})

	// Create user role
	eg.Go(func() error {
		_, err = qtx.CreateUserRole(ctx, sqlc.CreateUserRoleParams{
			UserID: uid,
			RoleID: constants.RegularUsers,
		})
		if err != nil {
			return err
		}

		return nil
	})

	// Send verification email concurrently (does not need to be within transaction)
	if !isEmailVerified {

		eg.Go(func() error {
			err = SendVerificationEmail(qtx, ctx, td, uid, req.Email)
			if err != nil {
				return err
			}

			return nil
		})
	}

	// Wait for both goroutines to complete
	if err := eg.Wait(); err != nil {
		tx.Rollback()
		return "", time.Time{}, errors.New("error registering " + err.Error())
	}

	return accessToken, accessPayload.Expires, nil
}

// ?----------------
func CheckEmailExistsError(ctx context.Context, qtx *sqlc.Queries, email string) error {
	// Check duplicate emails
	user, err := qtx.GetUserByIdentifier(ctx, email)
	if err != nil && err != sql.ErrNoRows {
		// An error occurred that isn't simply indicating no rows were found
		return err
	}

	if user.ID != uuid.Nil {
		// User exists, check if the account is marked as deleted
		if user.DeletedAt.Valid {
			// Check if the account is within the recovery period
			if time.Since(user.DeletedAt.Time) <= MaxAccountRecoveryDuration {
				// Account is within the recovery period and can be recovered
				return errors.New("account is deleted but can be recovered, please follow the account recovery process")
			} else {
				// Account is beyond the recovery period, append timestamp to the email to make it unique
				err = appendTimestampToEmail(ctx, qtx, user.Email, user.DeletedAt.Time)
				if err != nil {
					return fmt.Errorf("failed to update email for user with ID %s: %v", user.ID, err)
				}
			}
		} else {
			// Account exists and is not marked as deleted
			return errors.New("email already exists")
		}
	}
	return nil
}

// Placeholder store method to append a timestamp to the user's email
func appendTimestampToEmail(ctx context.Context, qtx *sqlc.Queries, email string, deletedAt time.Time) error {
	// Implement the logic to append a timestamp to the user's email.
	// This will involve updating the user record in the database.
	// Be careful to ensure that the new email remains unique and valid.
	// For example, you might append something like "_deleted_1612385610" to the email.
	newEmail := addDeleteTimeToEmail(email, deletedAt)

	_, err := qtx.UpdateUser(ctx, sqlc.UpdateUserParams{
		Email: sql.NullString{String: newEmail, Valid: true},
	})
	if err != nil {
		return err
	}

	return nil
}

func addDeleteTimeToEmail(email string, deletedAt time.Time) string {
	timestamp := deletedAt.Unix() // Convert time to Unix timestamp
	modifiedEmail := fmt.Sprintf("%s_deleted_%d", email, timestamp)
	return modifiedEmail
}

/*
func RunSyncUserCreationTasks(ctx context.Context, qtx *sqlc.Queries, tx *sql.Tx, config utils.Config, td worker.TaskDistributor,
	req RegisterReq, uid uuid.UUID, clientIP string, agent string) (string, time.Time, error) {

	type result struct {
		err error
	}

	tokenService := NewTokenService(config, cache.NewCache(config.RedisAddress, config.RedisUsername, config.RedisPwd, 0))

	// Create access token
	accessToken, accessPayload, err := tokenService.CreateAccessToken(req.Email, req.Username, "", true, false, uid, constants.RegularUsers, clientIP, agent)
	if err != nil {
		tx.Rollback()
		return "", time.Time{}, errors.New("unknown error")
	}

	// Channels to capture results
	profileCh := make(chan result, 1)
	roleCh := make(chan result, 1)
	emailCh := make(chan result, 1)

	// Create user profile sequentially
	go func() {
		profileErr := qtx.CreateUserProfile(ctx, sqlc.CreateUserProfileParams{
			UserID:    uid,
			FirstName: sql.NullString{String: req.FullName, Valid: true},
			LastName:  sql.NullString{String: req.FullName, Valid: true},
		})
		profileCh <- result{err: profileErr}
		close(profileCh)
	}()

	// Wait for user profile creation to complete
	profileResult := <-profileCh
	if profileResult.err != nil {
		tx.Rollback()
		return "", time.Time{}, fmt.Errorf("an unknown error occurred creating profile %v", profileResult.err)
	}

	// Create user role sequentially
	go func() {
		_, roleErr := qtx.CreateUserRole(ctx, sqlc.CreateUserRoleParams{
			UserID: uid,
			RoleID: constants.RegularUsers,
		})
		roleCh <- result{err: roleErr}
		close(roleCh)
	}()

	// Wait for user role creation to complete
	roleResult := <-roleCh
	if roleResult.err != nil {
		tx.Rollback()
		return "", time.Time{}, errors.New("error cannot proceed")
	}

	// Send verification email concurrently (does not need to be within transaction)
	go func() {
		sendEmailErr := SendVerificationEmail(qtx, ctx, td, uid, req.Email)
		emailCh <- result{err: sendEmailErr}
		close(emailCh)
	}()

	// Wait for email sending to complete
	emailResult := <-emailCh
	if emailResult.err != nil {
		tx.Rollback()
		return "", time.Time{}, errors.New("unable to resend email " + emailResult.err.Error())
	}

	return accessToken, accessPayload.Expires, nil
}
*/
