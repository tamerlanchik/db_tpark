package repository

import (
	//"database/sql"
	"db_tpark/structs"
	"fmt"
	pg "github.com/jackc/pgconn"
)

func (r *PostgresRepo) CreateAndReturnForum(slug, title, user string) (structs.Forum, error) {
	query := `INSERT INTO Forum (slug, title, usernick) VALUES ($1, $2, $3)
				RETURNING posts, threads, title, slug, usernick;`

	var forum structs.Forum
	err := r.DB.QueryRow(query, slug, title, user).Scan(&forum.Posts, &forum.Threads, &forum.Title, &forum.Slug, &forum.User)
	return forum, err
}

func (r *PostgresRepo) CreateForum(slug, title, user string) error {
	query := `INSERT INTO Forum (slug, title, usernick) VALUES ($1, $2, $3);`

	_, err := r.DB.Exec(query, slug, title, user)

	if e, ok := err.(*pg.PgError); ok {
		switch e.Code {
		case "23503":
			err = structs.InternalError{E: structs.ErrorNoUser}
			break
		case "23505":
			err = structs.InternalError{E: structs.ErrorDuplicateKey}
			break
		default:
			fmt.Println(e.Code)
		}
	}
	return err
}

func (r *PostgresRepo) GetForum(slug string) (structs.Forum, error) {
	query := `SELECT posts, threads, title, usernick FROM Forum WHERE slug=$1;`

	var forum structs.Forum
	err := r.DB.QueryRow(query, slug).Scan(&forum.Posts, &forum.Threads, &forum.Title, &forum.User)
	forum.Slug = slug
	return forum, err
}
