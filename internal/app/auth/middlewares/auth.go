package middlewares

import (
	"errors"
	"fmt"
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

// I want to add authorization to this auth middleware. The idea is that endpoint will have the permissions assigned to it. EG
// get_menu endpoint might have only 1,2 roles. And will be the accessibleRoles. Refactor this to be able to fit the narrative
// Also note the full method should be used to index it so as to get the permissions(roles)
func AuthMiddleWare(config utils.Config, tokenMaker token.Maker, cache cache.Cache, accessibleRoles map[string][]int8) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fullMethod := fmt.Sprintf("%s %s", ctx.Request.Method, ctx.FullPath())
		roles, ok := accessibleRoles[fullMethod]
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sorry you do not have enough permissions to access: " + fullMethod})
			return
		}

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
		if !payload.EmailVerified {
			fmt.Println("Error: Please verify your account")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "account not verified"})
			return
		}

		if !isAuthorized(payload.Role, roles) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "you don't have permission to access this resource"})
			return
		}

		ctx.Set(AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}

func isAuthorized(userRole int8, allowedRoles []int8) bool {
	for _, role := range allowedRoles {
		if userRole == role {
			return true
		}
	}
	return false
}
