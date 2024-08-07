package main

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"movie-library/internal/repository"
	"movie-library/internal/repository/dbrepo"
	"net/http"
	"os"
)

const port = 8080

type application struct {
	Domain string
	DSN    string
	DB     repository.DatabaseRepo
}

func main() {
	//set app config
	var app application

	//read from command line
	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5432 user=postgres password=postgres dbname=movies sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection string")
	flag.Parse()
	err := godotenv.Load()

	if err != nil {
		log.Fatal("could not get .env", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	app.DSN = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5", dbHost, dbPort, dbUser, dbPassword, dbName)

	//connect to db
	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	app.DB = &dbrepo.PostgresDBRepo{DB: conn}
	defer app.DB.Connection().Close()

	//start web server
	app.Domain = "example.com"
	log.Println("starting application on port:", port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())

	if err != nil {
		log.Fatal(err)
	}

}
