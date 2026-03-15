package controller

import (
	"errors"
	"net/http"
	"strconv"

	"web-scrapper/interfaces"
	"web-scrapper/model"
	"web-scrapper/repository"

	"github.com/gin-gonic/gin"
)

type JobApplicationController struct {
	repo interfaces.JobApplicationRepositoryInterface
}

func NewJobApplicationController(repo interfaces.JobApplicationRepositoryInterface) *JobApplicationController {
	return &JobApplicationController{repo: repo}
}

func (c *JobApplicationController) extractUser(ctx *gin.Context) (model.User, bool) {
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		ctx.Abort()
		return model.User{}, false
	}
	user, ok := userInterface.(model.User)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Tipo de usuário inválido no contexto"})
		ctx.Abort()
		return model.User{}, false
	}
	return user, true
}

var validStatuses = map[string]bool{
	"applied": true, "in_review": true, "technical_test": true,
	"interview": true, "offer": true, "hired": true,
	"rejected": true, "withdrawn": true,
}

func (c *JobApplicationController) Create(ctx *gin.Context) {
	user, ok := c.extractUser(ctx)
	if !ok {
		return
	}

	var req model.CreateApplicationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "job_id é obrigatório"})
		return
	}

	jobExists, err := c.repo.JobExists(req.JobID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}
	if !jobExists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Vaga não encontrada"})
		return
	}

	app, err := c.repo.Create(user.Id, req.JobID)
	if err != nil {
		if errors.Is(err, repository.ErrApplicationExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Candidatura já existe para esta vaga"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		}
		return
	}

	ctx.JSON(http.StatusCreated, app)
}

func (c *JobApplicationController) Update(ctx *gin.Context) {
	user, ok := c.extractUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var req model.UpdateApplicationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Validate status
	if req.Status != nil {
		if !validStatuses[*req.Status] {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Status inválido"})
			return
		}
		if *req.Status == "interview" {
			if req.InterviewRound == nil || *req.InterviewRound < 1 {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Número da entrevista é obrigatório"})
				return
			}
		} else {
			req.InterviewRound = nil
		}
	}

	app, err := c.repo.Update(id, user.Id, req)
	if err != nil {
		if errors.Is(err, repository.ErrApplicationNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Candidatura não encontrada"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		}
		return
	}

	ctx.JSON(http.StatusOK, app)
}

func (c *JobApplicationController) Delete(ctx *gin.Context) {
	user, ok := c.extractUser(ctx)
	if !ok {
		return
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := c.repo.Delete(id, user.Id); err != nil {
		if errors.Is(err, repository.ErrApplicationNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Candidatura não encontrada"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *JobApplicationController) GetAll(ctx *gin.Context) {
	user, ok := c.extractUser(ctx)
	if !ok {
		return
	}

	apps, err := c.repo.GetAllByUser(user.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, model.ApplicationsResponse{Applications: apps})
}
