package repository

import "db_tpark/structs"

type Repository interface{
	AddUser(user structs.User) error
	GetUser(email, nickname string) (structs.User, error)
	EditUser(user structs.User) error

	CreateForum(slug, title, user string) error
	GetForum(slug string) (structs.Forum, error)

	CreateThread(thread structs.Thread) (structs.Thread, error)
	GetThread(key interface{}) (structs.Thread, error)
	GetThreads(forumSlug string, limit int64, since string, desc bool) ([]structs.Thread, error)
	GetUsers(forumSlug string, limit int64, since string, desc bool) ([]structs.User, error)

	GetPost(id int64) (structs.Post, error)
	GetPostAccount(id int64, fields []string) (structs.PostAccount, error)
	GetPosts(threadSKey interface{}, limit int64, since string, sort string, desc bool) ([]structs.Post, error)
	EditPost(id int64, newPost structs.Post) error

	ClearAll() error
	GetDBAccount() (map[string]int64, error)

	CreatePost(threadId interface{}, post []structs.Post) ([]structs.Post, error)
	EditThread(thread structs.Thread) (error)
	VoteThread(threadId interface{}, user string, voice int) error

}