package structs

import (
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx"
	"time"
)

type Thread struct {
	Author string	`json:"author,omitempty"`
	Created string	`json:"created,omitempty"`
	Forum string	`json:"forum,omitempty"`
	Id int32		`json:"id,omitempty"`
	Message string	`json:"message,omitempty"`
	Slug string		`json:"slug,omitempty"`
	Title string	`json:"title,omitempty"`
	Votes int32		`json:"votes,omitempty"`
}

func (t *Thread) InflateFromSql(row pgx.Row) error {
	//	author, created, forum, id, message, slug, title, votes
	var created time.Time
	var slug pgtype.Text
	err :=  row.Scan(
			&t.Author,
			&created,
			&t.Forum,
			&t.Id,
			&t.Message,
			&slug,
			&t.Title,
			&t.Votes,
		)
	t.Created = created.Format(OutTimeFormat)
	t.Slug = slug.String
	return err
}
