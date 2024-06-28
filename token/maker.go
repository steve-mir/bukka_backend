package token

import (
	"context"
	"time"

	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/internal/cache"
)

type Maker interface {
	// CreateToken creates a new token for a specific username and duration
	CreateToken(payloadData PayloadData, duration time.Duration, tokenType TokenType) (string, *Payload, error)

	// VerifyToken checks if the token is valid or not
	VerifyToken(ctx context.Context, cache cache.Cache, token string, tokenType TokenType) (*Payload, error)

	// Add a revoke endpoint
	RevokeTokenAccessToken(token string, ctx context.Context, store sqlc.Store, cache cache.Cache) error
}
