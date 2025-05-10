package pgconnect

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

func pullConnString() string {

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error reading .env file")
		return ""
	}
	// Get the connection string from the environment variable
	host := os.Getenv("DOCKER_IP")
	port := os.Getenv("PG_PORT")
	user := os.Getenv("POST_USER")
	password := os.Getenv("POST_PASSWORD")
	dbname := os.Getenv("DBNAME")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
}

func TestConnection() error {
	connStr := pullConnString()
	if connStr == "" {
		return fmt.Errorf("failed to get connection string")
	}

	fmt.Println("Connection string: ", connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open connection: %v", err)
	}
	defer db.Close()

	return db.Ping()
}

func ListenForPG() {
	app := echo.New()
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())
	app.Use(middleware.CORS())
	app.Use(middleware.Gzip())
	app.Use(middleware.Secure())
}
