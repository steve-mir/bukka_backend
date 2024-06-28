package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/internal/app/auth/services"
)

func (s *Server) register(ctx *gin.Context) {
	var req services.RegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	clientIP := ctx.ClientIP()
	agent := ctx.Request.UserAgent()

	// Begin transaction
	tx, err := s.db.Begin()

	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	defer tx.Rollback()

	qtx := sqlc.New(tx)

	// Check db if user exists
	if err := services.CheckUserExists(ctx, qtx, req.Email, req.Username); err != nil {
		// return nil, status.Errorf(codes.AlreadyExists, err.Error())
		log.Err(err).Msg("Error2")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// hash pwd and generate uuid
	hashedPwd, uid, err := services.PrepareUserData(req.Password)
	if err != nil {
		// return nil, status.Errorf(codes.Internal, err.Error())
		log.Err(err).Msg("Error3")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	sqlcUser, err := services.CreateUserConcurrent(ctx, qtx /*tx,*/, uid, req.Email, req.Username, hashedPwd, false, false)
	if err != nil {
		// return nil, status.Errorf(codes.Internal, "error while creating user with email and password %s", err)
		log.Err(err).Msg("Error4")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Run concurrent operations
	accessToken, accessExp, err := services.RunConcurrentUserCreationTasks(ctx, s.tokenMaker, qtx, tx, s.config, s.taskDistributor, req, uid, clientIP, agent, false)
	if err != nil {
		// return nil, status.Errorf(codes.Internal, "error creating details: %s", err.Error())
		log.Err(err).Msg("Error5")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Only commit the transaction if all previous steps were successful
	if err := tx.Commit(); err != nil {
		// return nil, status.Errorf(codes.Internal, "an unexpected error occurred during transaction commit: %s", err)
		log.Err(err).Msg("Error6")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, services.UserAuthRes{
		Uid:             sqlcUser.ID,
		Username:        sqlcUser.Username.String,
		Email:           sqlcUser.Email,
		IsEmailVerified: sqlcUser.IsEmailVerified.Bool,
		CreatedAt:       sqlcUser.CreatedAt.Time,
		AuthToken: services.AuthToken{
			AccessToken:          accessToken,
			AccessTokenExpiresAt: accessExp,
		},
	})
}
