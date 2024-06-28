package services

import (
	"time"

	"github.com/google/uuid"
)

// Requests
type LoginReq struct {
	Identifier string `json:"identifier" binding:"required"`
	FcmToken   string `json:"fcm_token"`
	Password   string `json:"password" binding:"required,passwordValidator"`
}

type VerifyEmailReq struct {
	Token string `json:"token" binding:"required"` // TODO: Add verification to allow only digits within the appropriate length
}

type DeleteAccountReq struct {
	Password string `json:"password" binding:"required,passwordValidator"`
}

type AccountRecoveryReq struct {
	Email string `json:"email" binding:"required,emailValidator"`
}

type ChangePwdReq struct {
	OldPassword string `json:"old_password" binding:"required,passwordValidator"`
	NewPassword string `json:"new_password" binding:"required,passwordValidator"`
}

type ResetPwdReq struct {
	Token       string `json:"token" binding:"required"` // TODO: Add validation for token
	NewPassword string `json:"new_password" binding:"required,passwordValidator"`
}

// ! Responses
type UserAuthRes struct {
	Uid               uuid.UUID `json:"uid"`
	IsEmailVerified   bool      `json:"is_email_verified"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	CreatedAt         time.Time `json:"created_at"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	IsSuspended       bool      `json:"is_suspended"`
	IsMfaEnabled      bool      `json:"is_mfa_enabled"`
	IsDeleted         bool      `json:"is_deleted"`
	ImageUrl          string    `json:"image_url"`
	AuthToken
}

type VerifyEmailRes struct {
	Msg      string `json:"msg"`
	Verified bool   `json:"verified"`
}

type GenericRes struct {
	Msg string `json:"msg"`
}

type HomeRes struct {
	Msg string `json:"msg"`
}

type UserProfileRes struct {
	Uid        string    `json:"uid"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	ImageUrl   string    `json:"image_url"`
	CreatedAt  time.Time `json:"created_at"`
	IsVerified bool      `json:"is_verified"`
}
