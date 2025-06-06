package controller

import (
	"net/http"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)

type jobController struct {
	JobUseCase usecase.JobUseCase
}

func NewJobController(jobUseCase usecase.JobUseCase) jobController {
	return jobController{
		JobUseCase: jobUseCase,
	}
}

func (controller *jobController) ScrappeAndInsert(ctx *gin.Context){

	jobs ,err := controller.JobUseCase.ScrapeAndStoreJobs(ctx)
	if(err != nil){
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusCreated, jobs)
}