package repository

import (
	"2019_2_Next_Level/pkg/sqlTools"
	"database/sql"
	"db_tpark/structs"
	"fmt"
	"github.com/jackc/pgconn"
	"strconv"
	"strings"
	"time"
)

func (r *PostgresRepo) GetPost(id int64) (structs.Post, error) {
	query := queryGetPost

	var post structs.Post
	var created time.Time
	err := r.DB.QueryRow(query, id).
		Scan(&post.Author, &created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)
	post.Created = created.Format(structs.OutTimeFormat)
	post.ChangeParent()
	return post, err
}
func (r *PostgresRepo) GetPostAccount(id int64, fields []string) (structs.PostAccount, error) {
	var postAccount structs.PostAccount
	post, err := r.DB.Prepare(queryGetPost)
	if err != nil {
		return postAccount, err
	}
	var author, forum, thread *sql.Stmt
	for _, key := range fields {
		switch key {
		case "user":
			author, err = r.DB.Prepare(queryGetUserByNick)
			if err != nil {
				return postAccount, err
			}
			postAccount.Author = &structs.User{}
			break
		case "forum":
			thread, err = r.DB.Prepare(queryGetThread)
			if err != nil {
				return postAccount, err
			}
			postAccount.Forum = &structs.Forum{}
			break
		case "thread":
			thread, err = r.DB.Prepare(queryGetThread)
			if err != nil {
				return postAccount, err
			}
			postAccount.Thread = &structs.Thread{}
			break
		default:
			return postAccount, fmt.Errorf("Unknown elem: ", key)
		}
	}


	task := func() error {
		err := postAccount.Post.InflateFromSql(post.QueryRow(id))
		if err != nil {
			return err
		}

		if author != nil {
			err = postAccount.Author.InflateFromSql(author.QueryRow(postAccount.Post.Author))
			if err != nil {
				return err
			}
		}
		if forum != nil {
			err = postAccount.Forum.InflateFromSql(forum.QueryRow(postAccount.Post.Forum))
			if err != nil {
				return err
			}
		}

		if thread != nil {
			err = postAccount.Thread.InflateFromSql(thread.QueryRow(postAccount.Post.Thread))
			if err != nil {
				return err
			}
		}
		return nil
	}

	err = sqlTools.WithTransaction(r.DB, task)
	return postAccount, err
}

func (r *PostgresRepo) EditPost(id int64, newPost structs.Post) error {
	query := `UPDATE Post SET %s WHERE id=$1`
	paramCount := 1
	set := []string{}
	var params []interface{}
	params = append(params, id)
	if newPost.Message != "" {
		paramCount++
		set = append(set, "message=$"+strconv.Itoa(paramCount))
		params = append(params, newPost.Message)
	}
	if newPost.Parent!=0{
		paramCount++
		set = append(set, "parent=$"+strconv.Itoa(paramCount))
		params = append(params, newPost.Parent)
	}
	query = fmt.Sprintf(query, strings.Join(set, ", "))
	_, err := r.DB.Exec(query, params...)
	return err
}

func (r *PostgresRepo) CreatePost(thread interface{}, posts []structs.Post) ([]structs.Post, error) {
	if len(posts) == 0 {
		return posts, nil
	}
	query := `INSERT INTO Post (author, message, parent, thread) VALUES `
	postfix := `RETURNING forum, id, created`

	if len(posts) == 1{
		query = query + `($1, $2, $3, $4) ` + postfix
	}else{
		query = sqlTools.CreatePacketQuery(query, 4, len(posts), postfix)
	}

	threadId, err := r.getThreadId(thread)
	if err != nil {
		return posts, err
	}

	var params []interface{}
	for _, post := range posts {
		params = append(params, post.Author, post.Message, post.Parent, threadId)
	}

	rows, err := r.DB.Query(query, params...)
	if err != nil {
		switch err.(*pgconn.PgError).Code {
		default:
			return posts, structs.InternalError{E:"Unknown error"}
		}
	}
	i := 0
	for rows.Next() {
		var created time.Time
		err := rows.Scan(&(posts[i].Forum), &(posts[i].Id), &(created))
		if err != nil {
			return posts, structs.InternalError{E: err.Error()}
		}
		posts[i].Created = created.Format(structs.OutTimeFormat)
		posts[i].IsEdited = false
		posts[i].Thread = int32(threadId)
		i++
	}
	if i==0 && len(posts) > 0{
		// looking for exact error
		if _, err:=r.DB.Exec(`SELECT id from Thread WHERE id=$1;`, posts[0].Thread); err != nil {
			return posts, structs.InternalError{E:structs.ErrorNoThread}
		}else{
			return posts, structs.InternalError{E:structs.ErrorNoParent}
		}
	}
	return posts, nil
}

func (r *PostgresRepo) GetPosts(threadKey interface{}, limit int64, since string, sort string, desc bool) ([]structs.Post, error) {
	threads := make([]structs.Post, 0)
	threadId, err := r.getThreadId(threadKey)
	if err != nil {
		return threads, err
	}

	var rows *sql.Rows
	params := make([]interface{}, 0,2)
	params = append(params, threadId)
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
	//if counter==13{
	//	fmt.Println()
	//}
	if sort=="flat" {
		query := `SELECT author, forum, created, id, isEdited, message, parent, thread FROM Post WHERE thread=$1 %s ORDER BY created, id %s %s`

		query = fmt.Sprintf(query, placeholderSince, placeholderDesc, placeholderLimit)
		rows, err = r.DB.Query(query, params...)
		if err != nil {
			return threads, err
		}

		for rows.Next(){
			thread := structs.Post{}
			var created time.Time
			err := rows.Scan(&thread.Author, &thread.Forum, &created, &thread.Id, &thread.IsEdited, &thread.Message, &thread.Parent, &thread.Thread)
			thread.Created = created.Format(structs.OutTimeFormat)
			thread.ChangeParent()
			if err != nil {
				return threads, err
			}
			threads = append(threads, thread)
		}
		if len(threads) == 0 {
			var sl string
			err = r.DB.QueryRow(`SELECT slug from thread WHERE id=$1`, threadId).Scan(&sl)
			if err != nil {
				return threads, err
			}
		}
	}

	return threads, nil
}