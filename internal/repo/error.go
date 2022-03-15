package repo

import "strconv"

type Error struct {
	Msg  string
	Base error
}

func (e Error) Error() string {
	return e.Msg + ": " + e.Base.Error()
}

type RowError struct {
	Op   string
	Row  int
	Base error
}

func (e RowError) Error() string {
	return "unable to scan row #" + strconv.Itoa(e.Row) + " for operation " + e.Op + ": " + e.Base.Error()
}
