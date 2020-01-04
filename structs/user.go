package structs

import (
	//"database/sql"
	"github.com/jackc/pgx"
)

type User struct {
	About string `json:"about,omitempty"`
	Email string `json:"email"`
	Fullname string `json:"fullname"`
	Nickname string `json:"nickname"`
}

func (u *User) InflateFromSql(row pgx.Row) error {
	//	email, nickname, fullname, about
	return row.Scan(
			&u.Email,
			&u.Nickname,
			&u.Fullname,
			&u.About,
		)
}
