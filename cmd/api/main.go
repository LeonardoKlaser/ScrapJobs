// scrapper go
package main

import (
	"context"
	"os"
	"time"
	"web-scrapper/controller"
	"web-scrapper/infra/db"
	"web-scrapper/infra/s3"
	"web-scrapper/logging"
	"web-scrapper/middleware"
	"web-scrapper/model"
	"web-scrapper/repository"
	"web-scrapper/usecase"
	"web-scrapper/utils"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)



func main() {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173", "https://scrapjobs.com.br"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
        ExposeHeaders:    []string{"Content-Length"},
        MaxAge: 12 * time.Hour,
	}))

	if os.Getenv("GIN_MODE") != "release"{
		godotenv.Load()	
	}

	var err error
	var secrets *model.AppSecrets

	secretName := os.Getenv("APP_SECRET_NAME")
	if secretName  != ""{
		secrets, err = utils.GetAppSecrets(secretName)
		if err != nil {
			logging.Logger.Fatal().Err(err).Msg("Could not connect to database")
        }
	} else {
        secrets = &model.AppSecrets{
            DBHost:     os.Getenv("HOST_DB"),
            DBPort:     os.Getenv("PORT_DB"),
            DBUser: os.Getenv("USER_DB"),
            DBPassword: os.Getenv("PASSWORD_DB"),
            DBName:   os.Getenv("DBNAME"),
			RedisAddr: os.Getenv("REDIS_ADDR"),
        }
    }

	dbConnection, err := db.ConnectDB(secrets.DBHost, secrets.DBPort,secrets.DBUser,secrets.DBPassword,secrets.DBName)
	if(err != nil){
		logging.Logger.Fatal().Err(err).Msg("Could not connect to database")
	}
	
	logging.Logger.Info().Msg("successfully connected to the databse")
	
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: os.Getenv("REDIS_ADDR")})
    defer asynqClient.Close()

	awsCfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("failed to load AWS configuration")
	}
	
	s3BucketName := os.Getenv("S3_BUCKET_NAME")
	if s3BucketName == "" {
		logging.Logger.Fatal().Msg("S3_BUCKET_NAME environment variable not set")
	}

	s3Uploader := s3.NewUploader(awsCfg, s3BucketName)
	

	// Repositories
	userRepository := repository.NewUserRepository(dbConnection)
	curriculumRepository := repository.NewCurriculumRepository(dbConnection)
	userSiteRepository := repository.NewUserSiteRepository(dbConnection)
	siteCareerRepository := repository.NewSiteCareerRepository(dbConnection)
	dashboardRepository := repository.NewDashboardRepository(dbConnection)
	planRepository := repository.NewPlanRepository(dbConnection)
	requestedSiteRepository := repository.NewRequestedSiteRepository(dbConnection)

	// Usecases
	userUsecase := usecase.NewUserUsercase(userRepository)
	curriculumUsecase := usecase.NewCurriculumUsecase(curriculumRepository)
	UserSiteUsecase := usecase.NewUserSiteUsecase(userSiteRepository)
	SiteCareerUsecase := usecase.NewSiteCareerUsecase(siteCareerRepository, s3Uploader)
	planUsecase := usecase.NewPlanUsecase(planRepository)
	requestedSiteUsecase := usecase.NewRequestedSiteUsecase(requestedSiteRepository)

	// Controllers
	userController := controller.NewUserController(userUsecase)
	curriculumController := controller.NewCurriculumController(curriculumUsecase)
	userSiteController := controller.NewUserSiteController(UserSiteUsecase)
	siteCareerController := controller.NewSiteCareerController(SiteCareerUsecase, userSiteRepository)
	healthController := controller.NewHealthController(dbConnection, asynqClient)
	checkAuthController := controller.NewCheckAuthController()
	dashboardController := controller.NewDashboardDataController(dashboardRepository)
	planController := controller.NewPlanController(planUsecase)
	requestedSiteController := controller.NewRequestedSiteController(requestedSiteUsecase)

	//middleware
	middlewareAuth := middleware.NewMiddleware(userUsecase)

	//rate limiter 
	publicRateLimiter := middleware.RateLimiter(rate.Limit(5.0/60.0), 2)

	publicRoutes := server.Group("/")
	publicRoutes.Use(logging.GinMiddleware())
	publicRoutes.Use(publicRateLimiter)
	{
		publicRoutes.POST("/register", userController.SignUp)
		publicRoutes.POST("/login", userController.SignIn)
		publicRoutes.GET("/api/plans", planController.GetAllPlans)

	}

	privateRateLimiter := middleware.RateLimiter(rate.Limit(15.0/60.0),10)

	privateRoutes := server.Group("/")
	privateRoutes.Use(logging.GinMiddleware())
	privateRoutes.Use(middlewareAuth.RequireAuth)
	{
		privateRoutes.GET("api/me", checkAuthController.CheckAuthUser)
		privateRoutes.GET("api/dashboard", dashboardController.GetDashboardDataByUserId)
		privateRoutes.GET("api/getSites", siteCareerController.GetAllSites)
	}
	privateRoutes.Use(privateRateLimiter)
	{
		
		privateRoutes.POST("/curriculum", curriculumController.CreateCurriculum)
		privateRoutes.POST("/userSite", userSiteController.InsertUserSite)
		privateRoutes.POST("/siteCareer", siteCareerController.InsertNewSiteCareer)
		privateRoutes.POST("/scrape-sandbox", siteCareerController.SandboxScrape)
		privateRoutes.GET("/curriculum/:id", curriculumController.GetCurriculumByUserId)
		privateRoutes.POST("/api/logout", userController.Logout)
		privateRoutes.POST("api/request-site", requestedSiteController.Create)
	}

	 healthRoutes := server.Group("/health")
    {
        healthRoutes.GET("/live", healthController.Liveness)
        healthRoutes.GET("/ready", healthController.Readiness)
    }

	
	server.Run(":8080")
}