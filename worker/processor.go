package worker

import (
	"context"
	"database/sql"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"github.com/steve-mir/bukka_backend/db/sqlc"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  sqlc.Store
	db     *sql.DB
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store sqlc.Store, db *sql.DB) TaskProcessor {

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().Err(err).Str("type", task.Type()).
					Bytes("payload", task.Payload()).Msg("process task failed")
			}),
			Logger: NewLogger(),
		})

	return &RedisTaskProcessor{
		server: server,
		store:  store,
		db:     db,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	// Register tasks here. Very important
	mux.HandleFunc(TaskSendEmail, processor.ProcessTaskSendVerifyEmail)

	return processor.server.Start(mux)
}
