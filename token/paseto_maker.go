package token

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/o1egl/paseto"
	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/internal/cache"
	"golang.org/x/crypto/pbkdf2"
	// "github.com/aead/chacha20poly1305"
)

const minSecretKeySize = 32

type PasetoMaker struct {
	paseto              *paseto.V2
	accessSymmetricKey  []byte
	refreshSymmetricKey []byte
	// cache        cache.Cache
}

func NewPasetoMaker(accessKey, refreshKey string) (Maker, error) {
	// log.Println("Key size", chacha20poly1305.KeySize)
	if len(accessKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d bytes", minSecretKeySize)
	}
	if len(refreshKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d bytes", minSecretKeySize)
	}

	// Key derivation
	accessSymmetricKey := pbkdf2.Key([]byte(accessKey), []byte("paseto-key"), 10000, minSecretKeySize, sha256.New)
	refreshSymmetricKey := pbkdf2.Key([]byte(refreshKey), []byte("paseto-key"), 10000, minSecretKeySize, sha256.New)

	maker := &PasetoMaker{
		paseto:              paseto.NewV2(),
		accessSymmetricKey:  accessSymmetricKey,
		refreshSymmetricKey: refreshSymmetricKey,
		// cache:        *cache,
	}

	return maker, nil

}

// CreateToken implements Maker.
func (maker *PasetoMaker) CreateToken(payloadData PayloadData, duration time.Duration, tokenType TokenType) (string, *Payload, error) {
	payload, err := NewPayload(payloadData, duration)
	if err != nil {
		return "", &Payload{}, err
	}

	var key []byte
	switch tokenType {
	case AccessToken:
		key = maker.accessSymmetricKey
	case RefreshToken:
		key = maker.refreshSymmetricKey
	default:
		return "", nil, fmt.Errorf("invalid token type")
	}

	token, err := maker.paseto.Encrypt(key, payload, nil)
	return token, payload, err
}

func (maker *PasetoMaker) VerifyToken(ctx context.Context, cache cache.Cache, token string, tokenType TokenType) (*Payload, error) {
	payload := &Payload{}

	var key []byte
	switch tokenType {
	case AccessToken:
		key = maker.accessSymmetricKey
	case RefreshToken:
		key = maker.refreshSymmetricKey
	default:
		return nil, fmt.Errorf("invalid token type")
	}

	err := maker.paseto.Decrypt(token, key, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.ValidateExpiry()
	if err != nil {
		return nil, err
	}

	if payload.TokenType != tokenType {
		return nil, ErrInvalidToken
	}

	if payload.TokenType == AccessToken {
		sessionID, err := cache.GetKey(ctx, token)
		if err != nil || sessionID == "" {
			return nil, ErrInvalidToken
		}
	}

	return payload, nil
}

func (maker *PasetoMaker) RevokeTokenAccessToken(token string, ctx context.Context, store sqlc.Store, cache cache.Cache) error {
	payload, err := maker.VerifyToken(ctx, cache, token, TokenType(AccessToken))
	if err != nil {
		return err
	}

	if payload.TokenType == AccessToken {
		err = cache.DeleteKey(ctx, token)
		if err != nil {
			return err
		}
	}

	// Close the session associated with the token
	return store.RevokeSessionById(ctx, payload.SessionID)

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
