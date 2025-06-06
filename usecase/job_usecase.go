package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"web-scrapper/infra/ses"
	"web-scrapper/model"
	"web-scrapper/repository"
	"web-scrapper/scrapper"
)

type JobUseCase struct{
	Repository repository.JobRepository
	MailSender *ses.SESMailSender
	aiAnalyser AiAnalyser
	user UserUsecase
	curriculum CurriculumUsecase
}

func NewJobUseCase(jobRepo repository.JobRepository, mailSender *ses.SESMailSender, analyser AiAnalyser, usr UserUsecase, curric CurriculumUsecase ) JobUseCase{
	return JobUseCase{
		Repository: jobRepo,
		MailSender: mailSender,
		aiAnalyser: analyser,
		user: usr,
		curriculum: curric,
	}
}

func (job JobUseCase) CreateJob(jobData model.Job) (int, error){
	jobID , err := job.Repository.CreateJob(jobData);
	if(err != nil){
		return jobID, err
	}

	return jobID, nil
}

func (job JobUseCase) FindJobByRequisitionID(requisition_ID int) (bool, error){
	hasJob, err := job.Repository.FindJobByRequisitionID(requisition_ID);
	if(err != nil){
		return false, err
	}

	return hasJob, nil
}

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


func (uc *JobUseCase) ScrapeAndStoreJobs(ctx context.Context) ([]*model.Job, error) {
    jobs, err := scrapper.NewJobScraper().ScrapeJobs()
    if err != nil {
        return nil, err
    }
    var newJobsToDatabase []*model.Job
    for _, job := range jobs {
        exist, err := uc.Repository.FindJobByRequisitionID(job.Requisition_ID)
		
        if err != nil {
            return nil, err
        }
        if !exist {
            newJobsToDatabase = append(newJobsToDatabase, job)
			jobToInsert := model.Job{
				Title: job.Title,
				Location: job.Location,
				Company: job.Company,
				Job_link: job.Job_link,
				Requisition_ID: job.Requisition_ID,
			}
            uc.Repository.CreateJob(jobToInsert)
			user, err := uc.user.GetUserByEmail("leobkklaser@gmail.com")
			if err != nil {
				return nil, err
			}

			curriculum, err := uc.curriculum.GetCurriculumByUserId(user.Id)
			if err != nil {
				return nil, err
			}
			matchAnaliser, err := uc.aiAnalyser.AiAnalyzerMatch(ctx, curriculum, *job)
			if err != nil {
				return nil, err
			}
			log.Println("aqui")
			emailBody := "Uma nova vaga foi encontrada: " + job.Title + "\n" +
					"Link para saber mais sobre a vaga: " + "https://jobs.sap.com" + job.Job_link + "\n\n" + generateEmailBody(matchAnaliser)

			log.Println(emailBody)
			if uc.MailSender != nil{
				subject := "Nova Vaga Encontrada: " + job.Title
				body := emailBody
				to := "leobkklaser@gmail.com"

				go func() {
					err := uc.MailSender.SendEmail(ctx, to, subject, body)
					if err != nil {
						println("Erro ao enviar e-mail:", err.Error())
						return
					}
				}()
			}
        }
    }
    return newJobsToDatabase, nil
}
