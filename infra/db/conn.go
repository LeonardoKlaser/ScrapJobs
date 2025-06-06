package db

import(
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	//host = "go_scrapper_db"
	host = "localhost"
	port = 5432
	user = "postgres"
	password = "postgres"
	dbname = "web_scrapper"
)

func ConnectDB() (*sql.DB, error){
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected to the database!")
	return db, nil
}