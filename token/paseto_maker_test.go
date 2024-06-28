package token

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/steve-mir/bukka_backend/internal/cache"
	"github.com/stretchr/testify/require"
)

func TestNewPasetoMaker(t *testing.T) {
	accessSymmetricKey := "a_very_secret_key_with_sufficient_length"
	refreshSymmetricKey := "a_very_secret_key_with_sufficient_length"
	maker, err := NewPasetoMaker(accessSymmetricKey, refreshSymmetricKey)
	require.NoError(t, err)
	require.NotNil(t, maker)
}

func TestNewPasetoMakerInvalidKey(t *testing.T) {
	accessSymmetricKey := "short_key"
	refreshSymmetricKey := "short_key"
	maker, err := NewPasetoMaker(accessSymmetricKey, refreshSymmetricKey)
	require.Error(t, err)
	require.Nil(t, maker)
}

func TestCreateToken(t *testing.T) {
	accessSymmetricKey := "a_very_secret_key_with_sufficient_length"
	refreshSymmetricKey := "a_very_secret_key_with_sufficient_length"
	maker, err := NewPasetoMaker(accessSymmetricKey, refreshSymmetricKey)
	require.NoError(t, err)
	require.NotNil(t, maker)

	userID := uuid.New()
	sessionID := uuid.New()
	payloadData := PayloadData{
		Role:          1,
		Subject:       userID,
		Username:      "testuser",
		Email:         "testuser@example.com",
		Phone:         "123-456-7890",
		EmailVerified: true,
		SessionID:     sessionID,
		Issuer:        "yourapp",
		Audience:      "yourapp_users",
		IP:            "192.168.1.1",
		UserAgent:     "Mozilla/5.0",
		MfaPassed:     true,
	}

	duration := time.Hour
	token, payload, err := maker.CreateToken(payloadData, duration, TokenType(AccessToken))
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotNil(t, payload)
}

func TestVerifyToken(t *testing.T) {
	accessSymmetricKey := "a_very_secret_key_with_sufficient_length"
	refreshSymmetricKey := "a_very_secret_key_with_sufficient_length"

	maker, err := NewPasetoMaker(accessSymmetricKey, refreshSymmetricKey)
	require.NoError(t, err)
	require.NotNil(t, maker)

	userID := uuid.New()
	sessionID := uuid.New()
	payloadData := PayloadData{
		Role:          1,
		Subject:       userID,
		Username:      "testuser",
		Email:         "testuser@example.com",
		Phone:         "123-456-7890",
		EmailVerified: true,
		SessionID:     sessionID,
		Issuer:        "yourapp",
		Audience:      "yourapp_users",
		IP:            "192.168.1.1",
		UserAgent:     "Mozilla/5.0",
		MfaPassed:     true,
		TokenType:     TokenType(RefreshToken),
	}

	duration := time.Hour
	token, _, err := maker.CreateToken(payloadData, duration, TokenType(RefreshToken))
	require.NoError(t, err)
	require.NotEmpty(t, token)

	cache := cache.NewCache("0.0.0.0:6379", "default", "", 0)
	payload, err := maker.VerifyToken(context.Background(), *cache, token, TokenType(RefreshToken))
	require.NoError(t, err)
	require.NotNil(t, payload)
	require.Equal(t, payloadData.Subject, payload.Subject)
	require.Equal(t, payloadData.Username, payload.Username)
}

func TestVerifyTokenExpired(t *testing.T) {
	accessSymmetricKey := "a_very_secret_key_with_sufficient_length"
	refreshSymmetricKey := "a_very_secret_key_with_sufficient_length"
	maker, err := NewPasetoMaker(accessSymmetricKey, refreshSymmetricKey)
	require.NoError(t, err)
	require.NotNil(t, maker)

	userID := uuid.New()
	sessionID := uuid.New()
	payloadData := PayloadData{
		Role:          1,
		Subject:       userID,
		Username:      "testuser",
		Email:         "testuser@example.com",
		Phone:         "123-456-7890",
		EmailVerified: true,
		SessionID:     sessionID,
		Issuer:        "yourapp",
		Audience:      "yourapp_users",
		IP:            "192.168.1.1",
		UserAgent:     "Mozilla/5.0",
		MfaPassed:     true,
	}

	duration := -time.Hour // Token duration in the past to simulate expiration
	token, _, err := maker.CreateToken(payloadData, duration, TokenType(AccessToken))
	require.NoError(t, err)
	require.NotEmpty(t, token)

	cache := cache.NewCache("0.0.0.0:6379", "default", "", 0)
	payload, err := maker.VerifyToken(context.Background(), *cache, token, TokenType(AccessToken))
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload) // Ensure payload is nil when the token is expired. require.Nil(t, payload)
}

func TestVerifyTokenInvalidKey(t *testing.T) {
	accessSymmetricKey := "a_very_secret_key_with_sufficient_length"
	refreshSymmetricKey := "a_very_secret_key_with_sufficient_length"

	maker, err := NewPasetoMaker(accessSymmetricKey, refreshSymmetricKey)
	require.NoError(t, err)
	require.NotNil(t, maker)

	userID := uuid.New()
	sessionID := uuid.New()
	payloadData := PayloadData{
		Role:          1,
		Subject:       userID,
		Username:      "testuser",
		Email:         "testuser@example.com",
		Phone:         "123-456-7890",
		EmailVerified: true,
		SessionID:     sessionID,
		Issuer:        "yourapp",
		Audience:      "yourapp_users",
		IP:            "192.168.1.1",
		UserAgent:     "Mozilla/5.0",
		MfaPassed:     true,
	}

	duration := time.Hour
	token, _, err := maker.CreateToken(payloadData, duration, TokenType(AccessToken))
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Create a new maker with a different key
	invalidKey := "a_different_secret_key_with_sufficient_length"
	invalidKey2 := "a_different_secret_key_with_sufficient_length"
	invalidMaker, err := NewPasetoMaker(invalidKey, invalidKey2)
	require.NoError(t, err)
	require.NotNil(t, invalidMaker)

	cache := cache.NewCache("0.0.0.0:6379", "default", "", 0)
	payload, err := invalidMaker.VerifyToken(context.Background(), *cache, token, TokenType(AccessToken))
	require.Error(t, err)
	require.Nil(t, payload)
	require.EqualError(t, err, ErrInvalidToken.Error())
}
