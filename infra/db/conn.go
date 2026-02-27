package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

func applyPoolDefaults(db *sql.DB) {
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(3 * time.Minute)
}

func ConnectDB(host string, port string, user string, password string, dbname string, opts ...func(*sql.DB)) (*sql.DB, error) {
	portNumber, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("error converting port to int: %w", err)
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "require"
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, portNumber, user, password, dbname, sslmode)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	applyPoolDefaults(db)
	for _, opt := range opts {
		opt(db)
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// ConnectDBFromURL opens a Postgres connection from a connection string (e.g. DATABASE_URL).
func ConnectDBFromURL(dsn string, opts ...func(*sql.DB)) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	applyPoolDefaults(db)
	for _, opt := range opts {
		opt(db)
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
