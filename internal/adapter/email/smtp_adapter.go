package email

import (
	"be-ayaka/internal/core/port"
	"fmt"
	"net/smtp"
)

type smtpAdapter struct {
	host     string
	port     string
	user     string
	password string
}

func NewSMTPAdapter(host, port, user, password string) port.EmailSender {
	return &smtpAdapter{
		host:     host,
		port:     port,
		user:     user,
		password: password,
	}
}

func (s *smtpAdapter) SendEmail(to, subject, htmlBody string) error {
	auth := smtp.PlainAuth("", s.user, s.password, s.host)
	address := fmt.Sprintf("%s:%s", s.host, s.port)

	msg := fmt.Sprintf("MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\nFrom: %s\nTo: %s\nSubject: %s\n\n%s",
		s.user, to, subject, htmlBody)

	err := smtp.SendMail(address, auth, s.user, []string{to}, []byte(msg))

	return err
}
