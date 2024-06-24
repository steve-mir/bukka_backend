package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/steve-mir/bukka_backend/internal/app/auth/middlewares"
	"github.com/steve-mir/bukka_backend/token"
)

func (s *Server) logout(ctx *gin.Context) {

	authPayload := ctx.MustGet(middlewares.AuthorizationPayloadKey).(*token.Payload)

	authorizationHeader := ctx.GetHeader(middlewares.AuthorizationHeaderKey)
	if len(authorizationHeader) == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is not provided"})
		return
	}

	fields := strings.Fields(authorizationHeader)
	if len(fields) < 2 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
		return
	}

	accessToken := fields[1]
	token.RevokeToken(s.cache, accessToken)
	// log.Println("ID", authPayload.SessionID)

	err := s.store.RevokeSessionById(ctx, authPayload.SessionID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "successfully logged out " + authPayload.Email})

}
