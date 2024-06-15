package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/steve-mir/bukka_backend/db/sqlc"
)

type AuthTokenResp struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

type UserAuthRes struct {
	Uid               uuid.UUID `json:"uid"`
	IsEmailVerified   bool      `json:"is_email_verified"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	CreatedAt         time.Time `json:"created_at"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	IsSuspended       bool      `json:"is_suspended"`
	IsMfaEnabled      bool      `json:"is_mfa_enabled"`
	IsDeleted         bool      `json:"is_deleted"`
	ImageUrl          string    `json:"image_url"`
	AuthTokenResp
}

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
