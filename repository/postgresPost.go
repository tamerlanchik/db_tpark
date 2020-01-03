package repository

import (
	"database/sql"
	"db_tpark/pkg/sqlTools"
	"db_tpark/structs"
	"fmt"
	"github.com/jackc/pgconn"
	"strconv"
	"strings"
	"time"
)
var postCounter int64
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
			forum, err = r.DB.Prepare(queryGetForum)
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
	if len(set) == 0 {
		return nil
	}
	query = fmt.Sprintf(query, strings.Join(set, ", "))
	err := sqlTools.WithTransaction(r.DB, func() error {
		_, err := r.DB.Exec(query, params...)
		return err
	})
	return err
}

func (r *PostgresRepo) CreatePost(thread interface{}, posts []structs.Post) ([]structs.Post, error) {
	postCounter++
	threadId, err := r.getThreadId(thread)
	if err != nil {
		return posts, structs.InternalError{E: structs.ErrorNoThread}
	}

	var cnt int64;
	if row:=r.DB.QueryRow(`SELECT count(id) from Thread WHERE id=$1;`, threadId); row.Scan(&cnt)!=nil || cnt==0 {
		return posts, structs.InternalError{E: structs.ErrorNoThread}
	}
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

	var params []interface{}
	for _, post := range posts {
		var parent sql.NullInt64;
		parent.Int64 = post.Parent;
		if post.Parent!=0 {
			parent.Valid = true;
		}
		params = append(params, post.Author, post.Message, parent, threadId)
	}


	rows, err := r.DB.Query(query, params...)
	if err != nil || (rows!=nil && rows.Err()!=nil){
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

	//var cnt int64;
	if i==0 && len(posts) > 0{
		// looking for exact error
		if row:=r.DB.QueryRow(`SELECT count(id) from Thread WHERE id=$1;`, threadId); row.Scan(&cnt)!=nil || cnt==0 {
			return posts, structs.InternalError{E: structs.ErrorNoThread}
		} else if row:= r.DB.QueryRow(`SELECT COUNT(nickname) FROM Users WHERE nickname=$1`, posts[0].Author); row.Scan(&cnt)!=nil || cnt==0 {
			return posts, structs.InternalError{E: structs.ErrorNoThread}

		}else{
			return posts, structs.InternalError{E:structs.ErrorNoParent}
		}
	}
	return posts, nil
}

//func (r *PostgresRepo) CreatePost(thread interface{}, posts []structs.Post) ([]structs.Post, error) {
//	threadId, err := r.getThreadId(thread)
//	if err != nil {
//		return posts, structs.InternalError{E: structs.ErrorNoThread}
//	}
//
//	// Есть тест на несуществующий тред с пустым списком постов
//	if err := r.checkThreadExists(threadId); err != nil {
//		return posts, err
//	}
//	if len(posts) == 0 {
//		return posts, nil
//	}
//
//	query := `INSERT INTO Post (author, message, parent, thread, created) VALUES
//                                                          ($1, $2, $3, $4, $5)
//                                                          RETURNING forum, id, created`
//	var i int64
//	firstCreated := time.Now()
//	//createdString := firstCreated.Format(structs.OutTimeFormat)
//	for j, post := range posts {
//		var parent sql.NullInt64;
//		parent.Int64 = post.Parent;
//		if post.Parent!=0 {
//			parent.Valid = true;
//		}
//
//		rows, err := r.DB.Query(query, post.Author, post.Message, parent, threadId, firstCreated)
//		if err != nil || (rows!=nil && rows.Err()!=nil){
//			switch err.(*pgconn.PgError).Code {
//			default:
//				return posts, structs.InternalError{E:"Unknown error"}
//			}
//		}
//
//		var localCount int
//		for rows.Next() {
//			var created time.Time
//			err := rows.Scan(&(posts[i].Forum), &(posts[i].Id), &(created))
//			if err != nil {
//				return posts, structs.InternalError{E: err.Error()}
//			}
//			posts[j].Created = created.Format(structs.OutTimeFormat)
//			posts[j].IsEdited = false
//			posts[j].Thread = int32(threadId)
//			i++
//			localCount++
//		}
//		if localCount==0{
//			// выясняем проблему
//			var cnt int64
//			if row:=r.DB.QueryRow(`SELECT count(id) from Thread WHERE id=$1;`, threadId); row.Scan(&cnt)!=nil || cnt==0 {
//				return posts, structs.InternalError{E: structs.ErrorNoThread}
//			} else if row:= r.DB.QueryRow(`SELECT COUNT(nickname) FROM Users WHERE nickname=$1`, posts[0].Author); row.Scan(&cnt)!=nil || cnt==0 {
//				return posts, structs.InternalError{E: structs.ErrorNoThread}
//
//			}else{
//				return posts, structs.InternalError{E:structs.ErrorNoParent}
//			}
//		}
//	}
//	return posts, nil
//}


func (r *PostgresRepo) GetPosts(threadKey interface{}, limit int64, since string, sort string, desc bool) ([]structs.Post, error) {
	query := `SELECT author, forum, created, id, isEdited, message, coalesce(parent, 0), thread 
				FROM Post 
					WHERE thread=$1 %s ORDER BY %s %s`
	//		WHERE thread=$1
	threads := make([]structs.Post, 0)
	threadId, err := r.getThreadId(threadKey)
	if err != nil {
		return threads, err
	}

	var rows *sql.Rows
	params := make([]interface{}, 0,2)
	params = append(params, threadId)
	var placeholderSince, placeholderDesc, placeholderLimit string
	if desc {
		placeholderDesc = "DESC"
	} else {
		placeholderDesc = "ASC"
	}

	if limit != 0 {
		params = append(params, limit)
		placeholderLimit = `LIMIT $`+strconv.Itoa(len(params))
	}
	if since != "" {
		params = append(params, since)
		var compareSign string
		if desc {
			compareSign = "<"
		} else {
			compareSign = ">"
		}
		paramNum := len(params)
		queryGetPath := `SELECT %s FROM post AS since WHERE since.id=%s`
		switch sort {
		case "flat":
			placeholderSince = fmt.Sprintf(`AND id%s$%d`, compareSign, paramNum)
		case "tree":
			placeholderSince = fmt.Sprintf(
				`AND path%s(%s)`,
				compareSign,
				fmt.Sprintf(queryGetPath, `since.path`, fmt.Sprintf(`$%d`, paramNum)),
			)
		case "parent_tree":
			placeholderSince = fmt.Sprintf(
				`AND parents.path[1]%s(%s)`,
				compareSign,
				fmt.Sprintf(queryGetPath, `since.path[1]`, fmt.Sprintf(`$%d`, paramNum)),
			)
		}
	}
	switch sort {
	case "flat":
		condition := placeholderSince
		order := fmt.Sprintf(`(created, id) %s`, placeholderDesc)
		query = fmt.Sprintf(query, condition, order, placeholderLimit)
	case "tree":
		order := fmt.Sprintf(`(path, created) %s`, placeholderDesc)
		query = fmt.Sprintf(query, placeholderSince, order, placeholderLimit)
	case "parent_tree":
		condition := fmt.Sprintf(
			`AND path[1] IN (
						SELECT parents.id FROM post AS parents
						WHERE parents.thread=post.thread AND parents.parent IS NULL %s
						ORDER BY parents.path[1] %s 
						%s
						)`,
			placeholderSince, placeholderDesc, placeholderLimit,
		)
		order := fmt.Sprintf(`path[1] %s, path`, placeholderDesc)
		query = fmt.Sprintf(query, condition, order, "")
	}
	rows, err = r.DB.Query(query, params...)
	if err != nil {
		//buildmode.Log.Println(err)
		return threads, err
	}

	for rows.Next(){
		thread := structs.Post{}
		var created time.Time
		err := rows.Scan(&thread.Author, &thread.Forum, &created, &thread.Id, &thread.IsEdited, &thread.Message, &thread.Parent, &thread.Thread)
		thread.Created = created.Format(structs.OutTimeFormat)
		thread.ChangeParent()
		if err != nil {
			//buildmode.Log.Println(err)
			return threads, err
		}
		threads = append(threads, thread)
	}
	if len(threads) == 0 {
		var sl sql.NullString
		err = r.DB.QueryRow(`SELECT slug from thread WHERE id=$1`, threadId).Scan(&sl)
		if err != nil {
			return threads, err
		}
	}
	return threads, nil
}