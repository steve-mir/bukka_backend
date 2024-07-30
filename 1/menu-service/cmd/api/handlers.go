package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MenuAction string

const (
	MenuActionHome    MenuAction = "home"
	MenuActionDetails MenuAction = "details"
)

type MenuPayload struct {
	Action MenuAction `json:"menu_action"`
	ID     string     `json:"id,omitempty"`
}

type MenuResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

func (app *Config) Menu(w http.ResponseWriter, r *http.Request) {
	var payload MenuPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	var response MenuResponse
	log.Println("Action", payload.Action)
	switch payload.Action {
	case MenuActionHome:
		response = app.handleHome(payload)
	case MenuActionDetails:
		response = app.handleDetails(payload)
	default:
		response = MenuResponse{Error: true, Message: "Invalid menu action"}
	}

	app.writeJSON(w, http.StatusAccepted, response)

	// If this was called via RabbitMQ, send the response back
	if r.Header.Get("X-RabbitMQ-Reply-To") != "" {
		app.sendRabbitMQResponse(response, r.Header.Get("X-RabbitMQ-Reply-To"), r.Header.Get("X-Correlation-ID"))
	}
}

func (app *Config) handleHome(payload MenuPayload) MenuResponse {
	// Implement login logic here
	return MenuResponse{
		Error:   false,
		Message: fmt.Sprintf("Welcome to menu lists %s", payload.ID),
	}
}

func (app *Config) handleDetails(payload MenuPayload) MenuResponse {
	// Implement registration logic here
	return MenuResponse{
		Error:   false,
		Message: fmt.Sprintf("Calling menu details %s", payload.ID),
	}
}

func (app *Config) sendRabbitMQResponse(response interface{}, replyTo, correlationID string) {
	ch, err := app.Rabbit.Channel()
	if err != nil {
		log.Printf("Failed to open a channel: %s", err)
		return
	}
	defer ch.Close()

	body, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response: %s", err)
		return
	}

	err = ch.Publish(
		"amq.direct", // exchange
		replyTo,      // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			Body:          body,
		})
	if err != nil {
		log.Printf("Failed to publish a message: %s", err)
		return
	}
}
