package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var (
	mailFromAddr = ""
	mailFromName = ""
	sgAPIKey     = ""
)

// InitMail setup mail configuration
func InitMail(sendgridAPIKey, fromAddr, fromName string) {
	sgAPIKey = sendgridAPIKey
	mailFromAddr = fromAddr
	mailFromName = fromName
}

// SendMail sends email. The mail is sent as plain text if textHTML is empty.
func SendMail(ToAddr, ToName, subject, text, textHTML string) error {
	if sgAPIKey == "" {
		return nil
	} else if sgAPIKey == "DEBUG" {
		fmt.Printf("SendMail debug mode.\n"+
			"To: %s <%s>\n"+
			"Suject: %s\n"+
			"Contents:\n%s\n", ToAddr, ToName, subject, text)
		return nil
	}
	from := mail.NewEmail(mailFromName, mailFromAddr)
	to := mail.NewEmail(ToName, ToAddr)
	message := mail.NewSingleEmail(from, subject, to, text, textHTML)
	client := sendgrid.NewSendClient(sgAPIKey)
	res, err := client.Send(message)
	if res.StatusCode >= 400 {
		return errors.New(res.Body)
	}
	return err
}

// SendMailWithTemplate sends email using template file. The mail is sent as plain text if textHTML is empty.
func SendMailWithTemplate(ToAddr, ToName, subject, tmplfile string, dat map[string]string) error {
	// #nosec G304
	raw, err := ioutil.ReadFile(tmplfile)
	if err != nil {
		return err
	}
	mailtext := string(raw)
	for k, v := range dat {
		mailtext = strings.ReplaceAll(mailtext, "{{"+k+"}}", v)
	}
	return SendMail(ToAddr, ToName, subject, mailtext, "")
}
