package dev

type Error struct {
	msg  string
	base error
}

func (e Error) Error() string {
	return e.msg + ": " + e.base.Error()
}
