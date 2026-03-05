package interfaces

import "context"

// MailSender is the low-level contract for sending raw emails.
// Both SESMailSender and ResendMailSender implement this.
type MailSender interface {
	SendEmail(ctx context.Context, to string, subject string, bodyText string, bodyHtml string) error
}
