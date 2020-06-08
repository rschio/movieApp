package mail

import (
	"strconv"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Mailer struct {
	client *sendgrid.Client
	from   *mail.Email
}

// SendVerificationLink send a link to user email to verify account.
func (m *Mailer) SendVerificationLink(toUser, toAddr, link string) error {
	subject := "App list of movies verification."
	text := "Click here to verify your email: " + link
	return m.send(toUser, toAddr, subject, text)
}

// SendScheduledMovie send a email to user to remeber of a scheduled movie.
func (m *Mailer) SendScheduledMovie(toUser, toAddr string, movieID int) error {
	subject := "Watch your movie."
	text := "It's time, watch movie with ID: " + strconv.Itoa(movieID)
	return m.send(toUser, toAddr, subject, text)
}

func (m *Mailer) send(toUser, toAddr, subject, text string) error {
	to := mail.NewEmail(toUser, toAddr)
	message := mail.NewSingleEmail(m.from, subject, to, text, text)
	_, err := m.client.Send(message)
	return err
}

// NewMailer creates a new mailer with sender user and email and apiKey.
func NewMailer(fromUser, fromAddr, apiKey string) *Mailer {
	client := sendgrid.NewSendClient(apiKey)
	from := mail.NewEmail(fromUser, fromAddr)
	return &Mailer{
		client: client,
		from:   from,
	}
}
