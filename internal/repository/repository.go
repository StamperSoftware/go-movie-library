package repository

import (
	"database/sql"
	"movie-library/internal/models"
)

type DatabaseRepo interface {
	AllMovies() ([]*models.Movie, error)
	Connection() *sql.DB
	GetUserByEmail(email string) (*models.User, error)
}
