package token

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"github.com/redis/go-redis/v9"
	"github.com/steve-mir/bukka_backend/internal/cache"
	"golang.org/x/crypto/pbkdf2"
)

const minSecretKeySize = 32

type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
	cache        cache.Cache
}

func NewPasetoMaker(symmetricKey string, cache *cache.Cache) (Maker, error) {
	if len(symmetricKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d bytes", minSecretKeySize)
	}

	// Key derivation
	key := pbkdf2.Key([]byte(symmetricKey), []byte("paseto-key"), 10000, minSecretKeySize, sha256.New)

	maker := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: key,
		cache:        *cache,
	}

	return maker, nil

}

// CreateToken implements Maker.
func (maker *PasetoMaker) CreateToken(payloadData PayloadData, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(payloadData, duration)

	if err != nil {
		return "", &Payload{}, err
	}

	accessToken, err := maker.paseto.Encrypt(maker.symmetricKey, payload, nil)

	// If token is access token cache it.
	if payload.SessionID == uuid.Nil {
		err = maker.cache.SetKey(context.Background(), accessToken, "active", duration)
		if err != nil {
			return "", &Payload{}, err
		}
	}

	return accessToken, payload, err
}

// VerifyToken implements Maker.
func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// ! Check cache, if token is not in cache then the user might have logged out.
	if payload.SessionID == uuid.Nil {
		ok, err := maker.IsTokenActive(token)
		if err != nil {
			return nil, ErrExpiredToken
		}
		if !ok {
			return nil, ErrExpiredToken
		}
	}

	// Check for expiry if is refresh token
	err = payload.ValidateExpiry()
	if err != nil {
		return payload, err
	}

	return payload, nil
}

// IsTokenActive checks if the token is present and active in the cache
func (maker *PasetoMaker) IsTokenActive(token string) (bool, error) {
	val, err := maker.cache.GetKey(context.Background(), token)
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return val == "active", nil
}

// Add a revoke token. Used for logout
func RevokeToken(cache *cache.Cache, token string) error {
	return cache.DeleteKey(context.Background(), token)
}

// TODO: Add telemetry

// package stats

// func RecordTokenIssued() {
//   // record token issued metrics
// }

// func RecordTokenInvalid() {
//   // record invalid token metrics
// }

// stats.RecordTokenIssued()
// stats.RecordTokenInvalid()
