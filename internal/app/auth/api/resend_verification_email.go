package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/steve-mir/bukka_backend/internal/app/auth/middlewares"
	"github.com/steve-mir/bukka_backend/internal/app/auth/services"
	"github.com/steve-mir/bukka_backend/token"
)

func (s *Server) resendVerificationEmail(ctx *gin.Context) {
	authPayload := ctx.MustGet(middlewares.AuthorizationPayloadKey).(*token.Payload)
	err := services.ReSendVerificationEmail(s.store, ctx, s.taskDistributor, authPayload.Subject, authPayload.Email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, services.VerifyEmailRes{
		Msg:      "Verification email sent",
		Verified: true,
	})

}
