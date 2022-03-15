package database

import (
	"context"
	"database/sql"
	"fmt"
)

type Transaction interface {
	Commit() error
	InsertOne(ctx context.Context, query string, args ...interface{}) (int64, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Rollback(format string, v ...interface{}) string
}

type RealTransaction struct {
	wrapped *sql.Tx
}

func (tx RealTransaction) Commit() error {
	return tx.wrapped.Commit()
}

func (tx RealTransaction) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return tx.wrapped.PrepareContext(ctx, query)
}

func (tx RealTransaction) Rollback(format string, v ...interface{}) string {
	err := tx.wrapped.Rollback()
	msg := fmt.Sprintf(format, v...)
	if err != nil {
		return msg + ", couldn't rollback"
	}

	return msg
}

func (tx RealTransaction) InsertOne(ctx context.Context, query string, args ...interface{}) (int64, error) {
	insert, err := tx.wrapped.ExecContext(ctx, query, args...)
	if err != nil {
		debug := buildDebug(query, args)
		logger.Errorf("Unable to insert %s: %s", debug, err)
		return -1, Error{"unable to insert", debug, err}
	}

	count, err := insert.RowsAffected()
	if err != nil {
		msg := tx.Rollback("unexpected error when inserting")
		logger.Error(msg)
		return -1, unexpectedError{msg, err}
	}

	if count != 1 {
		msg := tx.Rollback("expected only 1 insert but got %d", count)
		logger.Error(msg)
		return -1, fmt.Errorf(msg)
	}

	id, err := insert.LastInsertId()
	if err != nil {
		msg := tx.Rollback("unexpected error when inserting")
		logger.Error(msg)
		return -1, unexpectedError{msg, err}
	}

	return id, nil
}
