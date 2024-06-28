package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/internal/cache"
	"github.com/steve-mir/bukka_backend/token"
	"github.com/steve-mir/bukka_backend/utils"
)

type RotateTokenReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func RotateUserToken(req RotateTokenReq, ts *TokenService, cache cache.Cache, store sqlc.Store, tokenMaker token.Maker, ctx context.Context, config utils.Config, clientIP, agent string) (AuthToken, error) {

	payload, err := tokenMaker.VerifyToken(ctx, cache, req.RefreshToken, token.RefreshToken)
	if err != nil {
		return AuthToken{}, fmt.Errorf("token verification failed: %v", err)
	}

	err = checkUserStatus(ctx, store, payload.Subject)
	if err != nil {
		return AuthToken{}, err
	}

	session, err := store.GetSessionAndUserByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			if blockErr := blockUser(ctx, store, payload.Subject); blockErr != nil {
				return AuthToken{}, blockErr
			}
			return AuthToken{}, errors.New("suspicious activity detected")
		}
		return AuthToken{}, fmt.Errorf("failed to get session: %v", err)
	}

	if !session.BlockedAt.Time.IsZero() {
		return AuthToken{}, errors.New("session blocked")
	}
	if !session.InvalidatedAt.Time.IsZero() {
		return AuthToken{}, errors.New("user not logged in")
	}

	authToken, err := ts.RotateTokens(ctx, req.RefreshToken, store, cache)

	if err != nil {
		return AuthToken{}, fmt.Errorf("could not rotate token: %v", err)
	}

	return AuthToken{
		AccessToken:           authToken.AccessToken,
		RefreshToken:          authToken.RefreshToken,
		AccessTokenExpiresAt:  authToken.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: authToken.RefreshTokenExpiresAt,
	}, nil
}

func checkUserStatus(ctx context.Context, store sqlc.Store, uid uuid.UUID) error {
	user, err := store.GetUserByID(ctx, uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %v", err)
	}

	if user.IsDeleted.Bool {
		return errors.New("account not found")
	}

	if user.IsSuspended.Bool {
		return errors.New("account suspended")
	}
	return nil
}

func blockUser(ctx context.Context, store sqlc.Store, userId uuid.UUID) error {
	// Create a channel to receive error results.
	errCh := make(chan error, 2) // Buffer size of 2 since we have two concurrent operations.
	var wg sync.WaitGroup

	// Start goroutine to block the user.
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := store.BlockUser(ctx, userId)
		errCh <- err
	}()

	// Start goroutine to block all user sessions.
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := store.BlockAllUserSession(ctx, userId)
		errCh <- err
	}()

	// Wait for the goroutines to finish.
	wg.Wait()
	close(errCh)

	// Check for errors.
	for err := range errCh {
		if err != nil {
			return fmt.Errorf("block operation error: %v", err)
		}
	}

	return nil
}
