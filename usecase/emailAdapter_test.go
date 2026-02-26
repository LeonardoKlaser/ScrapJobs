package usecase

import (
	"testing"
	"web-scrapper/model"

	"github.com/stretchr/testify/assert"
)

func TestGenerateEmailBodyHTML(t *testing.T) {
	t.Run("should generate HTML with job title and score", func(t *testing.T) {
		analysis := model.ResumeAnalysis{
			MatchAnalysis: model.MatchAnalysis{
				OverallScoreNumeric:     85,
				OverallScoreQualitative: "Bom",
				Summary:                 "Good match for the position",
			},
			StrengthsForThisJob: []model.Strength{
				{Point: "Go experience", RelevanceToJob: "Direct match with required skill"},
			},
			GapsAndImprovementAreas: []model.Gap{
				{AreaDescription: "Kubernetes", JobRequirementImpacted: "Infrastructure management"},
			},
			ActionableResumeSuggestions: []model.Suggestion{
				{Suggestion: "Add K8s", CurriculumSectionToApply: "Skills", ExampleWording: "Kubernetes cluster management", ReasoningForThisJob: "Key requirement"},
			},
			FinalConsiderations: "Strong candidate overall",
		}
		job := model.Job{ID: 1, Title: "Go Developer", Company: "Acme"}

		html, err := generateEmailBodyHTML(analysis, job)

		assert.NoError(t, err)
		assert.Contains(t, html, "Go Developer")
		assert.Contains(t, html, "85")
		assert.Contains(t, html, "Bom")
		assert.Contains(t, html, "Go experience")
		assert.Contains(t, html, "Kubernetes")
		assert.Contains(t, html, "Add K8s")
		assert.Contains(t, html, "Strong candidate overall")
	})

	t.Run("should handle empty strengths and gaps", func(t *testing.T) {
		analysis := model.ResumeAnalysis{
			MatchAnalysis: model.MatchAnalysis{
				OverallScoreNumeric:     50,
				OverallScoreQualitative: "Mediano",
				Summary:                 "Average match",
			},
			StrengthsForThisJob:        []model.Strength{},
			GapsAndImprovementAreas:    []model.Gap{},
			ActionableResumeSuggestions: []model.Suggestion{},
			FinalConsiderations:        "Needs improvement",
		}
		job := model.Job{Title: "Python Dev"}

		html, err := generateEmailBodyHTML(analysis, job)

		assert.NoError(t, err)
		assert.Contains(t, html, "Nenhum ponto forte específico identificado")
		assert.Contains(t, html, "Nenhuma lacuna específica identificada")
		assert.Contains(t, html, "Nenhuma sugestão específica identificada")
	})
}

func TestGenerateEmailBodyText(t *testing.T) {
	t.Run("should generate text with job title and score", func(t *testing.T) {
		analysis := model.ResumeAnalysis{
			MatchAnalysis: model.MatchAnalysis{
				OverallScoreNumeric:     75,
				OverallScoreQualitative: "Bom",
				Summary:                 "Solid match",
			},
			StrengthsForThisJob: []model.Strength{
				{Point: "React skills", RelevanceToJob: "Frontend position"},
			},
			FinalConsiderations: "Good fit",
		}
		job := model.Job{Title: "Frontend Dev"}

		text := generateEmailBodyText(analysis, job)

		assert.Contains(t, text, "Frontend Dev")
		assert.Contains(t, text, "75")
		assert.Contains(t, text, "Bom")
		assert.Contains(t, text, "React skills")
		assert.Contains(t, text, "Good fit")
	})
}

func TestGenerateWelcomeEmailBodyHTML(t *testing.T) {
	t.Run("should generate welcome HTML with name and link", func(t *testing.T) {
		html, err := generateWelcomeEmailBodyHTML("João Silva", "https://scrapjobs.com.br/app")

		assert.NoError(t, err)
		assert.Contains(t, html, "João Silva")
		assert.Contains(t, html, "https://scrapjobs.com.br/app")
		assert.Contains(t, html, "Bem-vindo")
	})
}

func TestGenerateWelcomeEmailBodyText(t *testing.T) {
	t.Run("should generate welcome text with name and link", func(t *testing.T) {
		text := generateWelcomeEmailBodyText("Maria", "https://scrapjobs.com.br/app")

		assert.Contains(t, text, "Maria")
		assert.Contains(t, text, "https://scrapjobs.com.br/app")
		assert.Contains(t, text, "Bem-vindo")
	})
}

func TestGenerateNewJobsEmailBodyHTML(t *testing.T) {
	t.Run("should generate new jobs HTML with table", func(t *testing.T) {
		jobs := []*model.Job{
			{Title: "Go Dev", Company: "Acme", Location: "Remote", JobLink: "https://acme.com/1"},
			{Title: "Python Dev", Company: "Beta", Location: "SP", JobLink: "https://beta.com/2"},
		}

		html, err := generateNewJobsEmailBodyHTML("Carlos", jobs)

		assert.NoError(t, err)
		assert.Contains(t, html, "Carlos")
		assert.Contains(t, html, "Go Dev")
		assert.Contains(t, html, "Python Dev")
		assert.Contains(t, html, "Acme")
		assert.Contains(t, html, "Beta")
		assert.Contains(t, html, "https://acme.com/1")
		assert.Contains(t, html, "https://beta.com/2")
	})

	t.Run("should handle multiple jobs in table", func(t *testing.T) {
		jobs := []*model.Job{
			{Title: "Job1", Company: "C1", Location: "L1", JobLink: "https://c1.com/1"},
			{Title: "Job2", Company: "C2", Location: "L2", JobLink: "https://c2.com/2"},
			{Title: "Job3", Company: "C3", Location: "L3", JobLink: "https://c3.com/3"},
		}

		html, err := generateNewJobsEmailBodyHTML("Ana", jobs)

		assert.NoError(t, err)
		assert.Contains(t, html, "Job1")
		assert.Contains(t, html, "Job2")
		assert.Contains(t, html, "Job3")
	})
}

func TestGenerateNewJobsEmailBodyText(t *testing.T) {
	t.Run("should generate text with job list", func(t *testing.T) {
		jobs := []*model.Job{
			{Title: "Go Dev", Company: "Acme", Location: "Remote", JobLink: "https://acme.com/1"},
		}

		text := generateNewJobsEmailBodyText("Pedro", jobs)

		assert.Contains(t, text, "Pedro")
		assert.Contains(t, text, "Go Dev")
		assert.Contains(t, text, "Acme")
		assert.Contains(t, text, "Remote")
		assert.Contains(t, text, "https://acme.com/1")
		assert.Contains(t, text, "1 nova(s) vaga(s)")
	})
}
