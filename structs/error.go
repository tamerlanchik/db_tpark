package structs

const (
	ErrorNoUser = "Nouser"
	ErrorDuplicateKey = "DuplicateKey"
	ErrorNoForum = "NoForum"
)

type Error struct {
	Message string `json:"message"`
}

type InternalError struct {
	E string
}

func (e InternalError) Error() string {
	return e.E
}