package repository

import (
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
