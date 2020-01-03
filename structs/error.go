package structs

const (
	ErrorNoUser = "Nouser"
	ErrorDuplicateKey = "DuplicateKey"
	ErrorNoForum = "NoForum"
	ErrorNoThread = "NoThread"
	ErrorNoParent = "NoParent"
)

type Error struct {
	Message string `json:"message"`
}

type InternalError struct {
	E string
	Explain string
}

func (e InternalError) Error() string {
	return e.E
}
