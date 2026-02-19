package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

func ConnectDB(host string, port string, user string, password string, dbname string) (*sql.DB, error) {
	portNumber, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("error converting port to int: %w", err)
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, portNumber, user, password, dbname, sslmode)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to the database!")
	return db, nil
}
