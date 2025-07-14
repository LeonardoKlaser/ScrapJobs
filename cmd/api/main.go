// scrapper go
package main

import (
	"log"
	"os"
	"web-scrapper/controller"
	"web-scrapper/infra/db"
	"web-scrapper/repository"
	"web-scrapper/usecase"
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

	// Usecases
	userUsecase := usecase.NewUserUsercase(userRepository)
	curriculumUsecase := usecase.NewCurriculumUsecase(curriculumRepository)

	// Controllers
	userController := controller.NewUserController(userUsecase)
	curriculumController := controller.NewCurriculumController(curriculumUsecase)

	server.GET("/curriculum/:id", curriculumController.GetCurriculumByUserId)
	server.POST("/curriculum", curriculumController.CreateCurriculum)
	server.POST("/user", userController.CreateUser)
	server.GET("/user/:email", userController.GetUserByEmail)



	server.Run(":8080")
}