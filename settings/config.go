package settings

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
)

const (
	configContextKey = contextKey("config")

	DefaultAccountNumber = "271828182845"
	DefaultRegion        = "us-west-2"

	DefaultDataPath = "data"

	DefaultLambdaPort = 9002
)

type contextKey string

type Config struct {
	AccountNumber string
	IsDebug       bool
	Region        string

	Database *Database

	dataPath string
}

func (config *Config) ArnFragment() string {
	return config.Region + ":" + config.AccountNumber
}

func (config *Config) CreateDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", config.DbConnectionString())
	if err != nil {
		log.Panicf("unable to open database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Panicf("unable to ping database: %v", err)
	}

	return db
}

func (config *Config) DataPath() string {
	if config.dataPath[0] == '/' {
		return config.dataPath
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return filepath.Join(cwd, config.dataPath)
}

func (config *Config) DbConnectionString() string {
	return config.Database.connectionString(config.dataPath)
}

func DefaultConfig() *Config {
	logger.Debugf("Creating directory %s if necessary ...", DefaultDataPath)

	err := os.MkdirAll(DefaultDataPath, 0755)
	if err != nil {
		panic(err)
	}

	return &Config{
		AccountNumber: DefaultAccountNumber,
		IsDebug:       false,
		Region:        DefaultRegion,
		Database:      DefaultDatabase(),
		dataPath:      DefaultDataPath,
	}
}
