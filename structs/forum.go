package structs

import (
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx"
)

type Forum struct {
	Posts int64 `json:"posts"`
	Slug string `json:"slug"`
	Threads int32 `json:"threads"`
	Title string `json:"title"`
	User string `json:"user"`
}
func (f *Forum) InflateFromSql(row pgx.Row) error{
	//	posts, threads, title, usernick
	var slug pgtype.Text
	err := row.Scan(
		&f.Posts,
		&f.Threads,
		&f.Title,
		&f.User,
		&slug,
	)
	f.Slug = slug.String
	return err
}