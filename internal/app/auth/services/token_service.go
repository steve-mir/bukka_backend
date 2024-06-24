package services

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/internal/cache"
	"github.com/steve-mir/bukka_backend/token"
	"github.com/steve-mir/bukka_backend/utils"
)

type TokenService struct {
	config utils.Config
	cache  cache.Cache
	// other dependencies as needed
}

func NewTokenService(config utils.Config, cache *cache.Cache) *TokenService {
	return &TokenService{
		config: config,
		cache:  *cache,
	}
}

type AuthToken struct {
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}

func NewAuthToken(accessToken string, refreshToken string, accessTokenExpiresAt time.Time, refreshTokenExpiresAt time.Time) *AuthToken {
	return &AuthToken{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
	}
}

func (t *TokenService) CreateAccessToken(email, username, phone string, mfaPassed, isEmailVerified bool, userId uuid.UUID, role int8,
	ip, userAgent string,
) (string, *token.Payload, error) {

	// Create a Paseto token and include user data in the payload
	maker, err := token.NewPasetoMaker(utils.GetKeyForToken(t.config, false), &t.cache)
	if err != nil {
		return "", &token.Payload{}, err
	}

	// Define the payload for the token (excluding the password)
	payloadData := token.PayloadData{
		Role:          role,
		Subject:       userId,
		Username:      username,
		Email:         email,
		Phone:         phone,
		EmailVerified: isEmailVerified,
		Issuer:        t.config.AppName,
		Audience:      "website users",
		IP:            ip,
		UserAgent:     userAgent,
		MfaPassed:     mfaPassed,
	}

	// Create the Paseto token
	pToken, payload, err := maker.CreateToken(payloadData, t.config.AccessTokenDuration) // Set the token expiration as needed
	return pToken, payload, err
}

func (t *TokenService) CreateRefreshToken(userId uuid.UUID, sessionID uuid.UUID, ip, userAgent string,
) (string, *token.Payload, error) {

	// Create a Paseto token and include user data in the payload
	maker, err := token.NewPasetoMaker(utils.GetKeyForToken(t.config, true), &t.cache)
	if err != nil {
		return "", &token.Payload{}, err
	}
	// Define the payload for the token (excluding the password)
	payloadData := token.PayloadData{
		Subject:   userId,
		SessionID: sessionID,
		Issuer:    t.config.AppName,
		Audience:  "website users",
		IP:        ip,
		UserAgent: userAgent,
	}

	// Create the Paseto token
	pToken, payload, err := maker.CreateToken(payloadData, t.config.RefreshTokenDuration) // Set the token expiration as needed
	return pToken, payload, err
}

func VerifyToken(tokenMaker token.Maker, token string) (*token.Payload, error) {
	payload, err := tokenMaker.VerifyToken(token)
	if err != nil {
		return nil, err
	}

	return payload, nil

}

func (t *TokenService) RotateToken(email, username, phone string, mfaPassed, isEmailVerified bool, userId uuid.UUID, role int8, sessionID uuid.UUID, clientIP, userAgent string,
	config utils.Config, store sqlc.Store,
) (AuthToken, error) {
	// Create a channel to receive token creation results.
	type tokenResult struct {
		token   string
		payload *token.Payload // Assuming TokenPayload is the type of accessPayload and refreshPayload
		err     error
	}

	// Define a WaitGroup to wait for goroutines to finish.
	var wg sync.WaitGroup

	// Start goroutine to create refresh token.
	refreshTokenCh := make(chan tokenResult, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		refreshToken, refreshPayload, err := t.CreateRefreshToken(userId, sessionID, clientIP, userAgent)
		refreshTokenCh <- tokenResult{refreshToken, refreshPayload, err}
	}()

	// Start goroutine to create access token.
	accessTokenCh := make(chan tokenResult, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		accessToken, accessPayload, err := t.CreateAccessToken(email, username, phone, mfaPassed, isEmailVerified, userId, role, clientIP, userAgent)
		accessTokenCh <- tokenResult{accessToken, accessPayload, err}
	}()

	// Wait for the token creation goroutines to finish.
	wg.Wait()
	close(refreshTokenCh)
	close(accessTokenCh)

	// Collect the results.
	refreshResult := <-refreshTokenCh
	if refreshResult.err != nil {
		return AuthToken{}, refreshResult.err
	}

	accessResult := <-accessTokenCh
	if accessResult.err != nil {
		return AuthToken{}, accessResult.err
	}

	// Rotate session tokens in the database.
	err := store.RotateSessionTokens(context.Background(), sqlc.RotateSessionTokensParams{
		ID:              sessionID,
		RefreshToken:    refreshResult.token,
		RefreshTokenExp: refreshResult.payload.Expires,
	})
	if err != nil {
		return AuthToken{}, err
	}

	return AuthToken{
		AccessToken:           accessResult.token,
		RefreshToken:          refreshResult.token,
		AccessTokenExpiresAt:  accessResult.payload.Expires,
		RefreshTokenExpiresAt: refreshResult.payload.Expires,
	}, nil
}
