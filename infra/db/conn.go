package db

import(
	"database/sql"
	"fmt"
	"strconv"
	_ "github.com/lib/pq"
)


func ConnectDB(host string, port string, user string ,password string, dbname string) (*sql.DB, error){
	portNumber, err := strconv.Atoi(port)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
		return nil, err
	}
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, portNumber, user, password, dbname)
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
