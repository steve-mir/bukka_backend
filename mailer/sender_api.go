// sender_api.go
package mailer

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/steve-mir/bukka_backend/utils"
)

type APIEmailSender struct {
	apiClient *utils.HTTPRequest
	from      EmailAddress
}

type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type Attachment struct {
	Content     string `json:"content"`
	Filename    string `json:"filename"`
	Type        string `json:"type"`
	Disposition string `json:"disposition"`
	ContentID   string `json:"content_id"`
}

type MailPayload struct {
	From        EmailAddress   `json:"from"`
	To          []EmailAddress `json:"to"`
	Cc          []EmailAddress `json:"cc,omitempty"`
	Bcc         []EmailAddress `json:"bcc,omitempty"`
	Subject     string         `json:"subject"`
	Text        string         `json:"text"`
	Category    string         `json:"category"`
	Attachments []Attachment   `json:"attachments,omitempty"`
}

// NewAPIEmailSender creates a new APIEmailSender instance
func NewAPIEmailSender(baseURL, apiKey, fromEmail, fromName string) *APIEmailSender {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", apiKey),
		"Content-Type":  "application/json",
	}
	apiClient := utils.NewHTTPRequest(baseURL, headers)
	from := EmailAddress{
		Email: fromEmail,
		Name:  fromName,
	}
	return &APIEmailSender{
		apiClient: apiClient,
		from:      from,
	}
}

// SendEmail sends an email via API
func (sender *APIEmailSender) SendEmail(subject, text string, to []string, cc []string, bcc []string, attachFiles []string) error {
	toAddresses := []EmailAddress{}
	for _, email := range to {
		toAddresses = append(toAddresses, EmailAddress{Email: email})
	}

	ccAddresses := []EmailAddress{}
	for _, email := range cc {
		ccAddresses = append(ccAddresses, EmailAddress{Email: email})
	}

	bccAddresses := []EmailAddress{}
	for _, email := range bcc {
		bccAddresses = append(bccAddresses, EmailAddress{Email: email})
	}

	attachments := []Attachment{}
	for _, filePath := range attachFiles {
		fileData, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error reading file: %v", err)
		}
		encodedFile := base64.StdEncoding.EncodeToString(fileData)
		attachment := Attachment{
			Content:     encodedFile,
			Filename:    filePath,
			Type:        "application/octet-stream",
			Disposition: "attachment",
			ContentID:   "file1",
		}
		attachments = append(attachments, attachment)
	}

	payload := MailPayload{
		From:        sender.from,
		To:          toAddresses,
		Cc:          ccAddresses,
		Bcc:         bccAddresses,
		Subject:     subject,
		Text:        text,
		Category:    "Integration Test",
		Attachments: attachments,
	}

	_, statusCode, err := sender.apiClient.Post("send", payload)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	if statusCode >= 300 {
		return fmt.Errorf("failed to send email, status code: %d", statusCode)
	}

	return nil
}
