package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/internal/app/auth/services"
	"golang.org/x/oauth2"
)

// https://localhost:8080/v1/auth/google/callback
func (server *Server) googleLogin(c *gin.Context) {
	fcmToken := c.Query("fcmToken")
	// url := server.oauthConfig.AuthCodeURL("state") + "&fcmToken=" + url.QueryEscape(fcmToken)
	url := server.oauthConfig.AuthCodeURL("state", oauth2.SetAuthURLParam("fcmToken", fcmToken))
	log.Println("Auth url", url)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// 1. Check if the user exists in the database
// 2. If not, create a new user
// 3. Generate a session token or JWT for the user
// 4. Return the token to the client or redirect to a frontend URL with the token
func (server *Server) googleCallback(c *gin.Context) {
	code := c.Query("code")
	fcmToken := c.Query("fcmToken")
	token, err := server.oauthConfig.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	client := server.oauthConfig.Client(c, token)
	userInfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	defer userInfo.Body.Close()

	var googleUser struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(userInfo.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Check if the user exists
	user, err := server.store.GetUserAndRoleByIdentifier(c, sql.NullString{String: googleUser.Email, Valid: true})
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var userData services.UserAuthRes
	if err == sql.ErrNoRows {
		// User doesn't exist, create a new account
		tx, err := server.db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		defer tx.Rollback()

		qtx := sqlc.New(tx)

		// Generate a random password for the user
		// randomPassword, _ := utils.GenerateUniqueToken(16) //utils.GenerateRandomString(16)
		_, uid, err := services.PrepareUserData("")
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		// Create the user
		sqlcUser, err := services.CreateUserConcurrent(c, qtx, uid, googleUser.Email, googleUser.Name, "", true, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		// Run concurrent operations (you may need to modify this to fit Google sign-in specifics)
		clientIP := c.ClientIP()
		agent := c.Request.UserAgent()
		accessToken, accessExp, err := services.RunConcurrentUserCreationTasks(c, server.tokenMaker, qtx, tx, server.config, server.taskDistributor, services.RegisterReq{Email: googleUser.Email, Username: googleUser.Name, FcmToken: fcmToken}, uid, clientIP, agent, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		userData = services.UserAuthRes{
			Uid:             sqlcUser.ID,
			Username:        sqlcUser.Username.String,
			Email:           sqlcUser.Email,
			IsEmailVerified: true, // Since it's Google-verified
			CreatedAt:       sqlcUser.CreatedAt.Time,
			AuthToken: services.AuthToken{
				AccessToken:          accessToken,
				AccessTokenExpiresAt: accessExp,
			},
		}
	} else {
		// Check if the user is an OAuth user
		if !user.IsOauthUser.Bool {
			c.JSON(http.StatusInternalServerError, errorResponse(errors.New("this account is not linked with OAuth")))
		}

		// User exists, log them in
		clientIP := c.ClientIP()
		agent := c.Request.UserAgent()
		var err error
		// TODO: Get fcm token
		userData, err = services.LogOAuthUserIn(services.LoginReq{Identifier: googleUser.Email, FcmToken: fcmToken}, *server.tokenService, server.store, c, server.config, clientIP, agent)
		if err != nil {
			c.JSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
	}

	c.JSON(http.StatusOK, userData)
}
