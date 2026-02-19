package usecase

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"web-scrapper/infra/ses"
	"web-scrapper/model"
	"html/template"
)

func generateWelcomeEmailBodyHTML(userName, dashboardLink string) (string, error) {
	const templateStr = `
    <!DOCTYPE html><html><head><meta charset="UTF-8"><style>/* Estilos aqui */
        .button { background-color: #28a745; color: white; padding: 12px 25px; text-align: center; text-decoration: none; display: inline-block; border-radius: 5px; font-weight: bold; }
    </style></head>
    <body><h2>Bem-vindo(a) ao ScrapJobs, {{.UserName}}!</h2>
    <p>Sua conta foi criada com sucesso!</p>
    <p>Agora você pode começar a automatizar sua busca por vagas e receber análises personalizadas diretamente no seu e-mail.</p>
    <p>Acesse seu painel para configurar os sites que deseja monitorar e fazer upload do seu currículo:</p>
    <p style="text-align: center; margin-top: 25px; margin-bottom: 25px;">
        <a href="{{.DashboardLink}}" class="button">Acessar meu Dashboard</a>
    </p>
    <p>Se tiver alguma dúvida, responda a este e-mail ou contate nosso suporte.</p>
    <p>Atenciosamente,<br/>Equipe ScrapJobs</p></body></html>`

	data := struct {
		UserName      string
		DashboardLink string
	}{UserName: userName, DashboardLink: dashboardLink}

	tmpl, err := template.New("welcomeEmail").Parse(templateStr)
	if err != nil {
		return "", err
	}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return "", err
	}
	return body.String(), nil
}

func generateWelcomeEmailBodyText(userName, dashboardLink string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Bem-vindo(a) ao ScrapJobs, %s!\n\n", userName))
	sb.WriteString("Sua conta foi criada com sucesso!\n\n")
	sb.WriteString("Agora você pode começar a automatizar sua busca por vagas e receber análises personalizadas diretamente no seu e-mail.\n\n")
	sb.WriteString("Acesse seu painel para configurar os sites que deseja monitorar e fazer upload do seu currículo:\n")
	sb.WriteString(dashboardLink + "\n\n")
	sb.WriteString("Se tiver alguma dúvida, responda a este e-mail ou contate nosso suporte.\n\n")
	sb.WriteString("Atenciosamente,\nEquipe ScrapJobs\n")
	return sb.String()
}

func generateEmailBodyHTML(analysis model.ResumeAnalysis, job model.Job) (string, error) {
    const emailTemplate = `
    <!DOCTYPE html>
    <html lang="pt-BR">
    <head>
        <meta charset="UTF-8">
        <style>
            body { font-family: sans-serif; line-height: 1.6; color: #333; }
            h2 { color: #0056b3; }
            h3 { border-bottom: 1px solid #eee; padding-bottom: 5px; }
            .card { border: 1px solid #ddd; border-radius: 5px; padding: 15px; margin-bottom: 20px; background-color: #f9f9f9; }
            .score { font-size: 1.2em; font-weight: bold; }
            ul { list-style-type: none; padding-left: 0; }
            li { margin-bottom: 10px; }
        </style>
    </head>
    <body>
        <h2>Análise de Vaga Encontrada: {{.Job.Title}}</h2>
        <p>Prezados(as),</p>
        <p>Segue abaixo a análise detalhada da compatibilidade do seu currículo com a vaga encontrada.</p>

        <div class="card">
            <h3>Análise de Compatibilidade</h3>
            <ul>
                <li><strong>Pontuação Geral:</strong> <span class="score">{{.Analysis.MatchAnalysis.OverallScoreNumeric}}</span></li>
                <li><strong>Avaliação Qualitativa:</strong> {{.Analysis.MatchAnalysis.OverallScoreQualitative}}</li>
                <li><strong>Resumo:</strong> {{.Analysis.MatchAnalysis.Summary}}</li>
            </ul>
        </div>

        <div class="card">
            <h3>Pontos Fortes para esta Vaga</h3>
            <ul>
                {{range .Analysis.StrengthsForThisJob}}
                    <li><strong>Ponto:</strong> {{.Point}}<br/><em>Relevância:</em> {{.RelevanceToJob}}</li>
                {{else}}
                    <li>Nenhum ponto forte específico identificado.</li>
                {{end}}
            </ul>
        </div>

        <div class="card">
            <h3>Lacunas e Áreas de Melhoria</h3>
            <ul>
                {{range .Analysis.GapsAndImprovementAreas}}
                    <li><strong>Área:</strong> {{.AreaDescription}}<br/><em>Impacto:</em> {{.JobRequirementImpacted}}</li>
                {{else}}
                    <li>Nenhuma lacuna específica identificada.</li>
                {{end}}
            </ul>
        </div>

        <div class="card">
            <h3>Sugestões para o Currículo</h3>
            <ul>
                {{range .Analysis.ActionableResumeSuggestions}}
                    <li><strong>Sugestão:</strong> {{.Suggestion}}<br/><em>Seção:</em> {{.CurriculumSectionToApply}}<br/><em>Exemplo:</em> "{{.ExampleWording}}"<br/><em>Justificativa:</em> {{.ReasoningForThisJob}}</li>
                {{else}}
                    <li>Nenhuma sugestão específica identificada.</li>
                {{end}}
            </ul>
        </div>
        
        <h3>Considerações Finais</h3>
        <p>{{.Analysis.FinalConsiderations}}</p>
        <br/>
        <p>Atenciosamente,<br/>Equipe ScrapJobs</p>
    </body>
    </html>
    `

    data := struct {
        Analysis model.ResumeAnalysis
        Job      model.Job
    }{
        Analysis: analysis,
        Job:      job,
    }

    tmpl, err := template.New("email").Parse(emailTemplate)
    if err != nil {
        return "", fmt.Errorf("ERROR to analyse email template: %w", err)
    }

    var body bytes.Buffer
    if err := tmpl.Execute(&body, data); err != nil {
        return "", fmt.Errorf("ERROR to execute email template: %w", err)
    }

    return body.String(), nil
}


func generateEmailBodyText(analysis model.ResumeAnalysis, job model.Job) string {
	var sb strings.Builder


	sb.WriteString("Prezados(as),\n\n")
	sb.WriteString(fmt.Sprintf("Segue abaixo a análise detalhada do currículo para a posição de %s:\n\n", job.Title))

	sb.WriteString("**Análise de Compatibilidade:**\n")
	sb.WriteString(fmt.Sprintf("*   **Pontuação Geral:** %d\n", analysis.MatchAnalysis.OverallScoreNumeric))
	sb.WriteString(fmt.Sprintf("*   **Avaliação Qualitativa:** %s\n", analysis.MatchAnalysis.OverallScoreQualitative))
	sb.WriteString(fmt.Sprintf("*   **Resumo:** %s\n", analysis.MatchAnalysis.Summary))
	sb.WriteString("\n---\n\n")

	sb.WriteString("**Pontos Fortes para esta Vaga:**\n\n")
	if len(analysis.StrengthsForThisJob) > 0 {
		for i, strength := range analysis.StrengthsForThisJob {
			sb.WriteString(fmt.Sprintf("%d.  **Ponto:** %s\n", i+1, strength.Point))
			sb.WriteString(fmt.Sprintf("    *   **Relevância:** %s\n\n", strength.RelevanceToJob))
		}
	} else {
		sb.WriteString("Nenhum ponto forte específico identificado.\n\n")
	}
	sb.WriteString("---\n\n")

	sb.WriteString("**Lacunas e Áreas de Melhoria:**\n\n")
	if len(analysis.GapsAndImprovementAreas) > 0 {
		for i, gap := range analysis.GapsAndImprovementAreas {
			sb.WriteString(fmt.Sprintf("%d.  **Área:** %s\n", i+1, gap.AreaDescription))
			sb.WriteString(fmt.Sprintf("    *   **Impacto no Requisito da Vaga:** %s\n\n", gap.JobRequirementImpacted))
		}
	} else {
		sb.WriteString("Nenhuma lacuna ou área de melhoria específica identificada.\n\n")
	}
	sb.WriteString("---\n\n")

	sb.WriteString("**Sugestões Acionáveis para o Currículo:**\n\n")
	if len(analysis.ActionableResumeSuggestions) > 0 {
		for i, suggestion := range analysis.ActionableResumeSuggestions {
			sb.WriteString(fmt.Sprintf("%d.  **Sugestão:** %s\n", i+1, suggestion.Suggestion))
			sb.WriteString(fmt.Sprintf("    *   **Seção do Currículo:** %s\n", suggestion.CurriculumSectionToApply))
			sb.WriteString(fmt.Sprintf("    *   **Exemplo de Redação:** %s\n", suggestion.ExampleWording))
			sb.WriteString(fmt.Sprintf("    *   **Justificativa:** %s\n\n", suggestion.ReasoningForThisJob))
		}
	} else {
		sb.WriteString("Nenhuma sugestão acionável para o currículo identificada.\n\n")
	}
	sb.WriteString("---\n\n")

	sb.WriteString("**Considerações Finais:**\n")
	sb.WriteString(analysis.FinalConsiderations)
	sb.WriteString("\n\n---\n")


	sb.WriteString("Atenciosamente,\n\n")
	sb.WriteString("Equipe ScrapJobs\n")

	return sb.String()
}

type SESSenderAdapter struct {
	mailSender *ses.SESMailSender
}

func NewSESSenderAdapter(mailSender *ses.SESMailSender) *SESSenderAdapter {
	return &SESSenderAdapter{
		mailSender: mailSender,
	}
}

func (adapter *SESSenderAdapter) SendAnalysisEmail(ctx context.Context, userEmail string, job model.Job, analysis model.ResumeAnalysis) error {
	subject := fmt.Sprintf("Análise de Vaga Encontrada: %s", job.Title)

	bodyHtml, err := generateEmailBodyHTML(analysis, job)
    if err != nil {
        return fmt.Errorf("ERROR to generate html body: %w", err)
    }

    bodyText := generateEmailBodyText(analysis, job)

    
    return adapter.mailSender.SendEmail(ctx, userEmail, subject, bodyText, bodyHtml)
}

func (adapter *SESSenderAdapter) SendWelcomeEmail(ctx context.Context, userEmail, userName, dashboardLink string) error {
	subject := "Bem-vindo(a) ao ScrapJobs!"

	bodyHtml, err := generateWelcomeEmailBodyHTML(userName, dashboardLink)
	if err != nil {
		return fmt.Errorf("erro ao gerar corpo HTML do email de boas-vindas: %w", err)
	}

	bodyText := generateWelcomeEmailBodyText(userName, dashboardLink)

	return adapter.mailSender.SendEmail(ctx, userEmail, subject, bodyText, bodyHtml)
}