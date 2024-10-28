package email

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

type SMTPService struct {
	auth *smtp.Auth
	from string
	host string
}

var smtp_email_service *SMTPService

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
func (service *SMTPService) SendEmail(subject string, body string, emailIds ...string) {
	msg := strings.Join([]string{
		"From: " + service.from,
		"Subject:" + subject,
		"",
		body,
	}, "\r\n")
	err := smtp.SendMail(service.host+":587", *service.auth, service.from, emailIds, []byte(msg))
	fmt.Println(err)
}
