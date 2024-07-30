package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

// AuthPayload is the embedded type (in RequestPayload) that describes an authentication request
type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LogPayload is the embedded type (in RequestPayload) that describes a request to log something
type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type Consumer struct {
	conn      *amqp.Connection
	queueName string
	channel   *amqp.Channel
}

func NewConsumerOld(conn *amqp.Connection) (Consumer, error) {
	channel, err := conn.Channel()
	if err != nil {
		return Consumer{}, err
	}

	consumer := Consumer{
		conn:    conn,
		channel: channel,
	}

	err = consumer.setup()
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	channel, err := conn.Channel()
	if err != nil {
		return Consumer{}, err
	}

	consumer := Consumer{
		conn:    conn,
		channel: channel,
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	return declareExchange(consumer.channel)
}

type AuthAction string

const (
	AuthActionLogin    AuthAction = "login"
	AuthActionRegister AuthAction = "register"
	AuthActionForgot   AuthAction = "forgot"
)

type Payload struct {
	Action   AuthAction `json:"auth_action"`
	Action2  string     `json:"auth_action2"`
	Name     string     `json:"name,omitempty"`
	Data     string     `json:"data,omitempty"`
	Email    string     `json:"email,omitempty"`
	Password string     `json:"password,omitempty"`
}

// func (consumer *Consumer) ListenOld(topics []string) error {
// 	q, err := declareRandomQueue(consumer.channel)
// 	if err != nil {
// 		return err
// 	}

// 	for _, s := range topics {
// 		consumer.channel.QueueBind(
// 			q.Name,
// 			s,
// 			"logs_topic",
// 			false,
// 			nil,
// 		)

// 		if err != nil {
// 			return err
// 		}
// 	}

// 	messages, err := consumer.channel.Consume(q.Name, "", true, false, false, false, nil)
// 	if err != nil {
// 		return err
// 	}

// 	forever := make(chan bool)
// 	go func() {
// 		for d := range messages {
// 			var payload Payload
// 			_ = json.Unmarshal(d.Body, &payload)
// 			go handlePayload(payload, d.ReplyTo, d.CorrelationId, consumer.channel)
// 		}
// 	}()

// 	fmt.Printf("Waiting for message [Exchange, Queue] [logs_topic, %s]\n", q.Name)
// 	<-forever

// 	return nil
// }

// func handlePayloadOld(payload Payload, replyTo, correlationId string, ch *amqp.Channel) {
// 	switch {
// 	case payload.Name != "":
// 		// This is a log event
// 		err := logEvent(payload)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 	case payload.Email != "":
// 		// This is an auth event
// 		response, err := authenticate(payload)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		// Send the response back through RabbitMQ
// 		err = sendResponse(response, replyTo, correlationId, ch)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 	default:
// 		log.Println("Unknown payload type")
// 	}
// }

func (consumer *Consumer) Listen() error {
	q, err := consumer.channel.QueueDeclare(
		"auth_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return err
	}

	// err = consumer.channel.Qos(
	// 	1,     // prefetch count
	// 	0,     // prefetch size
	// 	false, // global
	// )
	err = consumer.channel.Qos(
		3,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return err
	}

	messages, err := consumer.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)
			go handlePayload(payload, d.ReplyTo, d.CorrelationId, consumer.channel, d.DeliveryTag)
		}
	}()

	fmt.Printf("Waiting for messages on queue %s\n", q.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload, replyTo, correlationId string, ch *amqp.Channel, deliveryTag uint64) {
	switch {
	case payload.Email != "":
		// This is an auth event
		response, err := authenticate(payload)
		if err != nil {
			log.Println(err)
		}
		// Send the response back through RabbitMQ
		err = sendResponse(response, replyTo, correlationId, ch)
		if err != nil {
			log.Println(err)
		}
	default:
		log.Println("Unknown payload type")
	}

	// Acknowledge the message
	ch.Ack(deliveryTag, false)
}

func logEvent(entry Payload) error {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service/log"
	log.Printf("Inside log with %v", jsonData)

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return err
	}

	return nil
}

func authenticate(entry Payload) ([]byte, error) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	authServiceURL := "http://authentication-service/authenticate"
	// log.Printf("Inside authenticate with %v", jsonData)

	request, err := http.NewRequest("POST", authServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("authentication failed with status code: %d", response.StatusCode)
	}

	log.Println("Auth res", response.StatusCode)

	// Read the response body
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func sendResponse(response []byte, replyTo, correlationId string, ch *amqp.Channel) error {
	err := ch.Publish(
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
	if err != nil {
		return err
	}
	return nil
}
