package repository

import "database/sql"

type PostgresRepo struct{
	DB *sql.DB
}

func NewPostgresRepo() *PostgresRepo {
	return &PostgresRepo{}
}

func (r *PostgresRepo) Init(user, pass, host, port string) error {
	return nil
}


