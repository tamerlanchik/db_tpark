package structs

import "database/sql"

type Forum struct {
	Posts int64 `json:"posts"`
	Slug string `json:"slug"`
	Threads int32 `json:"threads"`
	Title string `json:"title"`
	User string `json:"user"`
}
func (f *Forum) InflateFromSql(row *sql.Row) error{
	//	posts, threads, title, usernick
	var slug sql.NullString
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