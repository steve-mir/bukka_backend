package event

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/steve-mir/bukka_backend/listener/factory"
)

// Consumer represents a RabbitMQ consumer
type Consumer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	handlers map[string]factory.CommandFactory
}

// NewConsumer creates a new Consumer
func NewConsumer(conn *amqp.Connection) (*Consumer, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &Consumer{
		conn:    conn,
		channel: channel,
		handlers: map[string]factory.CommandFactory{
			"auth_queue": factory.AuthCommandFactory{},
			"menu_queue": factory.MenuCommandFactory{},
		},
	}, nil
}

// Listen starts consuming messages from queues
func (c *Consumer) Listen() error {
	for queueName, factory := range c.handlers {
		if err := c.setupQueue(queueName); err != nil {
			return fmt.Errorf("failed to setup queue %s: %w", queueName, err)
		}

		messages, err := c.consumeQueue(queueName)
		if err != nil {
			return fmt.Errorf("failed to consume queue %s: %w", queueName, err)
		}

		go c.handleMessages(messages, factory)
	}

	log.Println("Waiting for messages on queues")
	select {}
}

func (c *Consumer) setupQueue(queueName string) error {
	_, err := c.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return err
}

func (c *Consumer) consumeQueue(queueName string) (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		queueName,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
}

func (c *Consumer) handleMessages(messages <-chan amqp.Delivery, factory factory.CommandFactory) {
	for d := range messages {
		go c.processMessage(d, factory)
	}
}

func (c *Consumer) processMessage(d amqp.Delivery, factory factory.CommandFactory) {
	command, err := factory.CreateCommand(d.Body)
	if err != nil {
		log.Printf("Error creating command: %v", err)
		c.channel.Nack(d.DeliveryTag, false, false)
		return
	}

	response, err := command.Execute()
	if err != nil {
		log.Printf("Error executing command: %v", err)
		c.channel.Nack(d.DeliveryTag, false, false)
		return
	}

	if err := sendResponse(response, d.ReplyTo, d.CorrelationId, c.channel); err != nil {
		log.Printf("Error sending response: %v", err)
		c.channel.Nack(d.DeliveryTag, false, false)
		return
	}

	c.channel.Ack(d.DeliveryTag, false)
}

func sendResponse(response []byte, replyTo, correlationId string, ch *amqp.Channel) error {
	return ch.Publish(
		"",      // exchange
		replyTo, // routing key
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationId,
			Body:          response,
		},
	)
}
