package structs

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
