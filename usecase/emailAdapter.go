package usecase

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"
	"web-scrapper/interfaces"
	"web-scrapper/model"
	"web-scrapper/templates"
)

var emailTemplates *template.Template

func init() {
	var err error
	emailTemplates, err = template.ParseFS(templates.EmailTemplates, "emails/*.html")
	if err != nil {
		panic(fmt.Sprintf("failed to parse email templates: %v", err))
	}
}

func generateWelcomeEmailBodyHTML(userName, dashboardLink string) (string, error) {
	data := struct {
		UserName      string
		DashboardLink string
	}{UserName: userName, DashboardLink: dashboardLink}

	var body bytes.Buffer
	if err := emailTemplates.ExecuteTemplate(&body, "welcome.html", data); err != nil {
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
	data := struct {
		Analysis      model.ResumeAnalysis
		Job           model.Job
		DashboardLink string
	}{
		Analysis:      analysis,
		Job:           job,
		DashboardLink: os.Getenv("FRONTEND_URL") + "/app",
	}

	var body bytes.Buffer
	if err := emailTemplates.ExecuteTemplate(&body, "job-analysis.html", data); err != nil {
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
	mailSender interfaces.MailSender
}

func NewSESSenderAdapter(mailSender interfaces.MailSender) *SESSenderAdapter {
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

func generateNewJobsEmailBodyHTML(userName string, jobs []*model.Job) (string, error) {
	data := struct {
		UserName      string
		Jobs          []*model.Job
		DashboardLink string
	}{UserName: userName, Jobs: jobs, DashboardLink: os.Getenv("FRONTEND_URL") + "/app"}

	var body bytes.Buffer
	if err := emailTemplates.ExecuteTemplate(&body, "new-jobs-alert.html", data); err != nil {
		return "", err
	}
	return body.String(), nil
}

func generateNewJobsEmailBodyText(userName string, jobs []*model.Job) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Novas vagas encontradas para você, %s!\n\n", userName))
	sb.WriteString(fmt.Sprintf("Encontramos %d nova(s) vaga(s) nos sites que você está monitorando:\n\n", len(jobs)))
	for i, job := range jobs {
		sb.WriteString(fmt.Sprintf("%d. %s — %s (%s)\n   Link: %s\n\n", i+1, job.Title, job.Company, job.Location, job.JobLink))
	}
	sb.WriteString("Acesse seu painel no ScrapJobs para analisar essas vagas com IA.\n\n")
	sb.WriteString("Atenciosamente,\nEquipe ScrapJobs\n")
	return sb.String()
}

func (adapter *SESSenderAdapter) SendNewJobsEmail(ctx context.Context, userEmail string, userName string, jobs []*model.Job) error {
	subject := fmt.Sprintf("ScrapJobs: %d nova(s) vaga(s) encontrada(s)!", len(jobs))

	bodyHtml, err := generateNewJobsEmailBodyHTML(userName, jobs)
	if err != nil {
		return fmt.Errorf("erro ao gerar corpo HTML do email de novas vagas: %w", err)
	}

	bodyText := generateNewJobsEmailBodyText(userName, jobs)

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

func generatePasswordResetEmailHTML(userName, resetLink string) (string, error) {
	data := struct {
		UserName  string
		ResetLink string
	}{UserName: userName, ResetLink: resetLink}

	var body bytes.Buffer
	if err := emailTemplates.ExecuteTemplate(&body, "password-reset.html", data); err != nil {
		return "", err
	}
	return body.String(), nil
}

func (adapter *SESSenderAdapter) SendPasswordResetEmail(ctx context.Context, email, userName, resetLink string) error {
	subject := "ScrapJobs — Redefinição de Senha"

	bodyHTML, err := generatePasswordResetEmailHTML(userName, resetLink)
	if err != nil {
		return fmt.Errorf("erro ao gerar corpo HTML do email de redefinição de senha: %w", err)
	}

	bodyText := fmt.Sprintf("Olá %s, clique no link para redefinir sua senha: %s (válido por 1 hora)", userName, resetLink)

	return adapter.mailSender.SendEmail(ctx, email, subject, bodyText, bodyHTML)
}
