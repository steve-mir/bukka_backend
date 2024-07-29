package main

import (
	"broker/event"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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
	Email    string `json:"email"`
	Password string `json:"password"`
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
			app.authenticateViaRabbit(ctx, requestPayload.Auth)
		case "log":
			app.logEventViaRabbit(ctx, requestPayload.Log)
		// case "mail":
		// 	app.sendMail(w, requestPayload.Mail)
		default:
			ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("unknown action")))
		}

	}
}

func (app *Config) authenticateViaRabbit(ctx *gin.Context, a AuthPayload) {
	err := app.pushToQueue(a.Email, a.Password, "auth.REQUEST")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authentication request sent via RabbitMQ"

	ctx.JSON(http.StatusAccepted, payload)
}

// authenticate calls the authentication microservice and sends back the appropriate response
func (app *Config) authenticate(ctx *gin.Context, a AuthPayload) {
	// create some json we'll send to the auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// call the service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code
	if response.StatusCode == http.StatusUnauthorized {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("invalid credentials")))

		return
	} else if response.StatusCode != http.StatusAccepted {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("error calling auth service")))
		return
	}

	// create a variable we'll read response.Body into
	var jsonFromService jsonResponse

	// decode the json from the auth service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if jsonFromService.Error {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	ctx.JSON(http.StatusOK, payload)
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

// pushToQueue pushes a message into RabbitMQ
func (app *Config) pushToQueue2(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	j, err := json.MarshalIndent(&payload, "", "\t")
	if err != nil {
		return err
	}

	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}
	return nil
}

func (app *Config) pushToQueue(name, msg, routingKey string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	// payload := LogPayload{
	// 	Name: name,
	// 	Data: msg,
	// }
	payload := make(map[string]interface{})

	switch routingKey {
	case "auth.REQUEST":
		payload["email"] = name
		payload["password"] = msg
	case "log.INFO":
		payload["name"] = name
		payload["data"] = msg
	default:
		return fmt.Errorf("unknown routing key: %s", routingKey)
	}

	j, err := json.MarshalIndent(&payload, "", "\t")
	if err != nil {
		return err
	}

	err = emitter.Push(string(j), routingKey)
	if err != nil {
		return err
	}
	return nil
}

/*
func (app *Config) LogViaGRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer conn.Close()

	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.writeJSON(w, http.StatusAccepted, payload)
}
*/
