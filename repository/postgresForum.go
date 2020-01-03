package repository

import (
	//"database/sql"
	"db_tpark/structs"
	//"fmt"
	pg "github.com/jackc/pgconn"
)

// вызывается мало, работает быстро
func (r *PostgresRepo) CreateForum(slug, title, user string) error {
	query := `INSERT INTO Forum (slug, title, usernick) VALUES ($1, $2, $3);`

	_, err := r.DB.Exec(query, slug, title, user)

	if e, ok := err.(*pg.PgError); ok {
		switch e.Code {
		case "23503", "23502":
			err = structs.InternalError{E: structs.ErrorNoUser}
			break
		case "23505":
			err = structs.InternalError{E: structs.ErrorDuplicateKey}
			break
		default:
			//buildmode.Log.Println(e.Code)
		}
	}
	return err
}

// работает быстро
func (r *PostgresRepo) GetForum(slug string) (structs.Forum, error) {
	query := `SELECT posts, threads, title, usernick, slug FROM Forum WHERE lower(slug)=lower($1);`

	var forum structs.Forum
	err := r.DB.QueryRow(query, slug).Scan(&forum.Posts, &forum.Threads, &forum.Title, &forum.User, &forum.Slug)
	return forum, err
}


