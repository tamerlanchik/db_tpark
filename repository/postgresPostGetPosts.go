package repository

import (
	"bytes"
	"context"
	"db_tpark/buildmode"
	"db_tpark/structs"
	"fmt"
	"github.com/jackc/pgtype"
	"strconv"
	"text/template"
	"time"
)
const (
	queryTemplateGetPostsSorted = `SELECT author, forum, created, Post.id, isEdited, message, coalesce(parent, 0), thread 
				FROM Post
					{{.Condition}}
					ORDER BY {{.OrderBy}}
					{{.Limit}}`
	queryTemplateGetPostsParentTree = `JOIN (
						SELECT parents.id FROM post AS parents
						WHERE parents.thread=$1 AND parents.parent IS NULL
							{{- if .Since}} AND {{.Since}}{{- end}}
						ORDER BY parents.path[1] {{.Desc}}
						{{.Limit}}
						) as p ON path[1]=p.id`
)

var getPostsTemplate *template.Template
var getPostsParentTreeTemplate *template.Template
func init() {
	var err error
	getPostsTemplate, err = template.New("getPosts").Parse(queryTemplateGetPostsSorted)
	if err != nil {
		fmt.Println("Error: cannot create getPostsTemplate template: ", err)
		panic(err)
	}

	getPostsParentTreeTemplate, err = template.New("parent_tree").Parse(queryTemplateGetPostsParentTree)
	if err != nil {
		fmt.Println("Error: cannot create getPostsParentTreeTemplate template: ", err)
		panic(err)
	}
}


func (r *PostgresRepo) GetPosts(threadKey interface{}, limit int64, since string, sort string, desc bool) ([]structs.Post, error) {
	mainTemplateArgs := struct{
		Condition string
		OrderBy string
		Limit string
	}{}

	threads := make([]structs.Post, 0)

	threadId, err := r.getThreadId(threadKey)
	if err != nil {
		return threads, err
	}

	params := make([]interface{}, 0,2)
	params = append(params, threadId)
	var placeholderSince, placeholderDesc string
	if desc {
		placeholderDesc = "DESC"
	} else {
		placeholderDesc = "ASC"
	}

	if limit != 0 {
		params = append(params, limit)
		mainTemplateArgs.Limit = `LIMIT $`+strconv.Itoa(len(params))
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
			//	AND id > $n
			placeholderSince = fmt.Sprintf(`AND id%s$%d`, compareSign, paramNum)
		case "tree":
			//	AND path[&n] > (SELECT since.path from Post AS since WHERE since.id=&n)
			placeholderSince = fmt.Sprintf(
				`AND path%s(%s)`,
				compareSign,
				fmt.Sprintf(queryGetPath, `since.path`, fmt.Sprintf(`$%d`, paramNum)),
			)
		case "parent_tree":
			//	AND parents[1] > (SELECT since.path[1] from Post AS since WHERE since.id=&n)
			placeholderSince = fmt.Sprintf(
				`parents.path[1]%s(%s)`,
				compareSign,
				fmt.Sprintf(queryGetPath, `since.path[1]`, fmt.Sprintf(`$%d`, paramNum)),
			)
		}
	}

	switch sort {
	case "flat":
		mainTemplateArgs.Condition = `WHERE thread=$1 ` + placeholderSince
		mainTemplateArgs.OrderBy = fmt.Sprintf(`(created, id) %s`, placeholderDesc)
	case "tree":
		mainTemplateArgs.OrderBy =  fmt.Sprintf(`(path, created) %s`, placeholderDesc)
		mainTemplateArgs.Condition = `WHERE thread=$1 ` + placeholderSince
	case "parent_tree":
		conditionBuffer := &bytes.Buffer{}
		err = getPostsParentTreeTemplate.Execute(conditionBuffer, struct{
			Since string
			Desc string
			Limit string
		}{Since: placeholderSince, Desc: placeholderDesc, Limit: mainTemplateArgs.Limit})
		if err != nil {
			return threads, err
		}
		mainTemplateArgs.Condition = conditionBuffer.String()
		mainTemplateArgs.OrderBy = fmt.Sprintf(`path[1] %s, path`, placeholderDesc)
		mainTemplateArgs.Limit = ""
	}

	queryBuffer:= &bytes.Buffer{}
	err = getPostsTemplate.Execute(queryBuffer, mainTemplateArgs)
	if err != nil {
		return threads, err
	}
	query := queryBuffer.String()

	rows, err := r.DB.Query(context.Background(), query, params...)
	if err != nil {
		buildmode.Log.Println(err)
		return threads, err
	}
	defer rows.Close()

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
		var sl pgtype.Text
		err = r.DB.QueryRow(context.Background(), `SELECT slug from thread WHERE id=$1`, threadId).Scan(&sl)
		if err != nil {
			return threads, err
		}
	}
	return threads, nil
}
