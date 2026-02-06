package sqlconnect

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDb() (*sql.DB, error) {

	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	params := "charset=utf8mb4&parseTime=True&loc=Local"

	encodedUserName := url.QueryEscape(username)
	encodedPassword := password //url.QueryEscape(password)

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", encodedUserName, encodedPassword, host, dbPort, dbName, params)
	fmt.Println("CONNECTION STRING :::: ", connectionString)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println("Connected to MariaDB")
	return db, nil
}
