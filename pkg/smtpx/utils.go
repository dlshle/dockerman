package smtpx

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/google/uuid"
)

type Mail struct {
	MessageID string
	Subject   string
	From      string
	Rcpts     []string
	Body      string
}

func SendMail(mail *Mail) error {
	if mail.MessageID == "" {
		domain := strings.Split(mail.From, "@")[1]
		mail.MessageID = fmt.Sprintf("<%s@%s>", uuid.NewString(), domain)
	}
	rcptByDomains := make(map[string][]string)
	for _, rcpt := range mail.Rcpts {
		svr, err := mxLookup(rcpt)
		if err != nil {
			return err
		}
		rcptByDomains[svr] = append(rcptByDomains[svr], rcpt)
	}
	for svr := range rcptByDomains {
		if err := sendToRecipiantsByServer(svr, mail); err != nil {
			return err
		}
	}
	return nil
}

func sendToRecipiantsByServer(svr string, mail *Mail) error {
	// Construct the email message
	header := make(map[string]string)
	header["Message-ID"] = mail.MessageID
	header["From"] = mail.From
	header["To"] = strings.Join(mail.Rcpts, ",")
	header["Subject"] = mail.Subject
	header["Content-Type"] = `text/plain; charset="UTF-8"`

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + mail.Body

	// Send the email
	err := smtp.SendMail(svr+":25", nil, mail.From, mail.Rcpts, []byte(message))
	if err != nil {
		return err
	}
	return nil
}

// mxLookup returns the primary (highest priority) MX record for a given email address.
func mxLookup(email string) (string, error) {
	// Get the domain part of the email address
	domain := strings.Split(email, "@")[1]

	// Perform MX lookup
	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		return "", err
	}

	// Check if we received any MX records
	if len(mxRecords) == 0 {
		return "", fmt.Errorf("no MX records found for domain %s", domain)
	}

	// Return the highest priority MX record host
	return mxRecords[0].Host, nil
}
