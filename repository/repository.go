package repository

import "db_tpark/structs"

type Repository interface{
	AddUser(user structs.User) error
	GetUser(email, nickname string) (structs.User, error)
	EditUser(user structs.User) error

	CreateForum(slug, title, user string) error
	CreateAndReturnForum(slug, title, user string) (structs.Forum, error)
	GetForum(slug string) (structs.Forum, error)
}