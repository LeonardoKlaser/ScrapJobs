// scrapper go
package main

import (
	"context"
	"web-scrapper/controller"
	"web-scrapper/infra/db"
	"web-scrapper/infra/ses"
	"web-scrapper/repository"
	"web-scrapper/usecase"

	"github.com/gin-gonic/gin"
)



func main() {
	server := gin.Default()

	dbConnection, err := db.ConnectDB()
	if(err != nil){
		panic((err))
	}
	println("Connected to the database!")

	mailSender, err := ses.NewSESMailSender(context.Background(), "leobkklaser@gmail.com")
	if err != nil {
		println("Erro ao criar o SESMailSender:", err)
	}
	JobRepository := repository.NewJobRepository(dbConnection)
	UserUseCase := usecase.NewJobUseCase(JobRepository, mailSender)
	JobController := controller.NewJobController(UserUseCase)

	server.GET("/scrape", JobController.ScrappeAndInsert)	

	server.Run(":8080")
}