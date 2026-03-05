package resend

import (
	"context"
	"fmt"
	"web-scrapper/interfaces"
	"web-scrapper/logging"

	resendSdk "github.com/resend/resend-go/v2"
)

var _ interfaces.MailSender = (*ResendMailSender)(nil)

type ResendMailSender struct {
	client *resendSdk.Client
	from   string
}

func NewResendMailSender(apiKey, from string) *ResendMailSender {
	client := resendSdk.NewClient(apiKey)
	return &ResendMailSender{
		client: client,
		from:   from,
	}
}

func (r *ResendMailSender) SendEmail(ctx context.Context, to string, subject string, bodyText string, bodyHtml string) error {
	params := &resendSdk.SendEmailRequest{
		From:    r.from,
		To:      []string{to},
		Subject: subject,
		Html:    bodyHtml,
		Text:    bodyText,
	}

	sent, err := r.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("resend send failed: %w", err)
	}

	logging.Logger.Info().
		Str("subject", subject).
		Str("to", to).
		Str("resend_id", sent.Id).
		Msg("E-mail enviado via Resend")

	return nil
}
