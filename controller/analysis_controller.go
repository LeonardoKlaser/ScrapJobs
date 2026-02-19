package controller

import (
	"net/http"
	"web-scrapper/interfaces"
	"web-scrapper/logging"
	"web-scrapper/model"

	"github.com/gin-gonic/gin"
)

type AnalysisController struct {
	analysisService    interfaces.AnalysisService
	curriculumRepo     interfaces.CurriculumRepositoryInterface
	jobRepo            interfaces.JobRepositoryInterface
	notificationRepo   interfaces.NotificationRepositoryInterface
	planRepo           interfaces.PlanRepositoryInterface
}

func NewAnalysisController(
	analysisService interfaces.AnalysisService,
	curriculumRepo interfaces.CurriculumRepositoryInterface,
	jobRepo interfaces.JobRepositoryInterface,
	notificationRepo interfaces.NotificationRepositoryInterface,
	planRepo interfaces.PlanRepositoryInterface,
) *AnalysisController {
	return &AnalysisController{
		analysisService:  analysisService,
		curriculumRepo:   curriculumRepo,
		jobRepo:          jobRepo,
		notificationRepo: notificationRepo,
		planRepo:         planRepo,
	}
}

type analyzeJobRequest struct {
	JobID int `json:"job_id" binding:"required"`
}

// AnalyzeJob executa uma análise de IA manual para um job específico.
// POST /api/analyze-job
func (ac *AnalysisController) AnalyzeJob(ctx *gin.Context) {
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	user, ok := userInterface.(model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Tipo de usuário inválido no contexto"})
		return
	}

	var body analyzeJobRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "job_id é obrigatório"})
		return
	}

	// Buscar currículo ativo do usuário
	curricula, err := ac.curriculumRepo.FindCurriculumByUserID(user.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar currículo"})
		return
	}

	var activeCurriculum *model.Curriculum
	for i := range curricula {
		if curricula[i].IsActive {
			activeCurriculum = &curricula[i]
			break
		}
	}
	if activeCurriculum == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Nenhum currículo ativo encontrado. Crie ou ative um currículo antes de analisar vagas."})
		return
	}

	// Verificar quota de análises do plano
	plan, err := ac.planRepo.GetPlanByUserID(user.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar plano do usuário"})
		return
	}

	if plan == nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Nenhum plano associado ao usuário. Assine um plano para usar análises de IA."})
		return
	}

	if plan.MaxAIAnalyses > 0 {
		monthlyCount, countErr := ac.notificationRepo.GetMonthlyAnalysisCount(user.Id)
		if countErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar limite de análises"})
			return
		}
		if monthlyCount >= plan.MaxAIAnalyses {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Limite de análises de IA atingido para este mês. Faça upgrade do seu plano para mais análises."})
			return
		}
	}

	// Buscar job
	job, err := ac.jobRepo.GetJobByID(body.JobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar vaga"})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Vaga não encontrada"})
		return
	}

	// Executar análise de IA
	analysis, err := ac.analysisService.Analyze(ctx.Request.Context(), *activeCurriculum, *job)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao executar análise de IA"})
		return
	}

	// Registrar notificação (a análise já foi feita, então logamos o erro mas retornamos o resultado)
	if err := ac.notificationRepo.InsertNewNotification(job.ID, user.Id); err != nil {
		logging.Logger.Error().Err(err).Int("job_id", job.ID).Int("user_id", user.Id).Msg("Erro ao registrar notificação de análise")
	}

	ctx.JSON(http.StatusOK, analysis)
}
