package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/steve-mir/bukka_backend/internal/app/auth/middlewares"
	"github.com/steve-mir/bukka_backend/internal/app/auth/services"
	"github.com/steve-mir/bukka_backend/token"
)

func (s *Server) viewProfile(ctx *gin.Context) {
	authPayload := ctx.MustGet(middlewares.AuthorizationPayloadKey).(*token.Payload)

	user, err := s.store.GetUserProfile(ctx, sql.NullString{String: authPayload.Subject.String(), Valid: true})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, services.UserProfileRes{
		Uid:        user.ID.String(),
		Username:   user.Username.String,
		Email:      user.Email,
		Phone:      user.Phone.String,
		FirstName:  user.FirstName.String,
		LastName:   user.LastName.String,
		ImageUrl:   user.ImageUrl.String,
		CreatedAt:  user.CreatedAt.Time,
		IsVerified: user.IsVerified.Bool,
	})

}
