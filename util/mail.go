package util

import (
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const (
	MailFromAddr = "cakemix@wonder-wonder.xyz"
	MailFromName = "Cakemix"
)

var (
	SendGridAPIKey = os.Getenv("SENDGRID_API_KEY")
)

// SendMail sends email. The mail is sent as plain text if textHTML is empty.
func SendMail(ToAddr, ToName, subject, text, textHTML string) error {
	if SendGridAPIKey == "" {
		return nil
	}
	from := mail.NewEmail(MailFromName, MailFromAddr)
	to := mail.NewEmail(ToName, ToAddr)
	message := mail.NewSingleEmail(from, subject, to, text, textHTML)
	client := sendgrid.NewSendClient(SendGridAPIKey)
	_, err := client.Send(message)
	return err
}
