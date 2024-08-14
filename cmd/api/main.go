package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"movie-library/internal/repository"
	"movie-library/internal/repository/dbrepo"
	"net/http"
	"os"
	"time"
)

const port = 8080

type application struct {
	Domain        string
	DSN           string
	DB            repository.DatabaseRepo
	auth          Auth
	JWTIssuer     string
	JWTAudience   string
	CookieDomain  string
	JWTSecret     string
	MovieDBAPIKey string
}

func main() {
	//set app config
	var app application

	err := godotenv.Load()

	if err != nil {
		log.Fatal("could not get .env", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	app.JWTSecret = os.Getenv("JWT_SECRET")
	app.JWTIssuer = os.Getenv("JWT_ISSUER")
	app.JWTAudience = os.Getenv("JWT_AUDIENCE")
	app.CookieDomain = os.Getenv("COOKIE_DOMAIN")
	app.Domain = os.Getenv("DOMAIN")
	app.MovieDBAPIKey = os.Getenv("MOVIE_DB_API_KEY")

	app.DSN = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5", dbHost, dbPort, dbUser, dbPassword, dbName)

	//connect to db
	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}

	app.DB = &dbrepo.PostgresDBRepo{DB: conn}
	defer app.DB.Connection().Close()

	app.auth = Auth{
		Issuer:        app.JWTIssuer,
		Audience:      app.JWTAudience,
		Secret:        app.JWTSecret,
		TokenExpiry:   time.Minute * 15,
		RefreshExpiry: time.Hour * 24,
		CookieDomain:  app.CookieDomain,
		CookiePath:    "/",
		CookieName:    "refresh",
	}

	//start web server
	log.Println("starting application on port:", port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())

	if err != nil {
		log.Fatal(err)
	}

}
