package repository

import (
	"database/sql"
	"movie-library/internal/models"
)

type DatabaseRepo interface {
	AllMovies() ([]*models.Movie, error)
	Genres() ([]*models.Genre, error)
	Connection() *sql.DB
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
	GetMovieByID(id int) (*models.Movie, error)
	DeleteMovie(id int) error
	UpdateMovie(movie models.Movie) error
	CreateMovie(movie models.Movie) (int, error)
	CreateMovieGenre(id int, genreIDs []int) error
}
