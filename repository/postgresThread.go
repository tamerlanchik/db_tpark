package repository

import (
	"db_tpark/structs"
	"fmt"
	pg "github.com/jackc/pgconn"
	"time"
)

func (r *PostgresRepo) CreateThread(thread structs.Thread) (structs.Thread, error) {
	query := `INSERT INTO Thread (author,forum,message,created,title) VALUES ($1, $2, $3, $4, $5) RETURNING id, slug, votes`

	err := r.DB.QueryRow(query, thread.Author, thread.Forum, thread.Message, thread.Created, thread.Title).
			Scan(&thread.Id, &thread.Slug, &thread.Votes)
	if err != nil {
		if e, ok := err.(*pg.PgError); ok {
			switch e.Code{
			case "23503":
				if e.ConstraintName == "thread_author_fkey" {
					err = structs.InternalError{E: structs.ErrorNoUser}
				} else {
					err = structs.InternalError{E: structs.ErrorNoForum}
				}
				break
			case "23505":
				err = structs.InternalError{E: structs.ErrorDuplicateKey}
				break
			default:
				fmt.Println(e.Code)
			}
		}
	}


	return thread, err
}

func (r *PostgresRepo) GetThread(slug string) (structs.Thread, error) {
	var thread structs.Thread
	query := `SELECT author, created, forum, id, message, slug, title, votes FROM Thread WHERE slug=$1`

	var created time.Time
	err := r.DB.QueryRow(query, slug).
			Scan(&thread.Author, &created, thread.Forum,
			thread.Id, thread.Message, thread.Slug, thread.Title, thread.Votes)
	thread.Created = created.Format(time.RFC3339)
	return thread, err
}
