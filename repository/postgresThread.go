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
	query := `INSERT INTO Thread (author,forum,message,created,title, slug) VALUES ($1, (SELECT slug from Forum WHERE lower(slug)=lower($2)), $3, %s::timestamptz, $5, %s) RETURNING id, slug, votes, forum`
	counter++
	query = fmt.Sprintf(query, `COALESCE($4, NOW())`, `NULLIF($6, '')`)
	var slug sql.NullString
	t, _ := time.Parse(structs.OutTimeFormat, thread.Created)
	err := r.DB.QueryRow(query, thread.Author, thread.Forum, thread.Message, t, thread.Title, thread.Slug).
			Scan(&thread.Id, &slug, &thread.Votes, &thread.Forum)
	thread.Slug = slug.String
	if counter>=125{
		fmt.Println(counter)
	}
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
		} else {
			err = structs.InternalError{E: structs.ErrorNoForum}
		}
	}

	return thread, err
}

func (r *PostgresRepo) GetThread(slug string) (structs.Thread, error) {
	var thread structs.Thread
	query := `SELECT author, created, forum, id, message, slug, title, votes FROM Thread WHERE lower(slug)=lower($1)`

	var created time.Time
	err := r.DB.QueryRow(query, slug).
		Scan(&thread.Author, &created, &thread.Forum,
			&thread.Id, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	thread.Created = created.Format(structs.OutTimeFormat)
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

func (r *PostgresRepo) GetThreadUnknownKey(key interface{}) (structs.Thread, error) {
	//if threadId, ok := key.(int64); ok {
	//	return r.GetThreadById(threadId)
	//} else {
	//	return r.GetThread(key.(string))
	//}
	threadId, err := r.getThreadId(key)
	if err != nil {
		return structs.Thread{}, err
	}
	return r.GetThreadById(threadId)
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

func (r *PostgresRepo) VoteThread(threadKey interface{}, user string, voice int) error {
	counter++
	fmt.Println(counter)
	id, err := r.getThreadId(threadKey)
	if err != nil {
		return err
	}
	if counter==131{
		fmt.Println(counter)
	}
	query := `SELECT vote_thread($1, $2, $3)`
	_, err = r.DB.Exec(query, id, user, voice)
	return err
}

func (r *PostgresRepo) getThreadId(threadKey interface{}) (int64, error) {

	var threadId int64
	var ok bool
	if threadId, ok = threadKey.(int64); ok {
		return threadId, nil
	}
	if threadId, err := strconv.ParseInt(threadKey.(string), 10, 64); err == nil {
		return threadId, nil
	}
	err := r.DB.QueryRow(`SELECT * FROM get_thread_id_by_slug($1)`, threadKey.(string)).Scan(&threadId)
	//queryGetThreadId := `SELECT id FROM Thread WHERE lower(slug)=lower($1);`
	//err := r.DB.QueryRow(queryGetThreadId, threadKey.(string)).Scan(&threadId)
	if err != nil {
		return -1, err
	}
	return threadId, nil
}