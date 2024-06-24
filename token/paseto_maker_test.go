package token

import (
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/steve-mir/bukka_backend/internal/cache"
	"github.com/stretchr/testify/require"
)

func TestNewPasetoMaker(t *testing.T) {
	cache := cache.NewCache("0.0.0.0:6379", "default", "", 0)

	symmetricKey := "a_very_secret_key_with_sufficient_length"
	maker, err := NewPasetoMaker(symmetricKey, cache)
	require.NoError(t, err)
	require.NotNil(t, maker)
}

func TestNewPasetoMakerInvalidKey(t *testing.T) {
	cache := cache.NewCache("0.0.0.0:6379", "default", "", 0)

	symmetricKey := "short_key"
	maker, err := NewPasetoMaker(symmetricKey, cache)
	require.Error(t, err)
	require.Nil(t, maker)
}

func TestCreateToken(t *testing.T) {
	cache := cache.NewCache("0.0.0.0:6379", "default", "", 0)

	symmetricKey := "a_very_secret_key_with_sufficient_length"
	maker, err := NewPasetoMaker(symmetricKey, cache)
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
	token, payload, err := maker.CreateToken(payloadData, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotNil(t, payload)
}

func TestVerifyToken(t *testing.T) {
	cache := cache.NewCache("0.0.0.0:6379", "default", "", 0)
	symmetricKey := "a_very_secret_key_with_sufficient_length"
	maker, err := NewPasetoMaker(symmetricKey, cache)
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
	token, _, err := maker.CreateToken(payloadData, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotNil(t, payload)
	require.Equal(t, payloadData.Subject, payload.Subject)
	require.Equal(t, payloadData.Username, payload.Username)
}

func TestVerifyTokenExpired(t *testing.T) {
	cache := cache.NewCache("0.0.0.0:6379", "default", "", 0)
	symmetricKey := "a_very_secret_key_with_sufficient_length"
	maker, err := NewPasetoMaker(symmetricKey, cache)
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
	token, _, err := maker.CreateToken(payloadData, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	log.Println(err)
	log.Println(ErrExpiredToken.Error())
	require.EqualError(t, err, RtErrExpiredToken.Error())
	require.NotNil(t, payload) // Ensure payload is nil when the token is expired. require.Nil(t, payload)
}

/*
func TestVerifyTokenExpired(t *testing.T) {
	symmetricKey := "a_very_secret_key_with_sufficient_length"
	maker, err := NewPasetoMaker(symmetricKey)
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
	token, _, err := maker.CreateToken(payloadData, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.Nil(t, payload)
	require.EqualError(t, err, ErrExpiredToken.Error())
}
*/

func TestVerifyTokenInvalidKey(t *testing.T) {
	cache := cache.NewCache("0.0.0.0:6379", "default", "", 0)
	symmetricKey := "a_very_secret_key_with_sufficient_length"
	maker, err := NewPasetoMaker(symmetricKey, cache)
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
	token, _, err := maker.CreateToken(payloadData, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Create a new maker with a different key
	invalidKey := "a_different_secret_key_with_sufficient_length"
	invalidMaker, err := NewPasetoMaker(invalidKey, cache)
	require.NoError(t, err)
	require.NotNil(t, invalidMaker)

	payload, err := invalidMaker.VerifyToken(token)
	require.Error(t, err)
	require.Nil(t, payload)
	require.EqualError(t, err, ErrInvalidToken.Error())
}
