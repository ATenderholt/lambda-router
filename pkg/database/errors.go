package database

import "fmt"

type unexpectedError struct {
	msg  string
	base error
}

func (e unexpectedError) Error() string {
	return e.msg + ": " + e.base.Error()
}

type Error struct {
	Msg   string
	Query string
	Base  error
}

func (e Error) Error() string {
	return e.Msg + " [" + e.Query + "] :" + e.Base.Error()
}

type UnexpectedRowCountError struct {
	Op       string
	Expected int64
	Actual   int64
}

func (e UnexpectedRowCountError) Error() string {
	return fmt.Sprintf("expected only %d during %s, but got %d", e.Expected, e.Op, e.Actual)
}
