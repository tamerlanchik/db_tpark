package structs

import (
	"database/sql"
	"time"
)

type Post struct {
	Author string	`json:"author,omitempty"`
	Created string	`json:"created,omitempty"`
	Forum string	`json:"forum,omitempty"`
	Id int64		`json:"id,omitempty"`
	IsEdited bool 	`json:"isEdited,omitempty"`
	Message string	`json:"message,omitempty"`
	Parent int64 `json:"parent"`
	Thread int32 `json:"thread,omitempty"`
}

type PostAccount struct {
	Author *User `json:"author,omitempty"`
	Forum *Forum `json:"forum,omitempty"`
	Post Post `json:"post"`
	Thread *Thread `json:"thread,omitempty"`
}

func (p *Post) InflateFromSql(row *sql.Row) error {
	var created time.Time
	err := row.Scan(
		&p.Author,
		&created,
		&p.Forum,
		&p.Id,
		&p.IsEdited,
		&p.Message,
		&p.Parent,
		&p.Thread,
	)
	p.Created = created.Format(time.RFC3339)
	return err
}