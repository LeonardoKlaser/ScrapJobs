package usecase

import (
	"context"
	"fmt"
	"web-scrapper/infra/ses"
	"web-scrapper/model"
)


type SESSenderAdapter struct {
	mailSender *ses.SESMailSender
}

func NewSESSenderAdapter(mailSender *ses.SESMailSender) *SESSenderAdapter {
	return &SESSenderAdapter{
		mailSender: mailSender,
	}
}

func (adapter *SESSenderAdapter) SendAnalysisEmail(ctx context.Context, userEmail string, job model.Job, analysis model.ResumeAnalysis) error {
	// 1. Criar o assunto do e-mail
	subject := fmt.Sprintf("Análise de Vaga Encontrada: %s", job.Title)

	// 2. Gerar o corpo do e-mail usando a função que já temos
    // (O ideal é passar o título da vaga para a análise antes para que o corpo possa exibi-lo)
	body := GenerateEmailBody(analysis)

	// 3. Chamar o serviço de baixo nível com os dados formatados
	return adapter.mailSender.SendEmail(ctx, userEmail, subject, body)
}