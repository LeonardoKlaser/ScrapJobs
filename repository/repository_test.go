package repository

import (
	"os"
	"testing"
	"database/sql"
	"web-scrapper/model"
)

var testDB *sql.DB

func TestMain(m *testing.M){
	var closeFunc func()
	testDB, closeFunc = OpenConnection()

	defer closeFunc()
	os.Exit(m.Run())
}


func TestNotification(t *testing.T){
	user := model.User{
		Name: "Leonardo",
		Email: "leobkklaser@gmail.com",
		Password: "leo310504",
	}

	userRepository := NewUserRepository(testDB)

	if userRepository.DB == nil {
		t.Fatal("userRepository.DB should not be nil")
	}

	newUser, err := userRepository.CreateUser(user)
	if err != nil {
		t.Fatal("error to insert default user")
	}

	if newUser.Id == 0 {
		t.Fatal("error to insert default user")
	}

	job := model.Job{
		Title: "Go Developer",
		Location: "São Leopoldo",
		Company: "SAP",
		Job_link: "https://www.sap.com.br",
		Requisition_ID: 50,
		Description: "Precisa ter experiencia em golang tmj",
	}

	
}