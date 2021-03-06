package repository

import (
	"context"
	"database/sql"
	"sync/atomic"

	//"db_tpark/buildmode"
	"db_tpark/pkg/sqlTools"
	"db_tpark/structs"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	//mutexMap = make(map[string]*sync.Mutex)
	mutexMapMutex = sync.Mutex{}
}

//func GetMutex(key string) {
//	mutex, ok := mutexMap[key]
//	if !ok {
//		mutex = &sync.Mutex{}
//		mutexMapMutex.Lock()
//		mutexMap[key]=mutex
//		mutexMapMutex.Unlock()
//	}
//	mutex.Lock()
//}
//func FreeMutex(key string) {
//	mutexMap[key].Unlock()
//}
var postCounter int64
//var mutexMap map[string]*sync.Mutex
var mutexMapMutex sync.Mutex
//var UsersForum map[string][]string

func (r *PostgresRepo) GetPost(id int64) (structs.Post, error) {
	query := queryGetPost
	var post structs.Post
	var created time.Time
	err := r.DB.QueryRow(context.Background(), query, id).
		Scan(&post.Author, &created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)
	post.Created = created.Format(structs.OutTimeFormat)
	post.ChangeParent()
	return post, err
}

func (r *PostgresRepo) GetPostAccount(id int64, fields []string) (structs.PostAccount, error) {
	var postAccount structs.PostAccount
	var author, forum, thread *sql.Stmt
	for _, key := range fields {
		switch key {
		case "user":
			author = &sql.Stmt{}
			postAccount.Author = &structs.User{}
			break
		case "forum":
			forum = &sql.Stmt{}
			postAccount.Forum = &structs.Forum{}
			break
		case "thread":
			thread = &sql.Stmt{}
			postAccount.Thread = &structs.Thread{}
			break
		default:
			return postAccount, fmt.Errorf("Unknown elem: ", key)
		}
	}


	task := func() error {
		err := postAccount.Post.InflateFromSql(r.DB.QueryRow(context.Background(), queryGetPost, id))
		if err != nil {
			return err
		}

		if author != nil {
			err = postAccount.Author.
				InflateFromSql(r.DB.QueryRow(context.Background(), queryGetUserByNick, postAccount.Post.Author))
			if err != nil {
				return err
			}
		}
		if forum != nil {
			if err:=r.loadForumPosts(); err != nil {
				return err
			}
			err = postAccount.Forum.
				InflateFromSql(r.DB.QueryRow(context.Background(), queryGetForum, postAccount.Post.Forum))
			if err != nil {
				return err
			}
		}

		if thread != nil {
			//if err:=r.loadThreadVotes(); err != nil {
			//	return err
			//}
			err = postAccount.Thread.
				InflateFromSql(r.DB.QueryRow(context.Background(), queryGetThread, postAccount.Post.Thread))
			if err != nil {
				return err
			}
		}
		return nil
	}

	err := task()
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
	if len(set) == 0 {
		return nil
	}
	query = fmt.Sprintf(query, strings.Join(set, ", "))
	_, err := r.DB.Exec(context.Background(), query, params...)
	//err := sqlTools.WithTransaction(r.DB, func() error {
	//	_, err := r.DB.Exec(context.Background(), query, params...)
	//	return err
	//})
	return err
}

func (r *PostgresRepo) CreatePost(thread interface{}, posts []structs.Post) ([]structs.Post, error) {
	postCounter++
	threadId, err := r.getThreadId(thread)
	if err != nil {
		return posts, structs.InternalError{E: structs.ErrorNoThread}
	}

	var cnt int64;
	if row:=r.DB.QueryRow(context.Background(), `SELECT count(id) from Thread WHERE id=$1;`, threadId); row.Scan(&cnt)!=nil || cnt==0 {
		return posts, structs.InternalError{E: structs.ErrorNoThread}
	}
	if len(posts) == 0 {
		return posts, nil
	}
	var forumSlug string
	err = r.DB.QueryRow(context.Background(), `SELECT forum FROM Thread WHERE Thread.id=$1`, threadId).Scan(&forumSlug)
	if err!=nil {
		return posts, structs.InternalError{E: structs.ErrorNoThread, Explain:err.Error()}
	}

	userList := make(map[string]bool)
	postPacketSize := 30
	firstCreated := time.Now()
	for i:=0; i<len(posts); i+=postPacketSize {
		currentPacket := posts[i:int(math.Min(float64(i+postPacketSize), float64(len(posts))))]
		currentPacket, err = r.createPostsByPacket(threadId, forumSlug, currentPacket, firstCreated)
		if err != nil {
			return posts, err
		}
		for j, post := range currentPacket{
			posts[i+j] = post
			userList[post.Author] = true
		}
	}

	query := `UPDATE ForumPosts SET posts=posts+$2 WHERE forum=$1;`
	_, err = r.DB.Exec(context.Background(), query, forumSlug, len(posts))
	if err != nil {
		return posts, structs.InternalError{E: structs.ErrorNoThread, Explain:err.Error()}
	}
	atomic.AddInt32(&forumPostsAccess.hasNewUpdates, 1)
	prefix := `INSERT INTO UsersInForum(nickname, forum) VALUES `
	postfix := `ON CONFLICT DO NOTHING`
	query = sqlTools.CreatePacketQuery(prefix, 2, len(userList), postfix)
	params := make([]interface{}, 0, len(userList))
	for key := range userList {
		params = append(params, key, forumSlug)
	}
	//GetMutex(forumSlug)
	//defer FreeMutex(forumSlug)
	mutexMapMutex.Lock()
	defer mutexMapMutex.Unlock()
	_, err = r.DB.Exec(context.Background(), query, params...)
	if err != nil {
		return posts, structs.InternalError{E: structs.ErrorNoThread, Explain:err.Error()}
	}
	return posts, nil
}

func (r *PostgresRepo) createPostsByPacket(threadId int64, forumSLug string, posts []structs.Post, created time.Time) ([]structs.Post, error) {
	var params []interface{}
	for _, post := range posts {
		var parent sql.NullInt64
		parent.Int64=post.Parent
		if post.Parent!=0 {
			parent.Valid=true
		}
		//var parent pgtype.Int8;
		//parent.Int = post.Parent;
		//if post.Parent!=0 {
		//	parent.Status = pgtype.Present
		//} else {
		//	parent.Status = pgtype.Null
		//}
		params = append(params, post.Author, post.Message, parent, threadId, created, forumSLug)
	}

	query := `INSERT INTO Post (author, message, parent, thread, created, forum) VALUES `
	postfix := `RETURNING forum, id, created`

	query = sqlTools.CreatePacketQuery(query, 6, len(posts), postfix)


	rows, err := r.DB.Query(context.Background(), query, params...)
	defer rows.Close()
	if err != nil || (rows!=nil && rows.Err()!=nil){
		//switch err.(*pgconn.PgError).Code {
		//default:
			return posts, structs.InternalError{E:"Unknown error", Explain:err.Error()}
		//}
	}
	i := 0
	for rows.Next() {
		var created time.Time
		err := rows.Scan(&(posts[i].Forum), &(posts[i].Id), &(created))
		if err != nil {
			return posts, structs.InternalError{E: err.Error(), Explain:err.Error()}
		}
		posts[i].Created = created.Format(structs.OutTimeFormat)
		posts[i].IsEdited = false
		posts[i].Thread = int32(threadId)
		i++
	}

	var cnt int64;
	if i==0 && len(posts) > 0{
		// looking for exact error
		if row:=r.DB.QueryRow(context.Background(), `SELECT count(id) from Thread WHERE id=$1;`, threadId); row.Scan(&cnt)!=nil || cnt==0 {
			return posts, structs.InternalError{E: structs.ErrorNoThread}
		} else if row:= r.DB.QueryRow(context.Background(), `SELECT COUNT(nickname) FROM Users WHERE nickname=$1`, posts[0].Author); row.Scan(&cnt)!=nil || cnt==0 {
			return posts, structs.InternalError{E: structs.ErrorNoThread}

		}else{
			return posts, structs.InternalError{E:structs.ErrorNoParent, Explain: "Third branch"}
		}
	}
	return posts, nil
}
