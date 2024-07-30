package strategies

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Strategy pattern along with the Factory pattern to make the broker even more flexible and extensible while maintaining optimal performance.
// This approach will allow to easily add new services without modifying existing code, adhering to the Open/Closed Principle.

// ServiceStrategy interface
type ServiceStrategy interface {
	Process(payload []byte, correlationID, replyTo string) error
	QueueName() string
}

// StrategyFactory for creating strategies
type StrategyFactory struct {
	strategies map[string]ServiceStrategy
}

// ! Add services here
func NewStrategyFactory(rabbitMQ *amqp.Connection) *StrategyFactory {
	return &StrategyFactory{
		strategies: map[string]ServiceStrategy{
			"auth": &AuthStrategy{rabbitMQ: rabbitMQ},
			"menu": &MenuStrategy{rabbitMQ: rabbitMQ},
		},
	}
}

func (f *StrategyFactory) GetStrategy(action string) (ServiceStrategy, error) {
	strategy, ok := f.strategies[action]
	if !ok {
		return nil, fmt.Errorf("unknown action: %s", action)
	}
	return strategy, nil
}

func pushToQueue(rabbitMQ *amqp.Connection, queueName string, payload []byte, correlationID, replyTo string) error {
	ch, err := rabbitMQ.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode:  amqp.Persistent,
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       replyTo,
			Body:          payload,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
