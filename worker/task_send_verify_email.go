package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"github.com/steve-mir/bukka_backend/mailer"
)

const TaskSendEmail = "task:send_email"

type PayloadSendEmail struct {
	Username string `json:"username"`
	Content  string
}

func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendEmail,
	opts ...asynq.Option,
) error {

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TaskSendEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	user, err := processor.store.GetUserByIdentifier(ctx, payload.Username)
	if err != nil {
		// if err == sql.ErrNoRows {
		// 	return fmt.Errorf("user not found: %w", err)
		// }
		return fmt.Errorf("failed to get user: %w", err)
	}

	// TODO: send email to use here
	log.Print("sending email", user.Email, " content", payload.Content)
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Str("email", user.Email).Msg("processed task")

	sender := mailer.NewSMTPSender("Bukka", processor.config.SMTPAddr, processor.config.SMTPHost, "2525", processor.config.SMTPUsername, processor.config.SMTPPassword)
	return sender.SendEmail("Email verification", payload.Content, []string{user.Email}, []string{}, []string{}, []string{})
	// return nil
}
