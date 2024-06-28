package services

import (
	"context"

	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/internal/cache"
	"github.com/steve-mir/bukka_backend/token"
	"github.com/steve-mir/bukka_backend/utils"
	"golang.org/x/sync/errgroup"
)

type TokenService struct {
	config     utils.Config
	cache      cache.Cache
	tokenMaker token.Maker
	// other dependencies as needed
}

func NewTokenService(config utils.Config, cache *cache.Cache, tokenMaker token.Maker) *TokenService {
	return &TokenService{
		config:     config,
		cache:      *cache,
		tokenMaker: tokenMaker,
	}
}

func (s *TokenService) CreateTokenPair(ctx context.Context, payloadData token.PayloadData) (AuthToken, error) {
	var eg errgroup.Group
	var err error
	var accessToken, refreshToken string
	var accessPayload, refreshPayload *token.Payload

	eg.Go(func() error {
		accessToken, accessPayload, err = s.tokenMaker.CreateToken(token.PayloadData{
			Role:          payloadData.Role,
			Subject:       payloadData.Subject,
			Username:      payloadData.Username,
			Email:         payloadData.Email,
			Phone:         payloadData.Phone,
			EmailVerified: payloadData.EmailVerified,
			Issuer:        payloadData.Issuer,
			Audience:      "website users",
			IP:            payloadData.IP,
			UserAgent:     payloadData.UserAgent,
			MfaPassed:     payloadData.MfaPassed,
			SessionID:     payloadData.SessionID,
			TokenType:     token.TokenType(token.AccessToken),
		},
			s.config.AccessTokenDuration, token.TokenType(token.AccessToken))
		if err != nil {
			return err
		}

		return nil
	})

	eg.Go(func() error {
		refreshToken, refreshPayload, err = s.tokenMaker.CreateToken(
			token.PayloadData{
				Subject:   payloadData.Subject,
				SessionID: payloadData.SessionID,
				Issuer:    payloadData.Issuer,
				Audience:  "website users",
				IP:        payloadData.IP,
				UserAgent: payloadData.UserAgent,
				TokenType: token.TokenType(token.RefreshToken),
			},
			s.config.RefreshTokenDuration, token.TokenType(token.RefreshToken))

		if err != nil {
			return err
		}

		return nil
	})

	// Wait for both goroutines to complete
	if err := eg.Wait(); err != nil {
		return AuthToken{}, err
	}

	err = s.cache.SetKey(ctx, accessToken, payloadData.SessionID.String(), s.config.AccessTokenDuration)
	if err != nil {
		return AuthToken{}, err
	}

	return *NewAuthToken(accessToken, refreshToken, accessPayload.Expires, refreshPayload.Expires), nil
	// return accessToken, accessPayload, refreshToken, refreshPayload, nil
}

func (s *TokenService) RotateTokens(ctx context.Context, refreshToken string, store sqlc.Store, cache cache.Cache) (AuthToken, error) {
	payload, err := s.tokenMaker.VerifyToken(ctx, cache, refreshToken, token.TokenType(token.RefreshToken))
	if err != nil {
		return AuthToken{}, err
	}

	if payload.TokenType != token.RefreshToken {
		return AuthToken{}, token.ErrInvalidToken
	}

	authToken, err := s.CreateTokenPair(ctx, payload.PayloadData)
	if err != nil {
		return AuthToken{}, err
	}

	// Rotate session tokens in the database.
	err = store.RotateSessionTokens(context.Background(), sqlc.RotateSessionTokensParams{
		ID:              payload.SessionID,
		RefreshToken:    authToken.RefreshToken,
		RefreshTokenExp: authToken.RefreshTokenExpiresAt,
	})
	if err != nil {
		return AuthToken{}, err
	}

	return authToken, nil
}
