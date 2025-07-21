package repository

import (
	"database/sql"
	"log"
	"web-scrapper/infra/db"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func OpenConnection() (*sql.DB, func()){
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
		return nil, nil
	}

	dbUser := "user_test"
	dbPass := "pass_test"
	dbName :=  "db_test"

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres", 
		Tag:        "13-alpine",    
		Env: []string{
			"POSTGRES_USER=" + dbUser,
			"POSTGRES_PASSWORD=" + dbPass,
			"POSTGRES_DB=" + dbName,
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		log.Fatalf("Could not creat postgre container: %s", err)
		return nil, nil
	}

	var dbConnection *sql.DB

	err = pool.Retry(func() error {
		
		port := resource.GetPort("5432/tcp")
		
       
		dbConnection, err = db.ConnectDB("localhost", port, dbUser, dbPass, dbName)
		if err != nil {
			log.Println("Database not ready yet, retrying...")
			return err
		}
		
        
		return dbConnection.Ping()
	})

	if err != nil {
		if purgeErr := pool.Purge(resource); purgeErr != nil {
			log.Fatalf("Could not purge resource: %s", purgeErr)
		}
		log.Fatalf("Could not connect to database: %s", err)
	}

	closeFunc := func() {
		if err := dbConnection.Close(); err != nil {
			log.Printf("Could not close database connection: %s", err)
		}
		
		if err := pool.Purge(resource); err != nil {
			log.Printf("Could not purge resource: %s", err)
		}
	}

	return dbConnection, closeFunc
}