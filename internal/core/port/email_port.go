package port

type EmailSender interface {
	SendEmail(to, subject, htmlBody string) error
}