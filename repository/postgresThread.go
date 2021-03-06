package repository

import (
	"context"
	"db_tpark/buildmode"
	"db_tpark/structs"
	"fmt"
	pg "github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"strconv"
	"strings"
	"time"
)

var CreateThreadCounter int


func (r *PostgresRepo) CreateThread(thread structs.Thread) (structs.Thread, error) {
	query := `INSERT INTO Thread (author,forum,message,created,title, slug) VALUES 
					($1, 
					(SELECT slug from Forum WHERE lower(slug)=lower($2)),
					$3, COALESCE($4, NOW())::timestamptz, $5, NULLIF($6, ''))
				RETURNING id, slug, 0, forum`

	var slug pgtype.Text
	t, _ := time.Parse(structs.OutTimeFormat, thread.Created)
	err := r.DB.QueryRow(context.Background(), query, thread.Author, thread.Forum, thread.Message, t, thread.Title, thread.Slug).
			Scan(&thread.Id, &slug, &thread.Votes, &thread.Forum)
	thread.Slug = slug.String

	if err != nil {
		buildmode.Log.Println("Error in PSCreateThread: ", err)
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
				buildmode.Log.Println(e.Code)
				err = structs.InternalError{E: structs.ErrorNoForum}
			}
		} else {
			err = structs.InternalError{E: structs.ErrorNoForum}
		}
	}

	return thread, err
}

func (r *PostgresRepo) GetThread(key interface{}) (structs.Thread, error) {
	threadId, err := r.getThreadId(key)	// резолвим slug
	if err != nil {
		return structs.Thread{}, err
	}
	return r.getThreadById(threadId)
}

// Get Threads of forum
func (r *PostgresRepo) GetThreads(forumSlug string, limit int64, since string, desc bool) ([]structs.Thread, error) {
	threads := make([]structs.Thread, 0)
	query := `SELECT author, forum, created, id, message, slug, title, votes FROM Thread 
					WHERE lower(forum)=lower($1) %s ORDER BY created %s %s`


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
	query = fmt.Sprintf(query, placeholderSince, placeholderDesc, placeholderLimit)

	//if err:=r.loadThreadVotes(); err != nil {
	//	return threads, err
	//}

	rows, err := r.DB.Query(context.Background(), query, params...)
	defer rows.Close()
	if err != nil {
		return threads, err
	}

	var created time.Time
	var slug pgtype.Text
	for rows.Next(){
		thread := structs.Thread{}
		err := rows.Scan(&thread.Author, &thread.Forum, &created, &thread.Id, &thread.Message, &slug, &thread.Title, &thread.Votes)
		thread.Created = created.Format(structs.OutTimeFormat)
		thread.Slug = slug.String
		if err != nil {
			return threads, err
		}
		threads = append(threads, thread)
	}
	if len(threads) == 0 {
		var sl pgtype.Text
		err = r.DB.QueryRow(context.Background(), `SELECT slug from Forum WHERE lower(slug)=lower($1)`, forumSlug).Scan(&sl)
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
	if paramCount<=1 {
		return nil
	}
	query = fmt.Sprintf(query, strings.Join(set, ", "), keyName)
	_, err := r.DB.Exec(context.Background(), query, params...)

	return err
}

func (r *PostgresRepo) VoteThread(threadKey interface{}, user string, voice int) error {
	var id int64
	var ok bool
	if id, ok = threadKey.(int64); !ok {
		var err error
		id, err = r.getThreadId(threadKey)
		if err != nil {
			return err
		}
	}

	query := `INSERT INTO vote(thread, author, vote) VALUES ($1, $2, $3)
               ON CONFLICT ON CONSTRAINT vote_thread_author_key DO
           UPDATE SET vote=$3 WHERE vote.thread=$1 AND lower(vote.author)=lower($2)`
	_, err := r.DB.Exec(context.Background(), query, id, user, voice)

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
	query := `SELECT id FROM Thread WHERE lower(slug)=lower($1)`
	err := r.DB.QueryRow(context.Background(), query, threadKey.(string)).Scan(&threadId)
	if err != nil {
		return -1, err
	}
	return threadId, nil
}

func (r *PostgresRepo) getThreadById(id int64) (structs.Thread, error) {
	query := `SELECT author, created, forum, id, message, slug, title, votes FROM Thread WHERE id=$1`

	var thread structs.Thread
	var created time.Time
	var slug pgtype.Text
	err := r.DB.QueryRow(context.Background(), query, id).
		Scan(&thread.Author, &created, &thread.Forum,
			&thread.Id, &thread.Message, &slug, &thread.Title, &thread.Votes)
	thread.Created = created.Format(structs.OutTimeFormat)
	thread.Slug = slug.String

	return thread, err
}

func (r *PostgresRepo) checkThreadExists(id interface{}) error{
	var cnt int64;
	if row:=r.DB.QueryRow(context.Background(), `SELECT count(id) from Thread WHERE id=$1;`, id); row.Scan(&cnt)!=nil || cnt==0 {
		return structs.InternalError{E: structs.ErrorNoThread}
	}
	return nil
}


