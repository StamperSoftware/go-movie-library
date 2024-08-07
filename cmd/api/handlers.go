package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	out, err := json.MarshalIndent(payload, "", "\t")

	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
}

func (app *application) Movies(w http.ResponseWriter, r *http.Request) {
	movies, err := app.DB.AllMovies()

	if err != nil {
		fmt.Println(err)
		return
	}

	out, err := json.MarshalIndent(movies, "", "\t")

	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
}
