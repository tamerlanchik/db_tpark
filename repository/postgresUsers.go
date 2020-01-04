package repository

import (
	"context"
	"db_tpark/structs"
	"fmt"
	"github.com/jackc/pgconn"
)

var counter int64

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
	//buildmode.Log.Println("a")
	return err
}

func (r *PostgresRepo) GetUsers(forumSlug string, limit int64, since string, desc bool) ([]structs.User, error) {
	users := make([]structs.User, 0)

	query := `SELECT about, email, fullname, nickname FROM Users
				JOIN (SELECT nickname from UsersInForum WHERE forum=$1 %s) as l
					USING (nickname) ORDER BY nickname %s %s`
	//query := `SELECT about, email, fullname, nickname FROM Users
	//			JOIN UsersInForum as l USING (nickname)
	//			WHERE forum=$1 %s ORDER BY nickname %s %s`
	var cmpPlaceholder, limitPlaceholder, orderPlaceholder string
	paramsCount := 1
	params := make([]interface{}, 0)
	params = append(params, forumSlug)
	if desc {
		if since!="" {
			cmpPlaceholder = `AND nickname<$2`
			paramsCount++
			params = append(params, since)
		}
		orderPlaceholder = "DESC"
	}else{
		if since!="" {
			cmpPlaceholder = `AND nickname>$2`
			paramsCount++
			params = append(params, since)
		}
		orderPlaceholder = "ASC"
	}
	if limit!=0 {
		paramsCount++
		limitPlaceholder = fmt.Sprintf(`LIMIT $%d`, paramsCount)
		params = append(params, limit)
	}
	query = fmt.Sprintf(query, cmpPlaceholder, orderPlaceholder, limitPlaceholder)

	var err error
	//buildmode.Log.Println(query)
	//buildmode.Log.Println(forumSlug, since, limit)
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
