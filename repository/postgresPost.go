package repository

import (
	"2019_2_Next_Level/pkg/sqlTools"
	"database/sql"
	"db_tpark/structs"
	"fmt"
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
	post.Created = created.Format(time.RFC3339)
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