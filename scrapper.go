// scrapper go
package main

import (
	"context"
	"os"
	"web-scrapper/controller"
	"web-scrapper/infra/db"
	"web-scrapper/infra/gemini"
	"web-scrapper/infra/ses"
	"web-scrapper/repository"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)



func main() {
	server := gin.Default()
	godotenv.Load()
	dbConnection, err := db.ConnectDB()
	if(err != nil){
		panic((err))
	}
	println("Connected to the database!")

	geminiConfig := gemini.Config{
		ApiKey: os.Getenv("GEMINI_KEY"),
		ApiModel: os.Getenv("AI_MODEL"),
	}

	client, err := gemini.GeminiClientModel(context.Background(), geminiConfig)
	if err != nil {
		println("Erro ao criar o GeminiClient:", err)
		return
	}
	prompt := "Quais são as vagas de emprego disponíveis no Brasil?"
	response, err := client.GeminiSearch(context.Background(), prompt)
	if err != nil {
		println("Erro ao fazer a pesquisa no Gemini:", err)
		return
	}
	println("Resposta do Gemini:", response)
	
	

	mailSender, err := ses.NewSESMailSender(context.Background(), "leobkklaser@gmail.com")
	if err != nil {
		println("Erro ao criar o SESMailSender:", err)
	}
	JobRepository := repository.NewJobRepository(dbConnection)
	UserUseCase := usecase.NewJobUseCase(JobRepository, mailSender)
	JobController := controller.NewJobController(UserUseCase)

	CurriculumRepository := repository.NewCurriculumRepository(dbConnection)
	CurriculumUsecase := usecase.NewCurriculumUsecase(CurriculumRepository)
	CurriculumController := controller.NewCurriculumController(CurriculumUsecase)

	UserRepository := repository.NewUserRepository(dbConnection)
	UserUsecase := usecase.NewUserUsercase(UserRepository)
	UserController := controller.NewUserController(UserUsecase)


	server.GET("/scrape", JobController.ScrappeAndInsert)
	server.GET("/curriculum", CurriculumController.GetCurriculumByUserId)
	server.POST("/curriculum", CurriculumController.CreateCurriculum)
	server.POST("/user", UserController.CreateUser)
	server.GET("/user", UserController.GetUserByEmail)



	server.Run(":8080")
}