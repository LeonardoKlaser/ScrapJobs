// scrapper go
package main

import (
	"log"
	"os"
	"web-scrapper/controller"
	"web-scrapper/infra/db"
	"web-scrapper/repository"
	"web-scrapper/usecase"
	"web-scrapper/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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
	SiteCareerUsecase := usecase.NewSiteCareerUsecase(&siteCareerRepository)

	// Controllers
	userController := controller.NewUserController(userUsecase)
	curriculumController := controller.NewCurriculumController(curriculumUsecase)
	userSiteController := controller.NewUserSiteController(UserSiteUsecase)
	siteCareerController := controller.NewSiteCareerController(SiteCareerUsecase)

	//middleware
	middlewareAuth := middleware.NewMiddleware(&userUsecase)

	//POST
	server.POST("/curriculum", middlewareAuth.RequireAuth ,curriculumController.CreateCurriculum)
	server.POST("/register", userController.SignUp)
	server.POST("/login", userController.SignIn)
	server.POST("/userSite", middlewareAuth.RequireAuth, userSiteController.InsertUserSite)
	server.POST("/siteCareer", middlewareAuth.RequireAuth ,siteCareerController.InsertNewSiteCareer)

	//GET
	server.GET("/curriculum/:id", middlewareAuth.RequireAuth ,curriculumController.GetCurriculumByUserId)


	server.Run(":8080")
}