package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"web-scrapper/interfaces"
	"web-scrapper/logging"
	"web-scrapper/model"

	"github.com/gin-gonic/gin"
)

type AnalysisController struct {
	analysisService  interfaces.AnalysisService
	curriculumRepo   interfaces.CurriculumRepositoryInterface
	jobRepo          interfaces.JobRepositoryInterface
	notificationRepo interfaces.NotificationRepositoryInterface
	planRepo         interfaces.PlanRepositoryInterface
	emailService     interfaces.EmailService
}

func NewAnalysisController(
	analysisService interfaces.AnalysisService,
	curriculumRepo interfaces.CurriculumRepositoryInterface,
	jobRepo interfaces.JobRepositoryInterface,
	notificationRepo interfaces.NotificationRepositoryInterface,
	planRepo interfaces.PlanRepositoryInterface,
	emailService interfaces.EmailService,
) *AnalysisController {
	return &AnalysisController{
		analysisService:  analysisService,
		curriculumRepo:   curriculumRepo,
		jobRepo:          jobRepo,
		notificationRepo: notificationRepo,
		planRepo:         planRepo,
		emailService:     emailService,
	}
}

type analyzeJobRequest struct {
	JobID        int `json:"job_id" binding:"required"`
	CurriculumID int `json:"curriculum_id"`
}

// AnalyzeJob godoc
// @Summary Analisar vaga com IA
// @Description Analisa compatibilidade do curriculo com uma vaga usando IA
// @Tags Analysis
// @Accept json
// @Produce json
// @Param body body model.AnalyzeJobRequest true "ID da vaga"
// @Success 200 {object} model.ResumeAnalysis
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/analyze-job [post]
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

	// Buscar currículo do usuário
	curricula, err := ac.curriculumRepo.FindCurriculumByUserID(user.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar currículo"})
		return
	}

	if len(curricula) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Nenhum currículo encontrado. Crie um currículo antes de analisar vagas."})
		return
	}

	var selectedCurriculum *model.Curriculum
	if body.CurriculumID > 0 {
		for i := range curricula {
			if curricula[i].Id == body.CurriculumID {
				selectedCurriculum = &curricula[i]
				break
			}
		}
		if selectedCurriculum == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Currículo não encontrado"})
			return
		}
	} else {
		selectedCurriculum = &curricula[0]
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
	analysis, err := ac.analysisService.Analyze(ctx.Request.Context(), *selectedCurriculum, *job)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao executar análise de IA"})
		return
	}

	// Registrar notificação com resultado da análise
	analysisJSON, _ := json.Marshal(analysis)
	if err := ac.notificationRepo.InsertNotificationWithAnalysis(job.ID, user.Id, selectedCurriculum.Id, analysisJSON); err != nil {
		logging.Logger.Error().Err(err).Int("job_id", job.ID).Int("user_id", user.Id).Msg("Erro ao registrar análise")
	}

	ctx.JSON(http.StatusOK, analysis)
}

type sendAnalysisEmailRequest struct {
	JobID    int                  `json:"job_id" binding:"required"`
	Analysis model.ResumeAnalysis `json:"analysis" binding:"required"`
}

// SendAnalysisEmail godoc
// @Summary Enviar analise por email
// @Description Envia resultado da analise de compatibilidade por email
// @Tags Analysis
// @Accept json
// @Produce json
// @Param body body model.SendAnalysisEmailRequest true "ID da vaga e resultado da analise"
// @Success 200 {object} model.MessageResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security CookieAuth
// @Router /api/analyze-job/send-email [post]
func (ac *AnalysisController) SendAnalysisEmail(ctx *gin.Context) {
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

	var body sendAnalysisEmailRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "job_id e analysis são obrigatórios"})
		return
	}

	// Buscar job para dados do email
	job, err := ac.jobRepo.GetJobByID(body.JobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar vaga"})
		return
	}
	if job == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Vaga não encontrada"})
		return
	}

	err = ac.emailService.SendAnalysisEmail(ctx.Request.Context(), user.Email, *job, body.Analysis)
	if err != nil {
		logging.Logger.Error().Err(err).Int("job_id", body.JobID).Int("user_id", user.Id).Msg("Erro ao enviar email de análise")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar email"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Email enviado com sucesso"})
}

func (ac *AnalysisController) GetAnalysisHistory(ctx *gin.Context) {
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}
	user, ok := userInterface.(model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Tipo de usuário inválido"})
		return
	}

	jobIDStr := ctx.Query("job_id")
	if jobIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "job_id é obrigatório"})
		return
	}
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "job_id inválido"})
		return
	}

	result, cvID, err := ac.notificationRepo.GetAnalysisHistory(user.Id, jobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar histórico"})
		return
	}
	if result == nil {
		ctx.JSON(http.StatusOK, gin.H{"has_analysis": false})
		return
	}

	var analysis model.ResumeAnalysis
	if err := json.Unmarshal(result, &analysis); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar análise"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"has_analysis":  true,
		"analysis":      analysis,
		"curriculum_id": cvID,
	})
}
