package repository

import (
	"database/sql"
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

func (r *PostgresRepo) GetThreads(forumSlug string, limit int64, since string, desc bool) ([]structs.Thread, error) {
	threads := make([]structs.Thread, 0)
	query := `SELECT author, created, id, message, slug, title, votes FROM Thread WHERE forum=$1 AND created>=$2 ORDER BY created LIMIT $3`
	queryDesc := `SELECT author, created, id, message, slug, title, votes FROM Thread WHERE forum=$1 AND created>=$2 ORDER BY created DESC LIMIT $3`

	var rows *sql.Rows
	var err error
	if desc {
		rows, err = r.DB.Query(queryDesc, forumSlug, since, limit)
	}else{
		rows, err = r.DB.Query(query, forumSlug, since, limit)
	}
	if err != nil {
		return threads, err
	}

	for rows.Next(){
		thread := structs.Thread{}
		var created time.Time
		err := rows.Scan(&thread.Author, &created, &thread.Id, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
		thread.Created = created.Format(time.RFC3339)
		if err != nil {
			return threads, err
		}
		threads = append(threads, thread)
	}
	return threads, nil

}
