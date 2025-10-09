package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"web-scrapper/model"
	"web-scrapper/repository"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)

type SiteCareerController struct{
	usecase *usecase.SiteCareerUsecase
	userSiteRepository *repository.UserSiteRepository
}

func NewSiteCareerController(usecase *usecase.SiteCareerUsecase, userSiteRepository *repository.UserSiteRepository) *SiteCareerController{
	return &SiteCareerController{
		usecase: usecase,
		userSiteRepository: userSiteRepository,
	}
}

func (usecase *SiteCareerController) GetAllSites(ctx *gin.Context){
	type siteDTO struct{
		SiteName string
		BaseURL string
		SiteId int
		LogoURL *string
		IsSubscribed bool
	}

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

	sites, err := usecase.usecase.GetAllSites();
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : "Erro ao buscar sites: " + err.Error()})
	}

	userSites, err := usecase.userSiteRepository.GetSubscribedSiteIDs(user.Id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : "Erro ao buscar sites por usuario: " + err.Error()})
	}

	var response []siteDTO
	for _, site := range sites{
		var newResponse siteDTO
		newResponse.BaseURL = site.BaseURL
		newResponse.SiteId = site.ID
		newResponse.SiteName = site.SiteName
		newResponse.LogoURL = site.LogoURL

		if _, ok := userSites[site.ID]; ok {
			newResponse.IsSubscribed = true
		}

		response = append(response, newResponse)
	}


	ctx.JSON(http.StatusOK, response)
}

func (usecase *SiteCareerController) InsertNewSiteCareer(ctx *gin.Context){
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

	if user.Email != "adminScrapjobs@gmail.com"{
		ctx.JSON(http.StatusBadRequest, gin.H{"error" : "only admins can add new sites"})
		return
	}

	err := ctx.Request.ParseMultipartForm(10 << 20) 
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar o formulário"})
        return
    }

    file, err := ctx.FormFile("logo")
    if err != nil && err != http.ErrMissingFile {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar o arquivo de logo"})
        return
    }

	siteJSON := ctx.Request.FormValue("siteData")
	var body model.SiteScrapingConfig
	if err := json.Unmarshal([]byte(siteJSON), &body); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Dados do site em formato JSON inválido"})
        return
    }

	if body.APIHeadersJSON != nil {
		var unescapedHeaders string
		if json.Unmarshal([]byte(*body.APIHeadersJSON), &unescapedHeaders) == nil {
			*body.APIHeadersJSON = unescapedHeaders
		}
	}

	if body.JSONDataMappings != nil {
		var unescapedMappings string
		if json.Unmarshal([]byte(*body.JSONDataMappings), &unescapedMappings) == nil {
			*body.JSONDataMappings = unescapedMappings
		}
	}

	log.Print("Site: "+ siteJSON)
	log.Print("APIHeaderJson: " + *body.APIHeadersJSON)
	log.Print("JSonDataMapping: " + *body.JSONDataMappings)
	log.Print(file)

	res, err := usecase.usecase.InsertNewSiteCareer(ctx, body, file)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error" : fmt.Errorf("ERROR to insert new site career:  %w", err).Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, res)
}

func (usecase *SiteCareerController) SandboxScrape(ctx *gin.Context){
	var config model.SiteScrapingConfig
	if err := ctx.ShouldBindJSON(&config); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Corpo da requisição inválido"})
		return
	}

	scrapedJobs, err := usecase.usecase.SandboxScrape(ctx, config)
	if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   err.Error(),
            "message": "Falha ao executar o scraping com a configuração fornecida.",
        })
        return
    }

	for _, job := range scrapedJobs{
		log.Printf("requisition ID: %v", job.Requisition_ID);
	}

	var ids = takeIDs(scrapedJobs);	
	for _, id := range ids{
		log.Printf("ids buscados: %v ", id)
	}


	if len(scrapedJobs) == 0 {
        ctx.JSON(http.StatusOK, gin.H{
            "success": true,
            "message": "A configuração funcionou, mas nenhuma vaga foi encontrada na primeira página.",
            "data":    []model.Job{},
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": fmt.Sprintf("%d vagas encontradas com sucesso.", len(scrapedJobs)),
        "data":    scrapedJobs,
    })
}

func takeIDs(jobs []*model.Job) []int64{
	var ids []int64
	for _, job := range(jobs){
		ids = append(ids, job.Requisition_ID)
	}
	return ids
}