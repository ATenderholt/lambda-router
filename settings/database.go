package settings

import (
	"fmt"
	"path/filepath"
)

const (
	DefaultDbFilename  = "db.sqlite3"
	InMemoryDbFilename = ":memory:"
)

type Database struct {
	Filename string
	Options  string
}

func (db *Database) connectionString(basePath string) string {
	if db.Filename == InMemoryDbFilename {
		return fmt.Sprintf("file:%s%s", db.Filename, db.Options)
	}

	path := filepath.Join(basePath, db.Filename)
	return fmt.Sprintf("file:%s", path)
}

func DefaultDatabase() *Database {
	return &Database{Filename: DefaultDbFilename, Options: ""}
}

func InMemoryDatabase() *Database {
	return &Database{Filename: ":memory:", Options: "?cache=shared"}
}
