package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrExpiredToken   = errors.New("access token has expired")
	RtErrExpiredToken = errors.New("refresh token has expired")
)

// User ID - The unique identifier of the logged in user. This allows fetching the user's attributes like roles/permissions from the database for authorization.
// Session ID - The id of the user's session in the sessions table. Used to validate the session is valid on each request.
// Issued at timestamp - When the token was issued. Used to calculate expiry.
// Expiry timestamp - When the token expires.

// Not before timestamp - Earliest time the token can be used (optional).
// Issuer - Your service's ID, e.g. https://myapp.com.
// Audience - Intended API audience, e.g. https://myapp.com/api.
// IP Address - The IP address of the client issued to. Can be used to detect suspicious usage.
// User agent - The user agent string of the client. Can be used to identify clients.

type PayloadData struct {
	Role          int8      `json:"role"`
	Subject       uuid.UUID `json:"sub"` // subject: the user ID
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	EmailVerified bool      `json:"email_verified"`
	SessionID     uuid.UUID `json:"session_id"`
	Issuer        string    `json:"iss"` // issuer
	Audience      string    `json:"aud"` // audience
	IP            string    `json:"ip"`  // assuming IP is a string for simplicity
	UserAgent     string    `json:"user_agent"`
	MfaPassed     bool      `json:"mfa_passed"`
}

type Payload struct {
	PayloadData
	NotBefore time.Time `json:"nbf"` // not before
	Expires   time.Time `json:"exp"` // expiration time
	IssuedAt  time.Time `json:"iat"` // issued at
}

func NewPayload(payload PayloadData, duration time.Duration) (*Payload, error) {
	now := time.Now()
	return &Payload{
		// ID:           tokenID,
		PayloadData: payload,
		Expires:     time.Now().Add(duration),
		IssuedAt:    now,
		NotBefore:   now,
	}, nil

}

func (payload *Payload) ValidateExpiry() error {
	currentTime := time.Now()
	if currentTime.After(payload.Expires) {
		if payload.SessionID != uuid.Nil {
			return RtErrExpiredToken
		}
		return ErrExpiredToken
	}
	return nil
}
