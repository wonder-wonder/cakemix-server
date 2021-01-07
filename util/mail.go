package util

import (
	"errors"
	"fmt"
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
	} else if SendGridAPIKey == "DEBUG" {
		fmt.Printf("SendMail debug mode.\n"+
			"To: %s <%s>\n"+
			"Suject: %s\n"+
			"Contents:\n%s\n", ToAddr, ToName, subject, text)
		return nil
	}
	from := mail.NewEmail(MailFromName, MailFromAddr)
	to := mail.NewEmail(ToName, ToAddr)
	message := mail.NewSingleEmail(from, subject, to, text, textHTML)
	client := sendgrid.NewSendClient(SendGridAPIKey)
	res, err := client.Send(message)
	if res.StatusCode >= 400 {
		return errors.New(res.Body)
	}
	return err
}
