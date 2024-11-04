package email

import (
	"fmt"
	"net/smtp"
	"os"
	"slices"
	"strings"
)

type SMTPService struct {
	auth *smtp.Auth
	from string
	host string
}

var smtp_email_service *SMTPService

var validTestEmails = []string{"crazinessspeaks@gmail.coms", "tkdazzles28@gmail.com"}

func GetSMTPService() *SMTPService {
	if smtp_email_service == nil {
		email := os.Getenv("smtp_email")
		password := os.Getenv("smtp_password")
		host := os.Getenv("smtp_host")
		tempAuth := smtp.PlainAuth("", email, password, host)

		smtp_email_service = &SMTPService{
			auth: &tempAuth,
			from: email,
			host: host,
		}
	}
	return smtp_email_service
}
func (service *SMTPService) SendEmail(subject string, body string, emailId string) {
	fmt.Println("Send Email Called")
	if !slices.Contains(validTestEmails, emailId) {
		return
	}

	msg := strings.Join([]string{
		"From: " + service.from,
		"Subject:" + subject,
		"",
		body,
	}, "\r\n")
	err := smtp.SendMail(service.host+":587", *service.auth, service.from, []string{emailId}, []byte(msg))
	fmt.Println(err)
}
