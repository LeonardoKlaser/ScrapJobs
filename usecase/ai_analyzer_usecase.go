package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"
	"strings"
	"web-scrapper/infra/gemini"
	"web-scrapper/model"
)

type AiAnalyser struct{
	client *gemini.GeminiClient
}

func NewAiAnalyser (configuration *gemini.GeminiClient) *AiAnalyser{
	return &AiAnalyser{
		client: configuration,
	}
}

func getPromptTemplateString() string{
	return `Você é um Analista de Carreira e Otimizador de Currículos altamente especializado, com vasta experiência em identificar o alinhamento entre candidatos e vagas de emprego em tecnologia, especialmente na área de desenvolvimento de software e engenharia. Sua missão é analisar de forma crítica e detalhada a descrição de uma vaga e o currículo de um candidato, fornecendo uma avaliação completa do 'match', sugestões concretas de melhoria no currículo (indicando onde aplicá-las) e destacando os pontos fortes do candidato para a vaga em questão. Sua análise deve ser profunda, objetiva e fornecer insights acionáveis para o candidato.

		**Contexto Fornecido:**

		**1. DESCRIÇÃO DA VAGA:**
		{{.JobDescriptionJSON}}

		**2. CURRÍCULO DO CANDIDATO (formato JSON):**
		{{.CurriculumJSON}}

		**Sua Tarefa Detalhada:**
		Com base EXCLUSIVAMENTE nos dados fornecidos acima (Descrição da Vaga e Currículo do Candidato), realize a seguinte análise e forneça a resposta no formato JSON especificado abaixo:
		1.  **Avaliação Geral do Match (Pontuação e Qualitativo):**
			*   Atribua uma pontuação numérica de 0 a 100 para o "match".
			*   Forneça uma avaliação qualitativa do match.
			*   Apresente um resumo conciso justificando.
		2.  **Destaque dos Pontos Fortes para ESTA VAGA:**
			*   Identifique 3 a 5 pontos fortes do currículo alinhados à vaga.
			*   Explique a relevância.
		3.  **Identificação de Gaps e Áreas de Melhoria para ESTA VAGA:**
			*   Identifique 2 a 4 principais lacunas.
		4.  **Sugestões de Melhoria ACIONÁVEIS para o Currículo (foco nesta vaga):**
			*   Forneça sugestões concretas e onde aplicá-las no currículo.
			*   Se possível, dê exemplos de texto.
		5.  **Responda em PT-br**

		**Formato da Resposta Esperado (JSON):**
		` + "```json\n" +
				`{
		"matchAnalysis": {
			"overallScoreNumeric": 0,
			"overallScoreQualitative": "",
			"summary": ""
		},
		"strengthsForThisJob": [
			{
			"point": "",
			"relevanceToJob": ""
			}
		],
		"gapsAndImprovementAreas": [
			{
			"areaDescription": "",
			"jobRequirementImpacted": ""
			}
		],
		"actionableResumeSuggestions": [
			{
			"suggestion": "",
			"curriculumSectionToApply": "",
			"exampleWording": "",
			"reasoningForThisJob": ""
			}
		],
		"finalConsiderations": ""
		}` + "\n ```"
}

func prompt_builder(curriculum model.Curriculum, job model.Job) (string, error) {
	
	curriculumJsonBytes, err := json.MarshalIndent(curriculum, "", " ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal curriculum: %w", err)
	}
	
	jobDataForPrompt := struct{
		Title string `json:"title"`
		Company string `json:"company"`
		Location        string `json:"location"`
		DescriptionFull string `json:"description_full"`
	}{
		Title: job.Title,
		Company: job.Company,
		Location: job.Location,
		DescriptionFull: job.Description,
	}
	
	jobDescriptionsJSONBytes, err := json.MarshalIndent(jobDataForPrompt, "", " ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal job informations: %w", err)
	}
	
	promptData := struct{
		CurriculumJSON string
		JobDescriptionJSON string
	}{
		CurriculumJSON: string(curriculumJsonBytes),
		JobDescriptionJSON: string(jobDescriptionsJSONBytes),
	}
	
	tmpl, err := template.New("jobMatchPrompt").Parse(getPromptTemplateString())
	if err != nil {
		return "", fmt.Errorf("failed to generate jobMatch template to return: %w", err)
	}
	
	var populatedPrompt bytes.Buffer
	if err := tmpl.Execute(&populatedPrompt, promptData); err != nil {
		return "", fmt.Errorf("failed to execute prompt template: %w", err)
	}
	
	return populatedPrompt.String(), nil
}

func extractJSON(text string) string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start == -1 || end == -1 || end <= start {
		return text
	}
	return text[start : end+1]
}

func (a *AiAnalyser) Analyze(ctx context.Context ,curriculum model.Curriculum, job model.Job) (model.ResumeAnalysis, error) {
	nullreturn := model.ResumeAnalysis{}
	if a.client == nil {
		return nullreturn, errors.New("gemini´s client isn´t initialized")
	}
	
	prompt , err := prompt_builder(curriculum, job)
	if err != nil {
		return nullreturn, fmt.Errorf("failed to build prompt for AI analysis: %w", err)
	}
	
	response, err := a.client.GeminiSearch(ctx, prompt)
	if err != nil {
		return nullreturn, fmt.Errorf("error to get prompt response: %w", err)
	}

	if response == nil || len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nullreturn, fmt.Errorf("gemini returned empty response with no candidates")
	}

	responseText := response.Text()
	if responseText == "" {
		return nullreturn, fmt.Errorf("no text content returned from gemini: %s", responseText)
	}

	var analysis model.ResumeAnalysis
	cleanedJSON := extractJSON(responseText)

	err = json.Unmarshal([]byte(cleanedJSON), &analysis)
	if err != nil {
		return nullreturn, fmt.Errorf("error to struct response: %w", err)
	}
	return analysis, nil
}
