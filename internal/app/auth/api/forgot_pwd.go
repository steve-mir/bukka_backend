package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/steve-mir/bukka_backend/constants"
	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/internal/app/auth/services"
)

func (s *Server) forgotPwd(ctx *gin.Context) {

	var req services.AccountRecoveryReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := services.RequestPwdReset(ctx, req.Email, s.store, s.taskDistributor)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, services.GenericRes{
		Msg: constants.ResetMsg,
	})

}

func (s *Server) resetPwd(ctx *gin.Context) {

	var req services.ResetPwdReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Begin transaction
	tx, err := s.db.Begin()

	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	defer tx.Rollback()

	qtx := sqlc.New(tx)

	err = services.ResetPassword(ctx, qtx, tx, s.store, req.Token, req.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, services.GenericRes{
		Msg: "Password changed successfully",
	})

}
