package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/pgxpool"
	_ "github.com/jackc/pgx/stdlib"
	//"time"
)

type PostgresRepo struct{
	DB *pgxpool.Pool
	queries map[int]string
}

const (
	queryGetUserByNick=`SELECT email, nickname, fullname, about FROM Users WHERE nickname=$1`
	queryGetPost=`SELECT author, created, forum, id, isEdited, message, coalesce(parent,0), thread FROM Post WHERE id=$1`
	queryGetForum = `SELECT (SELECT ForumPosts.posts FROM ForumPosts WHERE ForumPosts.forum=slug), threads, title, usernick, slug FROM Forum WHERE slug=$1`
	//queryGetThread=`SELECT author, created, forum, id, message, slug, title, tv.votes FROM Thread
	//				JOIN ThreadVotes as tv on tv.thread=id WHERE id=$1`
	queryGetThread=`SELECT author, created, forum, id, message, slug, title, votes FROM Thread 
					WHERE id=$1`
)

func NewPostgresRepo() *PostgresRepo {
	return &PostgresRepo{}
}

func (r *PostgresRepo) Init(user, pass, host, port, dbname string) error {
	dsnTemplate := "postgres://%s:%s@%s:%s/%s?pool_max_conns=10"
	dsn := fmt.Sprintf(dsnTemplate, user, pass, host, port, dbname)

	var err error
	//r.DB, err = sql.Open("pgx", dsn)
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return err
	}
	r.DB, err = pgxpool.ConnectConfig(context.Background(), poolConfig)

	//r.DB.SetMaxOpenConns(8)
	//r.DB.SetMaxIdleConns(4)
	//r.DB.SetConnMaxLifetime(time.Second*1)
	return err
}

func (r *PostgresRepo) ClearAll() error {
	query := `
			TRUNCATE ForumPosts;
			TRUNCATE UsersInForum;
			TRUNCATE TABLE vote CASCADE;
			TRUNCATE TABLE Post CASCADE;
			TRUNCATE TABLE Thread CASCADE;
			TRUNCATE TABLE Forum CASCADE;
			TRUNCATE TABLE Users CASCADE;			
		`
	_, err := r.DB.Exec(context.Background(), query)
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
	err := r.DB.QueryRow(context.Background(), query).Scan(&posts, &threads, &forums, &users)
	res := make(map[string]int64, 4)
	res["forum"] = forums
	res["thread"] = threads
	res["user"] = users
	res["post"] = posts
	return res, err
}

