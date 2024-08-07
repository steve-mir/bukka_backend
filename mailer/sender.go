// sender.go
package mailer

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

// EmailSender interface to send emails
type EmailSender interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFiles []string,
	) error
}

// SMTPSender struct for SMTP email sending
type SMTPSender struct {
	name     string
	address  string
	host     string
	port     string
	username string
	password string
}

// NewSMTPSender creates a new SMTPSender instance
func NewSMTPSender(name string, address string, host string, port string, username string, password string) *SMTPSender {
	return &SMTPSender{
		name:     name,
		address:  address,
		host:     host,
		port:     port,
		username: username,
		password: password,
	}
}

// SendEmail sends an email via SMTP
func (sender *SMTPSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.address)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	for _, f := range attachFiles {
		_, err := e.AttachFile(f)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", f, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", sender.username, sender.password, sender.host)
	return e.Send(fmt.Sprintf("%s:%s", sender.host, sender.port), smtpAuth)
}

/*package mailer

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

// type EmailSender interface {
// 	SendEmail(
// 		subject string,
// 		content string,
// 		to []string,
// 		cc []string,
// 		bcc []string,
// 		attachFiles []string,
// 	) error
// }

// type SMTPSender struct {
// 	name     string
// 	address  string
// 	host     string
// 	username string
// 	password string
// }

// func NewSMTPSender(name string, address string, host string, username string, password string) *SMTPSender {
// 	return &SMTPSender{
// 		name:     name,
// 		address:  address,
// 		host:     host,
// 		username: username,
// 		password: password,
// 	}
// }

// func (sender *SMTPSender) SendEmail(
// 	subject string,
// 	content string,
// 	to []string,
// 	cc []string,
// 	bcc []string,
// 	attachFiles []string,
// ) error {
// 	e := email.NewEmail()
// 	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.username)
// 	e.Subject = subject
// 	e.HTML = []byte(content)
// 	e.To = to
// 	e.Cc = cc
// 	e.Bcc = bcc

// 	for _, f := range attachFiles {
// 		_, err := e.AttachFile(f)
// 		if err != nil {
// 			return fmt.Errorf("failed to attache file %s: %w", f, err)
// 		}
// 	}

// 	smtpAuth := smtp.PlainAuth("", sender.username, sender.password, sender.host)
// 	return e.Send(sender.address, smtpAuth)
// }

type EmailSender interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFiles []string,
	) error
}

type SMTPSender struct {
	name     string
	address  string
	host     string
	port     string
	username string
	password string
}

func NewSMTPSender(name string, address string, host string, port string, username string, password string) *SMTPSender {
	return &SMTPSender{
		name:     name,
		address:  address,
		host:     host,
		port:     port,
		username: username,
		password: password,
	}
}

func (sender *SMTPSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.address)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	for _, f := range attachFiles {
		_, err := e.AttachFile(f)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", f, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", sender.username, sender.password, sender.host)
	return e.Send(fmt.Sprintf("%s:%s", sender.host, sender.port), smtpAuth)
}
*/
