package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type TokenType string
type AuthMethod string

const (
	AccessToken           TokenType  = "access"
	RefreshToken          TokenType  = "refresh"
	AuthEmailPassword     AuthMethod = "email_password"
	AuthPhone             AuthMethod = "phone"
	AuthGoogle            AuthMethod = "google"
	AuthApple             AuthMethod = "apple"
	DefaultNotBeforeDelay            = 15 * time.Minute
)

var (
	ErrInvalidToken        = errors.New("invalid token")
	ErrExpiredToken        = errors.New("access token has expired")
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
	ErrTokenNotYetValid    = errors.New("token is not yet valid")
)

type PayloadData struct {
	Role          int8      `json:"role"`
	Subject       uuid.UUID `json:"sub"` // subject: the user ID
	Username      string    `json:"username,omitempty"`
	Email         string    `json:"email,omitempty"`
	Phone         string    `json:"phone,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	SessionID     uuid.UUID `json:"session_id,omitempty"` // session ID is optional
	Issuer        string    `json:"iss"`                  // issuer
	Audience      string    `json:"aud"`                  // audience
	IP            string    `json:"ip"`                   // assuming IP is a string for simplicity
	UserAgent     string    `json:"user_agent"`
	MfaPassed     bool      `json:"mfa_passed"`
	TokenType     TokenType `json:"token_type"` // "access" or "refresh"

	PhoneVerified bool       `json:"phone_verified"` // TODO: Add to db
	DeviceID      string     `json:"device_id"`      // unique identifier for the device
	Platform      string     `json:"platform"`       // "android", "ios", "web", "desktop"
	OSVersion     string     `json:"os_version"`     // OS version of the device
	AppVersion    string     `json:"app_version"`    // version of your app
	AuthMethod    AuthMethod `json:"auth_method"`    // "email", "phone", "google", "apple"
	// Scope         []string  `json:"scope"`       // array of permission scopes, []string{"read", "write"},
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
		PayloadData: payload,
		Expires:     time.Now().Add(duration),
		IssuedAt:    now,
		NotBefore:   now.Add(DefaultNotBeforeDelay),
	}, nil

}

func (payload *Payload) ValidateExpiry() error {
	currentTime := time.Now()
	if payload.TokenType == RefreshToken && currentTime.Before(payload.NotBefore) {
		return ErrTokenNotYetValid
	}

	if currentTime.After(payload.Expires) {
		if payload.TokenType == RefreshToken {
			return ErrRefreshTokenExpired
		}
		return ErrExpiredToken
	}
	return nil
}
