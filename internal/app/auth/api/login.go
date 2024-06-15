package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/steve-mir/bukka_backend/internal/app/auth/services"
)

func (s *Server) login(ctx *gin.Context) {
	var req services.LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	clientIP := ctx.ClientIP()
	agent := ctx.Request.UserAgent()

	userData, err := services.LogUserIn(req, s.store, ctx, s.config, clientIP, agent)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, userData)

}
