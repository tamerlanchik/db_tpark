package repository

import (
	"database/sql"
	"db_tpark/structs"
	"fmt"
)

func (r *PostgresRepo) AddUser(user structs.User) error {
	query := `INSERT INTO Users (email, nickname, fullname, about) VALUES($1, $2, $3, $4);`

	_, err := r.DB.Exec(query, user.Email, user.Nickname, user.Fullname, user.About)
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
		query = query + `nickname=$1;`
		param = nickname
	}else{
		return user, fmt.Errorf("Empty params")
	}

	err := r.DB.QueryRow(query, param).Scan(&user.Email, &user.Nickname, &user.Fullname, &user.About)
	return user, err
}

func (r *PostgresRepo) EditUser(user structs.User) error {
	query := `UPDATE Users SET email=$1, fullname=$2, about=$3 WHERE nickname=$4;`

	_, err := r.DB.Exec(query, user.Email, user.Fullname, user.About, user.Nickname)
	return err
}

func (r *PostgresRepo) GetUsers(forumSlug string, limit int64, since string, desc bool) ([]structs.User, error) {
	users := make([]structs.User, 0)
	query := `SELECT about, email, fullname, nickname FROM Users WHERE nickname IN (
--     			(SELECT DISTINCT usernick as "author" FROM Forum WHERE slug=$1 AND usernick>=$2)
--     			UNION
				(SELECT DISTINCT author FROM Thread WHERE forum=$1 AND author>=$2)
				UNION
				(SELECT DISTINCT author FROM Post WHERE forum=$1 AND author>=$2)
				) ORDER BY nickname %s LIMIT $3;`
	if desc {
		query = fmt.Sprintf(query, "DESC")
	}else{
		query = fmt.Sprintf(query, "")
	}

	var rows *sql.Rows
	var err error
	rows, err = r.DB.Query(query, forumSlug, since, limit)
	if err != nil {
		return users, err
	}

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
