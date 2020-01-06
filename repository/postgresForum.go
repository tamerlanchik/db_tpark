package repository

import (
	"context"
	"sync"

	//"database/sql"
	"db_tpark/structs"
	//"fmt"
	pg "github.com/jackc/pgconn"
)

type SyncFlag struct {
	mutex sync.Mutex
	hasNewUpdates int32
}

var forumPostsAccess SyncFlag

func init() {
	forumPostsAccess.mutex = sync.Mutex{}
}

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


func (r *PostgresRepo) GetForum(slug string) (structs.Forum, error) {
	var forum structs.Forum
	if err:=r.loadForumPosts(); err != nil {
		return forum, err
	}

	query := `SELECT posts, threads, title, usernick, slug FROM Forum WHERE lower(slug)=lower($1);`
	//query := `SELECT (SELECT ForumPosts.posts FROM ForumPosts WHERE ForumPosts.forum=slug), threads, title, usernick, slug FROM Forum WHERE lower(slug)=lower($1);`

	err := r.DB.QueryRow(context.Background(), query, slug).Scan(&forum.Posts, &forum.Threads, &forum.Title, &forum.User, &forum.Slug)
	return forum, err
}


func (r *PostgresRepo) loadForumPosts() error{
	if forumPostsAccess.hasNewUpdates==0{
		return nil
	}
	tx, err := r.DB.Begin(context.Background())
	if err != nil {
		return err
	}
	forumPostsAccess.mutex.Lock()
	defer forumPostsAccess.mutex.Unlock()
	if forumPostsAccess.hasNewUpdates==0{
		return nil
	}
	err = func () error {
		_, err := tx.Exec(context.Background(), `UPDATE Forum SET posts=Forum.posts+temp.posts FROM ForumPosts as temp WHERE temp.forum=Forum.slug`)
		//_, err := tx.Exec(context.Background(), `UPDATE thread SET votes=thread.votes+temp.votes FROM threadvotes as temp WHERE temp.thread=Thread.id`)
		if err != nil {
			return err
		}
		//_, err = tx.Exec(context.Background(), `UPDATE threadvotes SET votes=0`)
		_, err = tx.Exec(context.Background(), `UPDATE ForumPosts SET posts=0`)
		return err
	}()
	if err != nil {
		return tx.Rollback(context.Background())
	}
	err =tx.Commit(context.Background())
	if err != nil {
		return err
	}
	forumPostsAccess.hasNewUpdates = 0
	return nil
}

