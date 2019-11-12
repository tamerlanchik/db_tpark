package repository

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
)

type PostgresRepo struct{
	DB *sql.DB
}

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


