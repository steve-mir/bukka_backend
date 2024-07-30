package strategies

import amqp "github.com/rabbitmq/amqp091-go"

type MenuStrategy struct {
	rabbitMQ *amqp.Connection
}

func (s *MenuStrategy) Process(payload []byte, correlationID, replyTo string) error {
	return pushToQueue(s.rabbitMQ, s.QueueName(), payload, correlationID, replyTo)
}

func (s *MenuStrategy) QueueName() string {
	return "menu_queue"
}
