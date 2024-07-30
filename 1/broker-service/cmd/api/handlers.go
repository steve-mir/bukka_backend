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

const (
	AuthActionLogin    AuthAction = "login"
	AuthActionRegister AuthAction = "register"
	AuthActionForgot   AuthAction = "forgot"
)

// RequestPayload describes the JSON that this service accepts as an HTTP Post request
type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
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
		// case "auth":
		// 	app.authenticateViaRabbit(ctx, requestPayload.Auth)
		case "auth":
			app.handleAuth(ctx, requestPayload.Auth)
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

// sendMail sends email by calling the mail microservice
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

/*
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
