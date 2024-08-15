package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Use(app.enableCORS)

	mux.Get("/", app.Home)
	mux.Get("/api/movies", app.Movies)
	mux.Get("/api/movies/{id}", app.Movie)
	mux.Get("/api/movies?genre={genre}", app.GetMoviesByGenre)
	mux.Get("/api/genres", app.Genres)
	mux.Post("/api/graph", app.GraphQL)

	mux.Get("/api/refresh", app.RefreshToken)
	mux.Post("/api/authenticate", app.Authenticate)
	mux.Get("/api/logout", app.Logout)

	mux.Route("/api/admin", func(adminMux chi.Router) {
		adminMux.Use(app.authRequired)
		adminMux.Get("/movies", app.MovieCatalog)
		adminMux.Post("/movies", app.PostCreateMovie)
		adminMux.Get("/movies/{id}", app.CreateMovie)
		adminMux.Put("/movies/{id}", app.PutUpdateMovie)
		adminMux.Delete("/movies/{id}", app.DeleteMovie)
	})
	return mux
}
