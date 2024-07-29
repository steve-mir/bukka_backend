package event

import (
	"bytes"
	"encoding/json"
	"fmt"
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
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(channel)
}

type Payload struct {
	Name     string `json:"name,omitempty"`
	Data     string `json:"data,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	for _, s := range topics {
		ch.QueueBind(
			q.Name,
			s,
			"logs_topic",
			false,
			nil,
		)

		if err != nil {
			return err
		}
	}

	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)
		}
	}()

	fmt.Printf("Waiting for message [Exchange, Queue] [logs_topic, %s]\n", q.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch {
	case payload.Name != "":
		// This is a log event
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	case payload.Email != "":
		// This is an auth event
		err := authenticate(payload)
		if err != nil {
			log.Println(err)
		}
	default:
		log.Println("Unknown payload type")
	}
}

// func handlePayload(payload Payload) {
// 	log.Println("Received new log from service", payload.Name, payload.Data)
// 	switch payload.Name {
// 	case "log", "event":
// 		// log whatever we get
// 		err := logEvent(payload)
// 		if err != nil {
// 			log.Println(err)
// 		}

// 	case "auth":
// 		err := authenticate(payload)
// 		if err != nil {
// 			log.Println(err)
// 		}

// 	// you can have as many cases as you want, as long as you write the logic

// 	default:
// 		log.Println("Because nothing is found")
// 		err := logEvent(payload)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 	}
// }

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

func authenticate(entry Payload) error {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	authServiceURL := "http://authentication-service/authenticate"
	log.Printf("Inside authenticate with %v", jsonData)

	request, err := http.NewRequest("POST", authServiceURL, bytes.NewBuffer(jsonData))
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
		return fmt.Errorf("authentication failed with status code: %d", response.StatusCode)
	}

	log.Println("Auth res", response.StatusCode)

	return nil
}
