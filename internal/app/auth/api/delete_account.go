package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/steve-mir/bukka_backend/internal/app/auth/middlewares"
	"github.com/steve-mir/bukka_backend/internal/app/auth/services"
	"github.com/steve-mir/bukka_backend/token"
)

func (s *Server) deleteAccount(ctx *gin.Context) {
	authPayload := ctx.MustGet(middlewares.AuthorizationPayloadKey).(*token.Payload)

	var req services.DeleteAccountReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := services.DeleteAccountRequest(ctx, req.Password, s.store, authPayload.Subject)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, services.GenericRes{
		Msg: "Account deleted successfully",
	})

}

func (s *Server) requestAccountRecovery(ctx *gin.Context) {
	var req services.AccountRecoveryReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := services.AccRecoveryRequest(ctx, s.store, s.taskDistributor, req.Email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, services.GenericRes{
		Msg: "URL has been sent to your email",
	})

}

func (s *Server) completeAccountRecovery(ctx *gin.Context) {
	token := ctx.Query("token")
	err := services.AccountRecovery(ctx, s.db, s.store, token)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, services.GenericRes{
		Msg: "Account recovered, you can now login to access your account.",
	})

}
