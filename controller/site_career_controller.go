package controller

import (
	"web-scrapper/model"
	"web-scrapper/usecase"
	"net/http"
	"fmt"
	"github.com/gin-gonic/gin"
)

type SiteCareerController struct{
	usecase *usecase.SiteCareerUsecase
}

func NewSiteCareerController(usecase *usecase.SiteCareerUsecase) *SiteCareerController{
	return &SiteCareerController{
		usecase: usecase,
	}
}

func (usecase *SiteCareerController) InsertNewSiteCareer(ctx *gin.Context){
	var body model.SiteScrapingConfig
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : fmt.Errorf("error to deserialize new job json body: %w", err)})
	}

	res, err := usecase.usecase.InsertNewSiteCareer(body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error" : fmt.Errorf("ERROR to insert new site career:  %w", err),
		})
		return
	}

	ctx.JSON(http.StatusCreated, res)
}