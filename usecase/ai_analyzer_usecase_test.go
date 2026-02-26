package usecase

import (
	"context"
	"testing"
	"web-scrapper/model"

	"github.com/stretchr/testify/assert"
)

func TestPromptBuilder(t *testing.T) {
	t.Run("should generate prompt with curriculum and job data", func(t *testing.T) {
		curriculum := model.Curriculum{
			Id:     1,
			Title:  "Senior Go Developer",
			Skills: "Go, Docker, Kubernetes",
			Summary: "Experienced developer",
			Experiences: []model.Experience{
				{Company: "Acme", Title: "Developer", Description: "Built APIs"},
			},
			Educations: []model.Education{
				{Institution: "USP", Degree: "CS", Year: "2020"},
			},
			Languages: "Portuguese, English",
		}
		job := model.Job{
			Title:       "Go Developer",
			Company:     "TechCorp",
			Location:    "Remote",
			Description: "We need a Go developer with K8s experience",
		}

		prompt, err := prompt_builder(curriculum, job)

		assert.NoError(t, err)
		assert.Contains(t, prompt, "Senior Go Developer")
		assert.Contains(t, prompt, "Go, Docker, Kubernetes")
		assert.Contains(t, prompt, "Go Developer")
		assert.Contains(t, prompt, "TechCorp")
		assert.Contains(t, prompt, "We need a Go developer with K8s experience")
		assert.Contains(t, prompt, "DESCRIÇÃO DA VAGA")
		assert.Contains(t, prompt, "CURRÍCULO DO CANDIDATO")
	})

	t.Run("should serialize curriculum and job as JSON in template", func(t *testing.T) {
		curriculum := model.Curriculum{
			Id:    2,
			Title: "Junior Dev",
			Experiences: []model.Experience{
				{Company: "StartupX", Title: "Intern", Description: "Frontend work"},
			},
		}
		job := model.Job{
			Title:       "Frontend Engineer",
			Company:     "BigCo",
			Location:    "São Paulo",
			Description: "React and TypeScript",
		}

		prompt, err := prompt_builder(curriculum, job)

		assert.NoError(t, err)
		// Verify JSON serialization happened
		assert.Contains(t, prompt, "Junior Dev")
		assert.Contains(t, prompt, "StartupX")
		assert.Contains(t, prompt, "Frontend Engineer")
		assert.Contains(t, prompt, "BigCo")
	})
}

func TestAiAnalyser_Analyze_NilClient(t *testing.T) {
	t.Run("should return error when client is nil", func(t *testing.T) {
		analyser := NewAiAnalyser(nil)

		curriculum := model.Curriculum{Id: 1}
		job := model.Job{ID: 1}

		_, err := analyser.Analyze(context.Background(), curriculum, job)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client isn't initialized")
	})
}
