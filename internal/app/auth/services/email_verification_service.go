package services

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/steve-mir/bukka_backend/db/sqlc"
	"github.com/steve-mir/bukka_backend/utils"
	"github.com/steve-mir/bukka_backend/worker"
)

const (
	codeLength    = int64(1000000)
	length        = 6
	ResetMsg      = "if an account exists a password reset email will be sent to you"
	UnexpectedErr = "an unexpected error occurred"
)

func SendVerificationEmail(qtx *sqlc.Queries, ctx context.Context, td worker.TaskDistributor, userId uuid.UUID, email string) error {
	log.Info().Msg("Sending email to new user...")
	verificationCode, err := utils.GenerateSecureRandomNumber(codeLength)
	if err != nil {
		return errors.New("failed to generate secure random number")
	}
	code := fmt.Sprintf("%06d", verificationCode)
	content := fmt.Sprintf("Use this to verify you email. code %s", code)

	// Use a WaitGroup to wait for both goroutines to complete
	var wg sync.WaitGroup
	wg.Add(2) // We have two goroutines

	// Error channel with buffer for two errors
	errChan := make(chan error, 2)

	go func() {
		defer wg.Done() // Notify the WaitGroup that this goroutine is done
		// Send email here.
		if err := SendEmail(td, ctx, email, content); err != nil {
			errChan <- errors.New("failed to send verification email")
		}
	}()
	return nil

}
