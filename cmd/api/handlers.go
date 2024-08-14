package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"io"
	"log"
	"movie-library/internal/models"
	"net/http"
	"net/url"
	"strconv"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	var payload = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Movie Library Home",
		Version: "1.0.0",
	}
	app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) Genres(w http.ResponseWriter, r *http.Request) {
	genres, err := app.DB.Genres()

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, genres)
}

func (app *application) Movies(w http.ResponseWriter, r *http.Request) {
	genreID, err := strconv.Atoi(r.URL.Query().Get("genre"))
	var movies []*models.Movie
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if genreID != 0 {
		movies, err = app.DB.AllMovies(genreID)
	} else {
		movies, err = app.DB.AllMovies()
	}

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, movies)
}

func (app *application) Movie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	movie, err := app.DB.GetMovieByID(id)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, movie)
}

func (app *application) PostCreateMovie(w http.ResponseWriter, r *http.Request) {
	var movie models.Movie
	err := app.readJSON(w, r, &movie)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	movie = app.GetMoviePoster(movie)
	newId, err := app.DB.CreateMovie(movie)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.DB.CreateMovieGenre(newId, movie.GenresArray)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	resp := JSONResponse{
		Error:   false,
		Message: "movie updated",
	}

	app.writeJSON(w, http.StatusAccepted, resp)
}

func (app *application) PutUpdateMovie(w http.ResponseWriter, r *http.Request) {
	var movie models.Movie
	err := app.readJSON(w, r, &movie)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.DB.UpdateMovie(movie)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.DB.CreateMovieGenre(movie.ID, movie.GenresArray)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	resp := JSONResponse{
		Error:   false,
		Message: "movie updated",
	}

	app.writeJSON(w, http.StatusAccepted, resp)
}

func (app *application) CreateMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	movie, err := app.DB.GetMovieByID(id)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := struct {
		Movies *models.Movie `json:"movies"`
	}{movie}

	app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.DB.DeleteMovie(id)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}{false, "Movie Deleted"}

	app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) Authenticate(w http.ResponseWriter, r *http.Request) {

	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	user, err := app.DB.GetUserByEmail(requestPayload.Email)
	if err != nil {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	valid, err := user.DoesPasswordMatch(requestPayload.Password)
	if err != nil || !valid {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	jwtUser := jwtUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	tokens, err := app.auth.GenerateTokenPair(&jwtUser)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)
	http.SetCookie(w, refreshCookie)
	app.writeJSON(w, http.StatusAccepted, tokens)
}

func (app *application) RefreshToken(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == app.auth.CookieName {
			claims := &Claims{}
			refreshToken := cookie.Value

			_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(app.JWTSecret), nil
			})

			if err != nil {
				app.errorJSON(w, errors.New("not authorized"), http.StatusUnauthorized)
				return
			}

			userID, err := strconv.Atoi(claims.Subject)

			if err != nil {
				app.errorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			user, err := app.DB.GetUserByID(userID)
			if err != nil {
				app.errorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			u := jwtUser{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			}

			tokenPairs, err := app.auth.GenerateTokenPair(&u)

			if err != nil {
				app.errorJSON(w, errors.New("error generating token"), http.StatusUnauthorized)
				return
			}
			http.SetCookie(w, app.auth.GetRefreshCookie(tokenPairs.RefreshToken))
			app.writeJSON(w, http.StatusOK, tokenPairs)
		}
	}
}

func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, app.auth.GetExpiredRefreshCookie())
	w.WriteHeader(http.StatusAccepted)

}

func (app *application) MovieCatalog(w http.ResponseWriter, r *http.Request) {
	movies, err := app.DB.AllMovies()

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, movies)
}

func (app *application) GetMoviePoster(movie models.Movie) models.Movie {
	type TheMovieDB struct {
		Page    int `json:"page"`
		Results []struct {
			PosterPath string `json:"poster_path"`
		} `json:"results"`
		TotalPages int `json:"total_pages"`
	}

	client := &http.Client{}
	movieDBUrl := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s", app.MovieDBAPIKey)

	req, err := http.NewRequest("GET", movieDBUrl+"&query="+url.QueryEscape(movie.Title), nil)

	if err != nil {
		log.Println(err)
		return movie
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return movie
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println(err)
		return movie
	}

	var responseObject TheMovieDB
	json.Unmarshal(bodyBytes, &responseObject)

	if len(responseObject.Results) > 0 {
		movie.Image = responseObject.Results[0].PosterPath
	}

	return movie
}

func (app *application) GetMoviesByGenre(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("genre"))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	movies, err := app.DB.AllMovies(id)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, movies)
}
