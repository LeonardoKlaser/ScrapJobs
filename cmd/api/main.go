// scrapper go
package main

import (
	"log"
	"os"
	"web-scrapper/controller"
	"web-scrapper/infra/db"
	"web-scrapper/middleware"
	"web-scrapper/repository"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)



func main() {
	server := gin.Default()
	godotenv.Load()
	dbConnection, err := db.ConnectDB(os.Getenv("HOST_DB"), os.Getenv("PORT_DB"),os.Getenv("USER_DB"),os.Getenv("PASSWORD_DB"),os.Getenv("DBNAME"))
	if(err != nil){
		panic((err))
	}
	log.Print("Connected to the database!")
	
	

	// Repositories
	userRepository := repository.NewUserRepository(dbConnection)
	curriculumRepository := repository.NewCurriculumRepository(dbConnection)
	userSiteRepository := repository.NewUserSiteRepository(dbConnection)
	siteCareerRepository := repository.NewSiteCareerRepository(dbConnection)

	// Usecases
	userUsecase := usecase.NewUserUsercase(userRepository)
	curriculumUsecase := usecase.NewCurriculumUsecase(curriculumRepository)
	UserSiteUsecase := usecase.NewUserSiteUsecase(userSiteRepository)
	SiteCareerUsecase := usecase.NewSiteCareerUsecase(siteCareerRepository)

	// Controllers
	userController := controller.NewUserController(userUsecase)
	curriculumController := controller.NewCurriculumController(curriculumUsecase)
	userSiteController := controller.NewUserSiteController(UserSiteUsecase)
	siteCareerController := controller.NewSiteCareerController(SiteCareerUsecase)

	//middleware
	middlewareAuth := middleware.NewMiddleware(userUsecase)

	//rate limiter 
	publicRateLimiter := middleware.RateLimiter(rate.Limit(5.0/60.0), 2)

	publicRoutes := server.Group("/")
	publicRoutes.Use(publicRateLimiter)
	{
		publicRoutes.POST("/register", userController.SignUp)
		publicRoutes.POST("/login", userController.SignIn)

	}

	privateRateLimiter := middleware.RateLimiter(rate.Limit(15.0/60.0),10)

	privateRoutes := server.Group("/")
	privateRoutes.Use(middlewareAuth.RequireAuth)
	privateRoutes.Use(privateRateLimiter)
	{
		privateRoutes.POST("/curriculum", curriculumController.CreateCurriculum)
		privateRoutes.POST("/userSite", userSiteController.InsertUserSite)
		privateRoutes.POST("/siteCareer", siteCareerController.InsertNewSiteCareer)

		privateRoutes.GET("/curriculum/:id", curriculumController.GetCurriculumByUserId)
	}

	
	server.Run(":8080")
}