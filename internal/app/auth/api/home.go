package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/steve-mir/bukka_backend/internal/app/auth/services"
)

func (s *Server) home(ctx *gin.Context) {
	ctx.JSON(http.StatusOK,
		services.HomeRes{
			Msg: "Welcome to Bukka homepage 😃 Auth service. Your account is safe with us",
		},
	)

}
