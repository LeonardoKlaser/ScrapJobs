// scrapper go
package main

import (
	"web-scrapper/db"
	"web-scrapper/repository"
	"web-scrapper/usecase"
	"web-scrapper/controller"

	"github.com/gin-gonic/gin"
)



func main() {
	server := gin.Default()

	dbConnection, err := db.ConnectDB()
	if(err != nil){
		panic((err))
	}
	println("Connected to the database!")

	JobRepository := repository.NewJobRepository(dbConnection)
	UserUseCase := usecase.NewJobUseCase(JobRepository)
	JobController := controller.NewJobController(UserUseCase)

	server.GET("/scrape", JobController.ScrappeAndInsert)	

	server.Run(":8080")
}