package repository

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"
	"web-scrapper/model"
)

var testDB *sql.DB

func TestMain(m *testing.M){
	var closeFunc func()
	testDB, closeFunc = OpenConnection()
	defer closeFunc()

	migrationsDir := "../migrations"

	upFiles, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		log.Fatalf("Could not find migration files: %s", err)
	}

	for _, file := range upFiles {
		log.Printf("  -> Applying %s\n", filepath.Base(file))
		migrationSQL, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Could not read migration file %s: %s", file, err)
		}

		_, err = testDB.Exec(string(migrationSQL))
		if err != nil {
			log.Fatalf("Error applying migration %s: %s", file, err)
		}
	}
	
	os.Exit(m.Run())
}


func TestNotification(t *testing.T){
	user := model.User{
		Name: "Leonardo",
		Email: "leobkklaser@gmail.com",
		Password: "leo310504",
	}

	userRepository := NewUserRepository(testDB)

	if userRepository == nil {
		t.Fatal("userRepository.DB should not be nil")
	}

	_, err := userRepository.CreateUser(user)
	if err != nil {
		t.Fatalf("error to insert default user: %s", err)
	}

	job := model.Job{
		Title: "Go Developer",
		Location: "São Leopoldo",
		Company: "SAP",
		Job_link: "https://www.sap.com.br",
		Requisition_ID: 50,
		Description: "Precisa ter experiencia em golang tmj",
	}

	jobRepository := NewJobRepository(testDB)

	newJobID, err := jobRepository.CreateJob(job)

	if err != nil {
		t.Fatal("error to create default job")
	}

	if newJobID == 0 {
		t.Fatal("the job ID didnt return")
	}
}