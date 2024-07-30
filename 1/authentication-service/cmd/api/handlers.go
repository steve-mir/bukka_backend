package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AuthAction string

const (
	AuthActionLogin    AuthAction = "login"
	AuthActionRegister AuthAction = "register"
	AuthActionForgot   AuthAction = "forgot"
)

type AuthPayload struct {
	Action   AuthAction `json:"auth_action"`
	Email    string     `json:"email"`
	Password string     `json:"password"`
	Username string     `json:"username,omitempty"`
}

type AuthResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	// Add other fields as needed
}

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var payload AuthPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	var response AuthResponse
	log.Println("Action", payload.Action)
	switch payload.Action {
	case AuthActionLogin:
		response = app.handleLogin(payload)
	case AuthActionRegister:
		response = app.handleRegister(payload)
	case AuthActionForgot:
		response = app.handleForgotPassword(payload)
	default:
		response = AuthResponse{Error: true, Message: "Invalid auth action"}
	}

	app.writeJSON(w, http.StatusAccepted, response)

	// If this was called via RabbitMQ, send the response back
	if r.Header.Get("X-RabbitMQ-Reply-To") != "" {
		app.sendRabbitMQResponse(response, r.Header.Get("X-RabbitMQ-Reply-To"), r.Header.Get("X-Correlation-ID"))
	}
}

func (app *Config) handleLogin(payload AuthPayload) AuthResponse {
	// Implement login logic here
	return AuthResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", payload.Email),
	}
}

func (app *Config) handleRegister(payload AuthPayload) AuthResponse {
	// Implement registration logic here
	return AuthResponse{
		Error:   false,
		Message: fmt.Sprintf("Registered user %s", payload.Email),
	}
}

func (app *Config) handleForgotPassword(payload AuthPayload) AuthResponse {
	// Implement forgot password logic here
	return AuthResponse{
		Error:   false,
		Message: fmt.Sprintf("Password reset initiated for user %s", payload.Email),
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

func (app *Config) AuthenticateOld(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// err := app.readJSON(w, r, &requestPayload)
	// if err != nil {
	// 	app.errorJSON(w, err, http.StatusBadRequest)
	// 	return
	// }

	// // validate the user against the database
	// user, err := app.Models.User.GetByEmail(requestPayload.Email)
	// if err != nil {
	// 	app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
	// 	return
	// }

	// valid, err := user.PasswordMatches(requestPayload.Password)
	// if err != nil || !valid {
	// 	app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
	// 	return
	// }

	// // log authentication
	// err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	// if err != nil {
	// 	app.errorJSON(w, err)
	// 	return
	// }

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", requestPayload.Email),
		// Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}
