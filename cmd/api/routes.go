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
	mux.Get("/api/refresh", app.RefreshToken)
	mux.Post("/api/authenticate", app.Authenticate)
	mux.Get("/api/logout", app.Logout)

	mux.Route("/api/admin", func(adminMux chi.Router) {
		adminMux.Use(app.authRequired)
		adminMux.Get("/movies", app.MovieCatalog)
	})
	return mux
}
