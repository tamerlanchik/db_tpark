package repository

import (
	"context"
	//"database/sql"
	"db_tpark/structs"
	//"fmt"
	pg "github.com/jackc/pgconn"
)

// вызывается мало, работает быстро
func (r *PostgresRepo) CreateForum(slug, title, user string) error {
	query := `INSERT INTO Forum (slug, title, usernick) VALUES ($1, $2, $3);`

	_, err := r.DB.Exec(context.Background(), query, slug, title, user)

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

	if err == nil {
		_, err = r.DB.Exec(context.Background(), `INSERT INTO ForumPosts(forum, posts) VALUES ($1, 0)`, slug)
	}
	return err
}

// работает быстро
func (r *PostgresRepo) GetForum(slug string) (structs.Forum, error) {
	query := `SELECT (SELECT ForumPosts.posts FROM ForumPosts WHERE ForumPosts.forum=slug), threads, title, usernick, slug FROM Forum WHERE lower(slug)=lower($1);`

	var forum structs.Forum
	err := r.DB.QueryRow(context.Background(), query, slug).Scan(&forum.Posts, &forum.Threads, &forum.Title, &forum.User, &forum.Slug)
	return forum, err
}


