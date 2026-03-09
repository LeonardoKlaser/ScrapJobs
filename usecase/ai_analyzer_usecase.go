package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"
	"web-scrapper/infra/openai"
	"web-scrapper/model"
)

type AiAnalyser struct {
	client *openai.OpenAIClient
}

func NewAiAnalyser(client *openai.OpenAIClient) *AiAnalyser {
	return &AiAnalyser{
		client: client,
	}
}

func getSystemPrompt() string {
	return `Você é um Especialista em ATS (Applicant Tracking System) e Tech Recruiter Sênior. Sua missão é cruzar os dados da Vaga fornecida com o Currículo do candidato e gerar um plano de ação tático.

REGRAS ESTRITAS (NÃO IGNORE):
1. ZERO ALUCINAÇÃO: Nunca invente habilidades, ferramentas ou experiências que não estejam explicitamente no currículo do candidato. Se a vaga pede AWS e ele não tem, aponte como um Gap. Não adapte mentiras.
2. FOCO EM ATS: Os sistemas de triagem buscam "Exact Matches" (palavras-chave exatas). Sua análise deve focar em incluir as palavras-chave da vaga no currículo do candidato de forma natural, baseando-se nas experiências REAIS dele.
3. PRAGMATISMO: Seja direto. O candidato não quer teoria, ele quer saber o que mudar.
4. TONE OF VOICE: Profissional, encorajador e altamente objetivo.

DIRETRIZES DE SAÍDA (Obrigatório seguir o JSON abaixo):
- Na seção 'actionableResumeSuggestions', você DEVE fornecer blocos de texto prontos (copiar e colar) que o candidato possa simplesmente inserir em seu currículo ou LinkedIn.
- Responda SEMPRE em Português do Brasil (PT-BR).
- Retorne EXCLUSIVAMENTE um JSON válido, sem formatação markdown externa (sem tags como ` + "```json" + `).

FORMATO JSON ESPERADO:
{
  "matchAnalysis": {
    "overallScoreNumeric": 0 (0-100%),
    "overallScoreQualitative": "Baixo, Médio, Alto ou Excelente",
    "summary": "2-3 linhas focado no porquê dessa nota baseada nos requisitos obrigatórios."
  },
  "atsKeywords": {
    "matched": ["Lista de palavras-chave da vaga que o candidato já tem"],
    "missing": ["Lista de palavras-chave da vaga que faltam no currículo"]
  },
  "strengthsForThisJob": [
    {
      "point": "Nome da força",
      "relevanceToJob": "Por que isso importa para essa vaga (1 frase)."
    }
  ],
  "gapsAndImprovementAreas": [
    {
      "areaDescription": "O que falta",
      "jobRequirementImpacted": "Qual requisito da vaga foi impactado."
    }
  ],
  "actionableResumeSuggestions": [
    {
      "curriculumSectionToApply": "Ex: Resumo Profissional, Experiência X",
      "suggestion": "Instrução curta do que fazer.",
      "exampleWording": "TEXTO PRONTO PARA COPIAR E COLAR. Reescreva a experiência real do candidato inserindo as palavras-chave que faltam, mas sem inventar mentiras.",
      "reasoningForThisJob": "Como esse texto ajuda a passar pelo filtro do ATS."
    }
  ],
  "finalConsiderations": "Uma dica final."
}`
}

func getUserPromptTemplateString() string {
	return `**1. DESCRIÇÃO DA VAGA:**
{{.JobDescriptionJSON}}

**2. CURRÍCULO DO CANDIDATO (formato JSON):**
{{.CurriculumJSON}}`
}

func prompt_builder(curriculum model.Curriculum, job model.Job) (string, error) {

	curriculumJsonBytes, err := json.MarshalIndent(curriculum, "", " ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal curriculum: %w", err)
	}

	jobDataForPrompt := struct {
		Title           string `json:"title"`
		Company         string `json:"company"`
		Location        string `json:"location"`
		DescriptionFull string `json:"description_full"`
	}{
		Title:           job.Title,
		Company:         job.Company,
		Location:        job.Location,
		DescriptionFull: job.Description,
	}

	jobDescriptionsJSONBytes, err := json.MarshalIndent(jobDataForPrompt, "", " ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal job informations: %w", err)
	}

	promptData := struct {
		CurriculumJSON     string
		JobDescriptionJSON string
	}{
		CurriculumJSON:     string(curriculumJsonBytes),
		JobDescriptionJSON: string(jobDescriptionsJSONBytes),
	}

	tmpl, err := template.New("userPrompt").Parse(getUserPromptTemplateString())
	if err != nil {
		return "", fmt.Errorf("failed to generate user prompt template: %w", err)
	}

	var populatedPrompt bytes.Buffer
	if err := tmpl.Execute(&populatedPrompt, promptData); err != nil {
		return "", fmt.Errorf("failed to execute prompt template: %w", err)
	}

	return populatedPrompt.String(), nil
}

func (a *AiAnalyser) Analyze(ctx context.Context, curriculum model.Curriculum, job model.Job) (model.ResumeAnalysis, error) {
	nullreturn := model.ResumeAnalysis{}
	if a.client == nil {
		return nullreturn, errors.New("openai client isn't initialized")
	}

	userPrompt, err := prompt_builder(curriculum, job)
	if err != nil {
		return nullreturn, fmt.Errorf("failed to build prompt for AI analysis: %w", err)
	}

	responseText, err := a.client.ChatCompletion(ctx, getSystemPrompt(), userPrompt)
	if err != nil {
		return nullreturn, fmt.Errorf("error to get prompt response: %w", err)
	}

	if responseText == "" {
		return nullreturn, fmt.Errorf("openai returned empty response")
	}

	var analysis model.ResumeAnalysis
	err = json.Unmarshal([]byte(responseText), &analysis)
	if err != nil {
		return nullreturn, fmt.Errorf("error to struct response: %w", err)
	}
	return analysis, nil
}
