package middlewares

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/steve-mir/bukka_backend/internal/cache"
	"github.com/steve-mir/bukka_backend/token"
	"github.com/steve-mir/bukka_backend/utils"
)

const (
	AuthorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
)

var (
	AuthorizationPayloadKey = "authorization_payload"
)

func AuthMiddlerWare(config utils.Config, tokenMaker token.Maker, cache cache.Cache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(AuthorizationHeaderKey)
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

		payload, err := tokenMaker.VerifyToken(ctx, cache, accessToken, token.AccessToken) // Decrypts the access token and returns the data stored in it
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

		log.Println("Good to go")
		ctx.Set(AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}
