package dbrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"movie-library/internal/models"
	"time"
)

type PostgresDBRepo struct {
	DB *sql.DB
}

const dbTimeout = time.Second * 3

func (m *PostgresDBRepo) Genres() ([]*models.Genre, error) {

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var genres []*models.Genre
	query := `select id, genre, created_at, updated_at from genres order by genres`
	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var genre models.Genre
		err := rows.Scan(&genre.ID, &genre.Genre, &genre.CreatedAt, &genre.UpdatedAt)
		if err != nil {
			return nil, err
		}

		genres = append(genres, &genre)
	}

	if err != nil {
		return genres, err
	}

	return genres, nil
}

func (m *PostgresDBRepo) AllMovies(genres ...int) ([]*models.Movie, error) {

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var movies []*models.Movie
	where := ""
	if len(genres) > 0 {
		where = fmt.Sprintf("where id in (select movie_id from movies_genres where genre_id = %d)", genres[0])
	}
	query := fmt.Sprintf(`select id, title, release_date, runtime, mpaa_rating, description, coalesce(image, ''), created_at, updated_at from movies %s order by title`, where)
	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var movie models.Movie
		err := rows.Scan(&movie.ID, &movie.Title, &movie.ReleaseDate, &movie.RunTime, &movie.MPAARating, &movie.Description, &movie.Image, &movie.CreatedAt, &movie.UpdatedAt)
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

func (m *PostgresDBRepo) GetMovieByID(id int) (*models.Movie, error) {

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	var movie models.Movie
	query := `select id, title, release_date, runtime, mpaa_rating, description, coalesce(image, ''), created_at, updated_at from movies where id = $1`
	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(&movie.ID, &movie.Title, &movie.ReleaseDate, &movie.RunTime, &movie.MPAARating, &movie.Description, &movie.Image, &movie.CreatedAt, &movie.UpdatedAt)
	if err != nil {
		return nil, err
	}

	query = `select g.id, g.genre from movies_genres mg left join genres g on (mg.genre_id = g.id) where mg.movie_id = $1 order by g.genre`
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	defer rows.Close()
	var genres []*models.Genre
	var genreArray []int

	for rows.Next() {
		var g models.Genre
		err := rows.Scan(&g.ID, &g.Genre)
		g.Checked = true
		if err != nil {
			return nil, err
		}
		genres = append(genres, &g)
		genreArray = append(genreArray, g.ID)
	}

	movie.Genres = genres
	movie.GenresArray = genreArray
	return &movie, nil
}

func (m *PostgresDBRepo) DeleteMovie(id int) error {

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	query := `delete from movies where id = $1`
	_, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	return nil
}

func (m *PostgresDBRepo) CreateMovie(movie models.Movie) (int, error) {

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `insert into movies (title, description, release_date, runtime, mpaa_rating, image, created_at, updated_at) 
values ($1, $2, $3, $4, $5, $6, $7, $8) returning id`
	var newId int
	result := m.DB.QueryRowContext(ctx, query, movie.Title, movie.Description, movie.ReleaseDate, movie.RunTime, movie.MPAARating, movie.Image, time.Now(), time.Now())
	err := result.Scan(&newId)
	if err != nil {
		return 0, err
	}

	return newId, nil
}

func (m *PostgresDBRepo) UpdateMovie(movie models.Movie) error {

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `update movies set title=$1, description=$2, release_date=$3, runtime=$4, mpaa_rating=$5, image=$6, updated_at=$7 where id = $8`
	_, err := m.DB.ExecContext(ctx, query, movie.Title, movie.Description, movie.ReleaseDate, movie.RunTime, movie.MPAARating, movie.Image, time.Now(), movie.ID)

	if err != nil {
		return err
	}

	return nil
}

func (m *PostgresDBRepo) Connection() *sql.DB {
	return m.DB
}

func (m *PostgresDBRepo) GetUserByEmail(email string) (*models.User, error) {

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	query := `select id, first_name, last_name, email, password, created_at, updated_at from users where email = $1`
	var user models.User
	row := m.DB.QueryRowContext(ctx, query, email)
	err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *PostgresDBRepo) GetUserByID(id int) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, first_name, last_name, email, password, created_at, updated_at from users where id = $1`
	var user models.User
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *PostgresDBRepo) CreateMovieGenre(id int, genreIDs []int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `delete from movies_genres where movie_id = $1`

	_, _ = m.DB.ExecContext(ctx, query, id)

	for _, n := range genreIDs {
		query := `insert into movies_genres (movie_id, genre_id) values ($1, $2)`
		_, err := m.DB.ExecContext(ctx, query, id, n)

		if err != nil {
			return err
		}
	}

	return nil
}
