package ses

import (
    "context"
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
func NewSESMailSender(sesClient *ses.Client, from string) (*SESMailSender) {
   return &SESMailSender{
		client: sesClient,
		from:   from,
	} 
}

// SendEmail envia um e-mail simples (texto) usando o SES
func (s *SESMailSender) SendEmail(ctx context.Context ,to string, subject string, body string) error {
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
                Text: &types.Content{
                    Data: aws.String(body),
                },
            },
        },
    }

    _, err := s.client.SendEmail(ctx, input)
	println("enviou o e-mail: " + subject + " para: " + to + "\n ")
    return err
}