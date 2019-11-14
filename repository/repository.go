package repository

import "db_tpark/structs"

type Repository interface{
	AddUser(user structs.User) error
	GetUser(email, nickname string) (structs.User, error)
	EditUser(user structs.User) error

	CreateForum(slug, title, user string) error
	CreateAndReturnForum(slug, title, user string) (structs.Forum, error)
	GetForum(slug string) (structs.Forum, error)

	CreateThread(thread structs.Thread) (structs.Thread, error)
	GetThread(slud string) (structs.Thread, error)
	GetThreads(forumSlug string, limit int64, since string, desc bool) ([]structs.Thread, error)
	GetUsers(forumSlug string, limit int64, since string, desc bool) ([]structs.User, error)

	GetPost(id int64) (structs.Post, error)
	GetPostAccount(id int64, fields []string) (structs.PostAccount, error)
	EditPost(id int64, newPost structs.Post) error

	ClearAll() error
}