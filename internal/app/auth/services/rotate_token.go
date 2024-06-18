package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/token"
	"github.com/steve-mir/bukka_backend/utils"
)

type RotateTokenReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func RotateUserToken(req RotateTokenReq, store sqlc.Store, ctx context.Context, config utils.Config, clientIP, agent string) (AuthTokenResp, error) {
	tokenMaker, err := token.NewPasetoMaker(config.RefreshTokenSymmetricKey)
	if err != nil {
		return AuthTokenResp{}, fmt.Errorf("cannot create token maker: %v", err)
	}

	payload, err := VerifyToken(tokenMaker, req.RefreshToken)
	if err != nil {
		return AuthTokenResp{}, fmt.Errorf("token verification failed: %v", err)
	}

	err = checkUserStatus(ctx, store, payload.Subject)
	if err != nil {
		return AuthTokenResp{}, err
	}

	session, err := store.GetSessionAndUserByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			if blockErr := blockUser(ctx, store, payload.Subject); blockErr != nil {
				return AuthTokenResp{}, blockErr
			}
			return AuthTokenResp{}, errors.New("suspicious activity detected")
		}
		return AuthTokenResp{}, fmt.Errorf("failed to get session: %v", err)
	}

	if !session.BlockedAt.Time.IsZero() && (session.BlockedAt.Time.Before(time.Now()) || session.InvalidatedAt.Time.Before(time.Now())) {
		return AuthTokenResp{}, errors.New("session blocked")
	}

	// ip := utils.GetIpAddr(clientIP)

	authToken, err := NewTokenService(config).
		RotateToken(session.Email, session.Username.String, session.Phone.String, true, session.IsEmailVerified.Bool, payload.Subject,
			int8(session.RoleID.Int32), session.ID, clientIP, agent, config, store)

	if err != nil {
		return AuthTokenResp{}, fmt.Errorf("could not rotate token: %v", err)
	}

	return AuthTokenResp{

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
