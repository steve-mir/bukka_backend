package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/steve-mir/bukka_backend/token"
	"github.com/steve-mir/bukka_backend/utils"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
)

var (
	AuthorizationPayloadKey = "authorization_payload"
)

func AuthMiddlerWare(config utils.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		accessToken := fields[1]
		tokenMaker, err := token.NewPasetoMaker(utils.GetKeyForToken(config, false))
		if err != nil {
			err := fmt.Errorf("could not init tokenMaker %s", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		payload, err := tokenMaker.VerifyToken(accessToken) // Decrypts the access token and returns the data stored in it
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Check if email is verified (remove if it is optional to verify email)
		// TODO: fix for 1 login or unverified emails
		// if !payload.IsUserVerified {
		// 	fmt.Println("Error: Please verify your account")
		// 	ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "account not verified"})
		// 	return
		// }

		ctx.Set(AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}
