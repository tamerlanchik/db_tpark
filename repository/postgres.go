package repository

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
)

type PostgresRepo struct{
	DB *sql.DB
	queries map[int]string
}

const (
	queryGetUserByNick=`SELECT email, nickname, fullname, about FROM Users WHERE nickname=$1`
	queryGetPost=`SELECT author, created, forum, id, isEdited, message, parent, thread FROM Post WHERE id=$1`
	queryGetForum = `SELECT posts, threads, title, usernick FROM Forum WHERE slug=$1`
	queryGetThread=`SELECT author, created, forum, id, message, slug, title, votes FROM Thread WHERE id=$1`
)

func NewPostgresRepo() *PostgresRepo {
	return &PostgresRepo{}
}

func (r *PostgresRepo) Init(user, pass, host, port, dbname string) error {
	dsnTemplate := "postgres://%s:%s@%s:%s/%s?sslmode=disable"
	dsn := fmt.Sprintf(dsnTemplate, user, pass, host, port, dbname)

	var err error
	r.DB, err = sql.Open("pgx", dsn)
	if err != nil {
		return err
	}

	return r.DB.Ping()
}

func (r *PostgresRepo) ClearAll() error {
	query := `
			TRUNCATE TABLE vote CASCADE;
			TRUNCATE TABLE Post CASCADE;
			TRUNCATE TABLE Thread CASCADE;
			TRUNCATE TABLE Forum CASCADE;
			TRUNCATE TABLE Users CASCADE;
		`
	_, err := r.DB.Exec(query)
	return err
}

func (r *PostgresRepo) GetDBAccount() (map[string]int64, error) {
	//query := `SELECT COUNT(Forum.slug) AS forum, COUNT(Post.id) AS post, COUNT(Thread.id) as thread, COUNT(Users.nickname) AS user
	//			FROM Forum `
	query := `SELECT * FROM (SELECT COUNT(Post.id) AS post FROM Post) AS Post,
							(SELECT COUNT(Thread.id) AS thread FROM Thread) AS Thread,
							(SELECT COUNT(Forum.slug) AS forum FROM Forum) AS Forum,
							(SELECT COUNT(Users.nickname) AS user FROM Users) AS Users;`
	var posts, threads, forums, users int64
	err := r.DB.QueryRow(query).Scan(&posts, &threads, &forums, &users)
	res := make(map[string]int64, 4)
	res["forum"] = forums
	res["thread"] = threads
	res["user"] = users
	res["post"] = posts
	return res, err
}

