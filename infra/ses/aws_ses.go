package ses

import (
    "context"
    "web-scrapper/logging"
    "github.com/aws/aws-sdk-go-v2/service/ses"
    "github.com/aws/aws-sdk-go-v2/service/ses/types"
    "github.com/aws/aws-sdk-go-v2/aws"
)

// SESMailSender encapsula o cliente SES e o remetente padrão
type SESMailSender struct {
    client *ses.Client
    from   string
}

// NewSESMailSender cria uma instância do SESMailSender carregando a configuração AWS
func NewSESMailSender(sesClient *ses.Client, from string) *SESMailSender {
   return &SESMailSender{
		client: sesClient,
		from:   from,
	} 
}

// SendEmail envia um e-mail simples (texto) usando o SES
func (s *SESMailSender) SendEmail(ctx context.Context ,to string, subject string, bodyText string, bodyHtml string) error {
    input := &ses.SendEmailInput{
        Source: aws.String(s.from),
        Destination: &types.Destination{
            ToAddresses: []string{to},
        },
        Message: &types.Message{
            Subject: &types.Content{
                Data: aws.String(subject),
            },
            Body: &types.Body{
                Html: &types.Content{
                    Data:    aws.String(bodyHtml),
                    Charset: aws.String("UTF-8"),
                },
                Text: &types.Content{
                    Data:    aws.String(bodyText),
                    Charset: aws.String("UTF-8"),
                },
            },
        },
    }

    _, err := s.client.SendEmail(ctx, input)
	logging.Logger.Info().Str("subject", subject).Str("to", to).Msg("E-mail enviado")
    return err
}