package database

import (
	"context"
	"database/sql"
	"log"
)

type Database interface {
	BeginTx(ctx context.Context) (Transaction, error)
	Close()
	Exec(query string, args ...interface{}) (sql.Result, error)
	InsertOne(ctx context.Context, query string, args ...interface{}) (int64, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type RealDatabase struct {
	Wrapped *sql.DB
}

func CreateConnection(connStr string) Database {
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		log.Panicf("unable to open database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Panicf("unable to ping database: %v", err)
	}

	var database Database = RealDatabase{db}
	return database
}

func (db RealDatabase) Close() {
	err := db.Wrapped.Close()
	if err != nil {
		log.Panicf("unable to close database: %v", err)
	}
}

func (db RealDatabase) BeginTx(ctx context.Context) (Transaction, error) {
	options := sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false}
	tx, err := db.Wrapped.BeginTx(ctx, &options)
	if err != nil {
		return nil, err
	}

	var transaction Transaction = RealTransaction{tx}
	return transaction, nil
}

func (db RealDatabase) InsertOne(ctx context.Context, query string, args ...interface{}) (int64, error) {
	insert, err := db.Wrapped.ExecContext(ctx, query, args...)
	if err != nil {
		debug := buildDebug(query, args)
		logger.Errorf("Unable to insert %s: %s", debug, err)
		return -1, Error{"unable to insert", debug, err}
	}

	count, err := insert.RowsAffected()
	if err != nil {
		logger.Error("Unexpected error when inserting: %v", err)
		return -1, unexpectedError{"unexpected error when inserting", err}
	}

	if count != 1 {
		logger.Error("Expected only 1 insert but got %d", count)
		return -1, UnexpectedRowCountError{Actual: count, Expected: 1, Op: "insert"}
	}

	id, err := insert.LastInsertId()
	if err != nil {
		logger.Error("Unexpected error when inserting")
		return -1, unexpectedError{"unexpected error when inserting", err}
	}

	return id, nil
}

func (db RealDatabase) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.Wrapped.Exec(query, args...)
}

func (db RealDatabase) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return db.Wrapped.PrepareContext(ctx, query)
}

func (db RealDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.Wrapped.QueryContext(ctx, query, args...)
}

func (db RealDatabase) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.Wrapped.QueryRow(query, args...)
}

func (db RealDatabase) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.Wrapped.QueryRowContext(ctx, query, args...)
}
