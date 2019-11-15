package repository

import (
	"database/sql"
	"db_tpark/structs"
	"fmt"
	pg "github.com/jackc/pgconn"
	"strconv"
	"strings"
	"time"
)

var counter int

func (r *PostgresRepo) CreateThread(thread structs.Thread) (structs.Thread, error) {
	query := `INSERT INTO Thread (author,forum,message,created,title, slug) VALUES ($1, $2, $3, %s::timestamptz, $5, %s) RETURNING id, slug, votes`
	//query = fmt.Sprintf(query, structs.NoEmptyWrapper("NOW()", 4))

	query = fmt.Sprintf(query, `COALESCE($4, NOW())`, `NULLIF($6, '')`)
	var slug sql.NullString
	t, _ := time.Parse(structs.OutTimeFormat, thread.Created)
	err := r.DB.QueryRow(query, thread.Author, thread.Forum, thread.Message, t, thread.Title, thread.Slug).
			Scan(&thread.Id, &slug, &thread.Votes)
	thread.Slug = slug.String
	if err != nil {
		if e, ok := err.(*pg.PgError); ok {
			switch e.Code{
			case "23503", "523502":
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
		Scan(&thread.Author, &created, &thread.Forum,
			&thread.Id, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	thread.Created = created.Format(time.RFC3339)
	return thread, err
}

func (r *PostgresRepo) GetThreadById(id int64) (structs.Thread, error) {
	var thread structs.Thread
	query := `SELECT author, created, forum, id, message, slug, title, votes FROM Thread WHERE id=$1`

	var created time.Time
	err := r.DB.QueryRow(query, id).
		Scan(&thread.Author, &created, &thread.Forum,
			&thread.Id, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	thread.Created = created.Format(structs.OutTimeFormat)
	return thread, err
}

func (r *PostgresRepo) GetThreads(forumSlug string, limit int64, since string, desc bool) ([]structs.Thread, error) {
	counter++
	threads := make([]structs.Thread, 0)
	query := `SELECT author, forum, created, id, message, slug, title, votes FROM Thread WHERE lower(forum)=lower($1) %s ORDER BY created %s %s`

	var rows *sql.Rows
	var err error
	params := make([]interface{}, 0,2)
	params = append(params, forumSlug)
	var placeholderSince, placeholderDesc, placeholderLimit string
	if since != "" {
		params = append(params, since)
		placeholderSince = `AND created>=$`+strconv.Itoa(len(params))
		if desc {
			placeholderSince = `AND created<=$`+strconv.Itoa(len(params))
		}
	}
	if desc {
		placeholderDesc = `DESC`
	}
	if limit != 0 {
		params = append(params, limit)
		placeholderLimit = `LIMIT $`+strconv.Itoa(len(params))
	}
	if counter==13{
		fmt.Println()
	}
	query = fmt.Sprintf(query, placeholderSince, placeholderDesc, placeholderLimit)
	rows, err = r.DB.Query(query, params...)
	if err != nil {
		return threads, err
	}

	for rows.Next(){
		thread := structs.Thread{}
		var created time.Time
		err := rows.Scan(&thread.Author, &thread.Forum, &created, &thread.Id, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
		thread.Created = created.Format(structs.OutTimeFormat)
		if err != nil {
			return threads, err
		}
		threads = append(threads, thread)
	}
	if len(threads) == 0 {
		var sl string
		err = r.DB.QueryRow(`SELECT slug from Forum WHERE lower(slug)=lower($1)`, forumSlug).Scan(&sl)
		if err != nil {
			return threads, err
		}
	}
	return threads, nil

}

func (r *PostgresRepo) EditThread(thread structs.Thread) (error) {
	query := `UPDATE Thread SET %s WHERE %s=$1`

	paramCount := 1
	set := []string{}
	var params []interface{}
	var key interface{}
	var keyName string

	if thread.Id != 0 {
		key=thread.Id
		keyName="id"
	} else {
		key = thread.Slug
		keyName="slug"
	}
	params = append(params, key)

	if thread.Message != "" {
		paramCount++
		set = append(set, "message=$"+strconv.Itoa(paramCount))
		params = append(params, thread.Message)
	}
	if thread.Title != "" {
		paramCount++
		set = append(set, "title=$"+strconv.Itoa(paramCount))
		params = append(params, thread.Title)
	}

	query = fmt.Sprintf(query, strings.Join(set, ", "), keyName)
	_, err := r.DB.Exec(query, params...)


	return err
}
