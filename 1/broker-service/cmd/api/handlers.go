package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/steve-mir/bukka_backend/broker/cmd/strategies"
)

// RequestPayload describes the JSON that this service accepts as an HTTP Post request
type RequestPayload struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

// Broker is a simple endpoint handler
func (app *Config) Broker() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Hit the broker"})
	}
}

// HandleSubmission processes incoming requests
func (app *Config) HandleSubmission() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestPayload RequestPayload
		log.Println("Payload action is", requestPayload.Action)

		if err := ctx.ShouldBindJSON(&requestPayload); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		strategy, err := app.StrategyFactory.GetStrategy(requestPayload.Action)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		app.handleRequest(ctx, strategy, requestPayload.Payload)
	}
}

func (app *Config) handleRequest(ctx *gin.Context, strategy strategies.ServiceStrategy, payload json.RawMessage) {
	correlationID := uuid.New().String()

	responseQueue, queueName, err := createResponseQueue(app.Rabbit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer responseQueue.Close()

	if err := strategy.Process(payload, correlationID, queueName); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	select {
	case response := <-waitForResponse(responseQueue, correlationID, queueName):
		ctx.JSON(http.StatusOK, response)
	case <-time.After(5 * time.Second):
		ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "request timed out"})
	}
}

func createResponseQueue(rabbitMQ *amqp.Connection) (*amqp.Channel, string, error) {
	ch, err := rabbitMQ.Channel()
	if err != nil {
		return nil, "", fmt.Errorf("failed to open channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		return nil, "", fmt.Errorf("failed to declare queue: %w", err)
	}

	return ch, q.Name, nil
}

func waitForResponse(ch *amqp.Channel, correlationID, queueName string) <-chan interface{} {
	responses := make(chan interface{})
	go func() {
		defer close(responses)

		msgs, err := ch.Consume(
			queueName, // queue
			"",        // consumer
			true,      // auto-ack
			false,     // exclusive
			false,     // no-local
			false,     // no-wait
			nil,       // args
		)
		if err != nil {
			return
		}

		for msg := range msgs {
			if msg.CorrelationId == correlationID {
				var response interface{}
				if err := json.Unmarshal(msg.Body, &response); err == nil {
					responses <- response
				}
				return
			}
		}
	}()
	return responses
}

/*
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ServiceStrategy interface
type ServiceStrategy interface {
	Process(payload []byte, correlationID, replyTo string) error
	QueueName() string
}

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

type MenuStrategy struct {
	rabbitMQ *amqp.Connection
}

func (s *MenuStrategy) Process(payload []byte, correlationID, replyTo string) error {
	return pushToQueue(s.rabbitMQ, s.QueueName(), payload, correlationID, replyTo)
}

func (s *MenuStrategy) QueueName() string {
	return "menu_queue"
}

// StrategyFactory for creating strategies
type StrategyFactory struct {
	strategies map[string]ServiceStrategy
}

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

// RequestPayload describes the JSON that this service accepts as an HTTP Post request
type RequestPayload struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

// Broker is a simple endpoint handler
func (app *Config) Broker() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Hit the broker"})
	}
}

// HandleSubmission processes incoming requests
func (app *Config) HandleSubmission() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestPayload RequestPayload
		log.Println("Payload action is", requestPayload.Action)

		if err := ctx.ShouldBindJSON(&requestPayload); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		strategy, err := app.StrategyFactory.GetStrategy(requestPayload.Action)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		app.handleRequest(ctx, strategy, requestPayload.Payload)
	}
}

func (app *Config) handleRequest(ctx *gin.Context, strategy ServiceStrategy, payload json.RawMessage) {
	correlationID := uuid.New().String()

	responseQueue, queueName, err := createResponseQueue(app.Rabbit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer responseQueue.Close()

	if err := strategy.Process(payload, correlationID, queueName); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	select {
	case response := <-waitForResponse(responseQueue, correlationID, queueName):
		ctx.JSON(http.StatusOK, response)
	case <-time.After(5 * time.Second):
		ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "request timed out"})
	}
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

func createResponseQueue(rabbitMQ *amqp.Connection) (*amqp.Channel, string, error) {
	ch, err := rabbitMQ.Channel()
	if err != nil {
		return nil, "", fmt.Errorf("failed to open channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		return nil, "", fmt.Errorf("failed to declare queue: %w", err)
	}

	return ch, q.Name, nil
}

func waitForResponse(ch *amqp.Channel, correlationID, queueName string) <-chan interface{} {
	responses := make(chan interface{})
	go func() {
		defer close(responses)

		msgs, err := ch.Consume(
			queueName, // queue
			"",        // consumer
			true,      // auto-ack
			false,     // exclusive
			false,     // no-local
			false,     // no-wait
			nil,       // args
		)
		if err != nil {
			return
		}

		for msg := range msgs {
			if msg.CorrelationId == correlationID {
				var response interface{}
				if err := json.Unmarshal(msg.Body, &response); err == nil {
					responses <- response
				}
				return
			}
		}
	}()
	return responses
}

*/

// ************************
// package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// 	amqp "github.com/rabbitmq/amqp091-go"
// )

// // Action represents different service actions
// type Action string

// // ServicePayload is an interface for all service payloads
// type ServicePayload interface {
// 	QueueName() string
// }

// // RequestPayload describes the JSON that this service accepts as an HTTP Post request
// type RequestPayload struct {
// 	Action  Action          `json:"action"`
// 	Payload json.RawMessage `json:"payload"`
// }

// // AuthPayload is the payload for authentication requests
// type AuthPayload struct {
// 	Action   Action `json:"auth_action"`
// 	Email    string `json:"email"`
// 	Password string `json:"password"`
// }

// func (AuthPayload) QueueName() string {
// 	return "auth_queue"
// }

// // MenuPayload is the payload for menu requests
// type MenuPayload struct {
// 	Action Action `json:"menu_action"`
// 	ID     string `json:"id,omitempty"`
// }

// func (MenuPayload) QueueName() string {
// 	return "menu_queue"
// }

// // Broker is a simple endpoint handler
// func (app *Config) Broker() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		ctx.JSON(http.StatusOK, gin.H{"message": "Hit the broker"})
// 	}
// }

// // HandleSubmission processes incoming requests
// func (app *Config) HandleSubmission() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		var requestPayload RequestPayload

// 		if err := ctx.ShouldBindJSON(&requestPayload); err != nil {
// 			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		var payload ServicePayload

// 		switch requestPayload.Action {
// 		case "auth":
// 			payload = &AuthPayload{}
// 		case "menu":
// 			payload = &MenuPayload{}
// 		default:
// 			ctx.JSON(http.StatusBadRequest, gin.H{"error": "unknown action"})
// 			return
// 		}

// 		if err := json.Unmarshal(requestPayload.Payload, payload); err != nil {
// 			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		app.handleRequest(ctx, payload)
// 	}
// }

// func (app *Config) handleRequest(ctx *gin.Context, payload ServicePayload) {
// 	correlationID := uuid.New().String()

// 	responseQueue, queueName, err := app.createResponseQueue()
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer responseQueue.Close()

// 	if err := app.pushToQueue(payload, correlationID, queueName); err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	select {
// 	case response := <-app.waitForResponse(responseQueue, correlationID, queueName):
// 		ctx.JSON(http.StatusOK, response)
// 	case <-time.After(5 * time.Second):
// 		ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "request timed out"})
// 	}
// }

// func (app *Config) pushToQueue(payload ServicePayload, correlationID, replyTo string) error {
// 	ch, err := app.Rabbit.Channel()
// 	if err != nil {
// 		return fmt.Errorf("failed to open channel: %w", err)
// 	}
// 	defer ch.Close()

// 	q, err := ch.QueueDeclare(
// 		payload.QueueName(), // name
// 		true,                // durable
// 		false,               // delete when unused
// 		false,               // exclusive
// 		false,               // no-wait
// 		nil,                 // arguments
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to declare queue: %w", err)
// 	}

// 	body, err := json.Marshal(payload)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal payload: %w", err)
// 	}

// 	err = ch.Publish(
// 		"",     // exchange
// 		q.Name, // routing key
// 		false,  // mandatory
// 		false,  // immediate
// 		amqp.Publishing{
// 			DeliveryMode:  amqp.Persistent,
// 			ContentType:   "application/json",
// 			CorrelationId: correlationID,
// 			ReplyTo:       replyTo,
// 			Body:          body,
// 		})
// 	if err != nil {
// 		return fmt.Errorf("failed to publish message: %w", err)
// 	}

// 	return nil
// }

// func (app *Config) createResponseQueue() (*amqp.Channel, string, error) {
// 	ch, err := app.Rabbit.Channel()
// 	if err != nil {
// 		return nil, "", fmt.Errorf("failed to open channel: %w", err)
// 	}

// 	q, err := ch.QueueDeclare(
// 		"",    // name
// 		false, // durable
// 		false, // delete when unused
// 		true,  // exclusive
// 		false, // no-wait
// 		nil,   // arguments
// 	)
// 	if err != nil {
// 		ch.Close()
// 		return nil, "", fmt.Errorf("failed to declare queue: %w", err)
// 	}

// 	return ch, q.Name, nil
// }

// func (app *Config) waitForResponse(ch *amqp.Channel, correlationID, queueName string) <-chan interface{} {
// 	responses := make(chan interface{})
// 	go func() {
// 		defer close(responses)

// 		msgs, err := ch.Consume(
// 			queueName, // queue
// 			"",        // consumer
// 			true,      // auto-ack
// 			false,     // exclusive
// 			false,     // no-local
// 			false,     // no-wait
// 			nil,       // args
// 		)
// 		if err != nil {
// 			return
// 		}

// 		for msg := range msgs {
// 			if msg.CorrelationId == correlationID {
// 				var response interface{}
// 				if err := json.Unmarshal(msg.Body, &response); err == nil {
// 					responses <- response
// 				}
// 				return
// 			}
// 		}
// 	}()
// 	return responses
// }

// ***
/*
package main

import (
	"bytes"
	"encoding/json"
	"errors"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// AuthAction represents different authentication actions
type AuthAction string
type MenuAction string

const (
	AuthActionLogin    AuthAction = "login"
	AuthActionRegister AuthAction = "register"
	AuthActionForgot   AuthAction = "forgot"
	MenuActionHome     MenuAction = "home"
	MenuActionDetails  MenuAction = "details"
)

// RequestPayload describes the JSON that this service accepts as an HTTP Post request
type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Menu   MenuPayload `json:"menu,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

// MailPayload is the embedded type (in RequestPayload) that describes an email message to be sent
type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// AuthPayload is the embedded type (in RequestPayload) that describes an authentication request
type AuthPayload struct {
	Action   AuthAction `json:"auth_action"`
	Email    string     `json:"email"`
	Password string     `json:"password"`
}

type MenuPayload struct {
	Action MenuAction `json:"menu_action"`
	ID     string     `json:"id,omitempty"`
}

// LogPayload is the embedded type (in RequestPayload) that describes a request to log something
type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) Broker() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		payload := jsonResponse{
			Error:   false,
			Message: "Hit the broker",
		}

		ctx.JSON(http.StatusOK, payload)
	}
}

func (app *Config) HandleSubmission() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var requestPayload RequestPayload

		if err := ctx.ShouldBindJSON(&requestPayload); err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}

		switch requestPayload.Action {
		case "auth":
			app.handleAuth(ctx, requestPayload.Auth)
		case "menu":
			app.handleMenu(ctx, requestPayload.Menu)
		// case "log":
		// 	app.logEventViaRabbit(ctx, requestPayload.Log)
		// case "mail":
		// 	app.sendMail(w, requestPayload.Mail)
		default:
			ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("unknown action")))
		}

	}
}

func (app *Config) handleAuth(ctx *gin.Context, auth AuthPayload) {
	// Generate a unique correlation ID
	correlationID := uuid.New().String()

	// Create a response queue
	responseQueue, err := app.createResponseQueue()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	defer responseQueue.Close()

	// Get the queue name
	queue, err := responseQueue.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Push the authentication request to RabbitMQ
	err = app.pushAuthToQueue(auth, correlationID, queue.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Wait for the response
	select {
	case response := <-app.waitForResponse(responseQueue, correlationID, queue.Name):
		ctx.JSON(http.StatusOK, response)
	case <-time.After(5 * time.Second):
		ctx.JSON(http.StatusRequestTimeout, errorResponse(errors.New("authentication request timed out")))
	}
}

func (app *Config) handleMenu(ctx *gin.Context, auth MenuPayload) {
	// Generate a unique correlation ID
	correlationID := uuid.New().String()

	// Create a response queue
	responseQueue, err := app.createResponseQueue()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	defer responseQueue.Close()

	// Get the queue name
	queue, err := responseQueue.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Push the authentication request to RabbitMQ
	err = app.pushMenuToQueue(auth, correlationID, queue.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Wait for the response
	select {
	case response := <-app.waitForResponse(responseQueue, correlationID, queue.Name):
		ctx.JSON(http.StatusOK, response)
	case <-time.After(5 * time.Second):
		ctx.JSON(http.StatusRequestTimeout, errorResponse(errors.New("menu request timed out")))
	}
}

func (app *Config) pushMenuToQueue(auth MenuPayload, correlationID, replyTo string) error {
	ch, err := app.Rabbit.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"menu_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(auth)
	if err != nil {
		return err
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
			Body:          body,
		})
	if err != nil {
		return err
	}

	return nil
}

func (app *Config) pushAuthToQueue(auth AuthPayload, correlationID, replyTo string) error {
	ch, err := app.Rabbit.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
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

	body, err := json.Marshal(auth)
	if err != nil {
		return err
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
			Body:          body,
		})
	if err != nil {
		return err
	}

	return nil
}

func (app *Config) createResponseQueue() (*amqp.Channel, error) {
	return app.Rabbit.Channel()
}

func (app *Config) waitForResponse(ch *amqp.Channel, correlationID, queueName string) <-chan interface{} {
	responses := make(chan interface{})
	go func() {
		msgs, err := ch.Consume(
			queueName, // queue
			"",        // consumer
			true,      // auto-ack
			false,     // exclusive
			false,     // no-local
			false,     // no-wait
			nil,       // args
		)
		if err != nil {
			close(responses)
			return
		}

		for msg := range msgs {
			if msg.CorrelationId == correlationID {
				var response interface{}
				err := json.Unmarshal(msg.Body, &response)
				if err == nil {
					responses <- response
				}
				return
			}
		}
	}()
	return responses
}
*/

/*
// ! sendMail sends email by calling the mail microservice
func (app *Config) sendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData, _ := json.MarshalIndent(msg, "", "\t")

	// call the mail service
	mailServiceURL := "http://mailer-service/send"

	// post to mail service
	request, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the right status code
	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling mail service"))
		return
	}

	// send back json
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Message sent to " + msg.To

	app.writeJSON(w, http.StatusAccepted, payload)

}

// logEventViaRabbit logs an event using the logger-service. It makes the call by pushing the data to RabbitMQ.
func (app *Config) logEventViaRabbit(ctx *gin.Context, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data, "log.INFO")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged via RabbitMQ"

	ctx.JSON(http.StatusOK, payload)
}
*/
