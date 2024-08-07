package dbrepo

import (
	"context"
	"database/sql"
	"log"
	"movie-library/internal/models"
	"time"
)

type PostgresDBRepo struct {
	DB *sql.DB
}

const dbTimeout = time.Second * 3

func (m *PostgresDBRepo) AllMovies() ([]*models.Movie, error) {

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var movies []*models.Movie
	query := `select id, title, release_date, runtime, mpaa_rating, description, coalesce(image, ''), created_at, updated_at from movies order by title`
	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var movie models.Movie
		err := rows.Scan(&movie.ID, &movie.Name, &movie.ReleaseDate, &movie.RunTime, &movie.MPAARating, &movie.Description, &movie.Image, &movie.CreatedAt, &movie.UpdatedAt)
		if err != nil {
			return nil, err
		}

		movies = append(movies, &movie)
	}

	if err != nil {
		return movies, err
	}

	return movies, nil
}

func (m *PostgresDBRepo) Connection() *sql.DB {
	return m.DB
}
