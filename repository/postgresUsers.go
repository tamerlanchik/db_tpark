package repository

import (
	"bytes"
	"context"
	"db_tpark/structs"
	"fmt"
	"github.com/jackc/pgconn"
	"text/template"
)

var counter int64
var getForumUsersTemplate *template.Template
const (
	queryTemplateGetForumUsers = `SELECT about, email, fullname, nickname FROM Users
				JOIN (SELECT nickname from UsersInForum WHERE forum=$1 
					{{.Since}} ORDER BY nickname {{.Desc}} {{.Limit}}) as l
					USING (nickname) ORDER BY nickname {{.Desc}}`
)

func init() {
	var err error
	getForumUsersTemplate, err = template.New("getForumUsers").Parse(queryTemplateGetForumUsers)
	if err != nil {
		fmt.Println("Error: cannot create getForumUsersTemplate template: ", err)
		panic(err)
	}
}

func (r *PostgresRepo) AddUser(user structs.User) error {
	query := `INSERT INTO Users (email, nickname, fullname, about) VALUES($1, $2, $3, $4);`
	_, err := r.DB.Exec(context.Background(), query, user.Email, user.Nickname, user.Fullname, user.About)
	return err
}

func (r *PostgresRepo) GetUser(email, nickname string) (structs.User, error) {
	var user structs.User
	query := `SELECT email, nickname, fullname, about FROM Users WHERE `
	var param string
	if email!=""{
		query = query + `email=$1;`
		param = email
	} else if nickname!="" {
		query = query + `lower(nickname)=lower($1);`
		param = nickname
	}else{
		return user, fmt.Errorf("Empty params")
	}

	err := r.DB.QueryRow(context.Background(), query, param).Scan(&user.Email, &user.Nickname, &user.Fullname, &user.About)
	return user, err
}

func (r *PostgresRepo) EditUser(user structs.User) error {
	query := `UPDATE Users SET email=%s, fullname=%s, about=%s WHERE lower(nickname)=lower($4);`
	statements := []interface{}{
		structs.NoEmptyWrapper("email", 1),
		structs.NoEmptyWrapper("fullname", 2),
		structs.NoEmptyWrapper("about", 3),
	}
	query = fmt.Sprintf(query, statements...)

	_, err := r.DB.Exec(context.Background(), query, user.Email, user.Fullname, user.About, user.Nickname)
	if err != nil {
		switch err.(*pgconn.PgError).Code{
		case "23505":
			return structs.InternalError{E:structs.ErrorDuplicateKey}
		default:
			return structs.InternalError{E:structs.ErrorNoUser}
		}
	}
	return err
}

// Users of forum
func (r *PostgresRepo) GetUsers(forumSlug string, limit int64, since string, desc bool) ([]structs.User, error) {
	users := make([]structs.User, 0)
	templateArgs := struct {
		Since string
		Limit string
		Desc string
	}{}

	paramsCount := 1
	params := make([]interface{}, 0)
	params = append(params, forumSlug)
	if desc {
		if since!="" {
			templateArgs.Since = `AND nickname<$2`
			paramsCount++
			params = append(params, since)
		}
		templateArgs.Desc = "DESC"
	}else{
		if since!="" {
			templateArgs.Since = `AND nickname>$2`
			paramsCount++
			params = append(params, since)
		}
		templateArgs.Desc = "ASC"
	}
	if limit!=0 {
		paramsCount++
		templateArgs.Limit = fmt.Sprintf(`LIMIT $%d`, paramsCount)
		params = append(params, limit)
	}
	queryBuf := &bytes.Buffer{}
	err := getForumUsersTemplate.Execute(queryBuf, templateArgs)
	if err != nil {
		return users, err
	}
	query := queryBuf.String()
	rows, err := r.DB.Query(context.Background(), query, params...)
	if err != nil {
		return users, err
	}
	defer func() {
		rows.Close()
	}()


	for rows.Next(){
		user := structs.User{}
		err := rows.Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname)
		if err != nil {
			return users, err
		}
		users = append(users, user)
	}
	return users, nil
}
