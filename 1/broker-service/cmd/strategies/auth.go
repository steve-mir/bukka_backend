package strategies

import amqp "github.com/rabbitmq/amqp091-go"

// Concrete strategies
type AuthStrategy struct {
	rabbitMQ *amqp.Connection
}

func (s *AuthStrategy) Process(payload []byte, correlationID, replyTo string) error {
	return pushToQueue(s.rabbitMQ, s.QueueName(), payload, correlationID, replyTo)
}

func (s *AuthStrategy) QueueName() string {
	return "auth_queue"
}
