package usecase

import (
	"context"
	"fmt"
	"web-scrapper/infra/ses"
	"web-scrapper/model"
	"strings"
)

func generateEmailBody (analysis model.ResumeAnalysis) string {
	var sb strings.Builder


	sb.WriteString("Prezados(as),\n\n")
	sb.WriteString("Segue abaixo a análise detalhada do currículo para a posição de Senior Developer - SAP Datasphere Repository:\n\n")

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
	sb.WriteString("[Seu Nome/Nome da Equipe]\n") // Placeholder para assinatura

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

	body := generateEmailBody(analysis)

	return adapter.mailSender.SendEmail(ctx, userEmail, subject, body)
}