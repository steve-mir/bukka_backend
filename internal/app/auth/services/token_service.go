package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/steve-mir/bukka_backend/token"
	"github.com/steve-mir/bukka_backend/utils"
)

type TokenService struct {
	config utils.Config
	// other dependencies as needed
}

func NewTokenService(config utils.Config) *TokenService {
	return &TokenService{
		config: config,
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
	maker, err := token.NewPasetoMaker(utils.GetKeyForToken(t.config, false))
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
	maker, err := token.NewPasetoMaker(utils.GetKeyForToken(t.config, true))
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
