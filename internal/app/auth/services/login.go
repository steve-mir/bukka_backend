package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/netip"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/steve-mir/bukka_backend/constants"
	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/utils"
)

type LoginReq struct {
	Identifier string `json:"identifier" binding:"required"`
	FcmToken   string `json:"fcm_token"`
	Password   string `json:"password" binding:"required,passwordValidator"`
}

func LogUserIn(req LoginReq, store sqlc.Store, ctx context.Context, config utils.Config, clientIP, agent string) (UserAuthRes, error) {
	if err := validateLoginUserRequest(req.Identifier); err != nil {
		return UserAuthRes{}, err
	}

	sessionID, err := uuid.NewRandom()
	if err != nil {
		return UserAuthRes{}, fmt.Errorf("error creating uid %s", err)
	}

	user, err := store.GetUserAndRoleByIdentifier(ctx, pgtype.Text{String: req.Identifier, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			return UserAuthRes{}, errors.New(constants.LoginError)
		}
		return UserAuthRes{}, errors.New(constants.LoginError)
	}

	err = utils.CheckPassword(req.Password, user.PasswordHash)
	if err != nil {
		return UserAuthRes{}, errors.New(constants.LoginError)
	}

	// Check if user should gain access
	err = checkAccountStat(user.IsSuspended.Bool, user.IsDeleted.Bool)
	if err != nil {
		return UserAuthRes{}, errors.New(constants.LoginError)
	}

	var mfaPassed bool
	if user.IsMfaEnabled.Bool {
		mfaPassed = false
	} else {
		mfaPassed = true
	}

	tokenService := NewTokenService(config)
	// Refresh token
	refreshToken, refreshPayload, err := tokenService.CreateRefreshToken(user.ID, sessionID, clientIP, agent)
	if err != nil {
		return UserAuthRes{}, fmt.Errorf("error creating refresh token %s", err)
	}

	// Access token
	accessToken, accessPayload, err := tokenService.CreateAccessToken(user.Email, user.Username.String, user.Phone.String, mfaPassed,
		user.IsEmailVerified.Bool, user.ID, int8(user.RoleID), clientIP, agent)
	if err != nil {
		return UserAuthRes{}, fmt.Errorf("error creating access token %s", err)
	}

	ip, err := utils.GetIpAddr(clientIP)
	if err != nil {
		return UserAuthRes{}, fmt.Errorf("error getting ip address %s", err)
	}

	_, err = store.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:              sessionID,
		UserID:          user.ID,
		RefreshToken:    refreshToken,
		RefreshTokenExp: refreshPayload.Expires,
		UserAgent:       agent,
		IpAddress:       *ip,
		FcmToken:        pgtype.Text{String: req.FcmToken, Valid: true},
	})

	if err != nil {
		log.Println("Session ID Error", err)
		return UserAuthRes{}, fmt.Errorf("error creating session id %s", err)
	}

	//! 3 User logged in successfully. Record it
	err = recordLoginSuccess(ctx, store, user.ID, agent, ip)
	if err != nil {
		return UserAuthRes{}, fmt.Errorf("error creating login record %s", err)
	}

	// return resp
	return UserAuthRes{
		Uid:                   user.ID,
		IsEmailVerified:       user.IsEmailVerified.Bool,
		Username:              user.Username.String,
		Email:                 user.Email,
		IsDeleted:             user.IsDeleted.Bool,
		IsSuspended:           user.IsSuspended.Bool,
		IsMfaEnabled:          user.IsMfaEnabled.Bool,
		ImageUrl:              user.ImageUrl.String,
		CreatedAt:             user.CreatedAt.Time,
		PasswordChangedAt:     user.PasswordLastChanged.Time,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.Expires,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.Expires,
	}, nil
}

func validateLoginUserRequest(identifier string) error {
	if utils.IsEmailFormat(identifier) { // Assuming there's a function to check if the format is an email
		if ok := utils.ValidateEmail(identifier); !ok {
			return errors.New(constants.InvalidEmail)
		}
	} else if utils.IsPhoneFormat(identifier) {
		if !utils.ValidatePhone(identifier) {
			return errors.New(constants.InvalidPhone)
		}
	} else { // Default to username validation if it's not email or phone
		if !utils.ValidateUsername(identifier) {
			return errors.New(constants.InvalidUsername)
		}
	}

	return nil
}

func checkAccountStat(isSuspended bool, isDeleted bool) error {
	fmt.Printf("Is Suspended %v is deleted %v", isSuspended, isDeleted)
	if isSuspended {
		log.Println("Account deleted: ", isSuspended)
		return errors.New("account suspended")
	}

	// Check if user should gain access
	if isDeleted {
		log.Println("Account deleted: ", isDeleted)
		return errors.New(constants.LoginError)
	}
	return nil
}

func recordLoginSuccess(ctx context.Context, dbStore sqlc.Store, userId uuid.UUID, userAgent string, ipAddrs *netip.Addr) error {

	_, err := dbStore.CreateUserLogin(ctx, sqlc.CreateUserLoginParams{
		UserID: userId,
		UserAgent: pgtype.Text{
			String: userAgent,
			Valid:  true,
		},
		IpAddress: ipAddrs,
	})
	return err
}