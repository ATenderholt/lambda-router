package settings

import (
	"bytes"
	"database/sql"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultAccountNumber = "271828182845"
	DefaultRegion        = "us-west-2"

	DefaultBasePort = 9050
	DefaultDataPath = "data"

	DefaultDevConfigFile = "functions.yml"
	DefaultSqsEndpoint   = "http://localhost:9324"
	DefaultNetworks      = "lambda"
)

type Config struct {
	AccountNumber string
	IsDebug       bool
	IsLocal       bool
	Region        string

	Database *Database

	BasePort int
	dataPath string

	DevConfigFile string
	Networks      []string
	SqsEndpoint   string
}

func (config *Config) ArnFragment() string {
	return config.Region + ":" + config.AccountNumber
}

func (config *Config) CreateDatabase() *sql.DB {
	connStr := config.DbConnectionString()
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		log.Panicf("unable to open database %s: %v", connStr, err)
	}

	err = db.Ping()
	if err != nil {
		log.Panicf("unable to ping database %s: %v", connStr, err)
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
		IsLocal:       true,
		Region:        DefaultRegion,
		Database:      DefaultDatabase(),
		BasePort:      DefaultBasePort,
		dataPath:      DefaultDataPath,
		DevConfigFile: DefaultDevConfigFile,
		SqsEndpoint:   DefaultSqsEndpoint,
		Networks:      []string{DefaultNetworks},
	}
}

type NetworkValue struct {
	networks []string
}

func (v *NetworkValue) Set(s string) error {
	v.networks = strings.Split(s, ",")
	return nil
}

func (v *NetworkValue) String() string {
	if len(v.networks) > 0 {
		return strings.Join(v.networks, ",")
	}

	return ""
}

func FromFlags(name string, args []string) (*Config, string, error) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)

	var buf bytes.Buffer
	flags.SetOutput(&buf)

	var cfg Config
	var dbFileName string
	networks := NetworkValue{[]string{DefaultNetworks}}
	flags.StringVar(&cfg.AccountNumber, "account-number", DefaultAccountNumber, "Account number returned in ARNs")
	flags.BoolVar(&cfg.IsDebug, "debug", false, "Enable debug logging")
	flags.BoolVar(&cfg.IsLocal, "local", true, "Application should use localhost when routing lambda")
	flags.StringVar(&cfg.Region, "region", DefaultRegion, "Region returned in ARNs")
	flags.IntVar(&cfg.BasePort, "port", DefaultBasePort, "Port used for HTTP and start of port range for individual lambdas")
	flags.StringVar(&cfg.dataPath, "data-path", DefaultDataPath, "Path to persist data and lambdas")
	flags.StringVar(&cfg.DevConfigFile, "config", DefaultDevConfigFile, "Config file for starting lambdas in Development mode")
	flags.StringVar(&cfg.SqsEndpoint, "sqs-endpoint", DefaultSqsEndpoint, "Endpoint for SQS services (i.e. lambda triggers)")
	flags.Var(&networks, "networks", "Comma-separated list of Networks for lambda containers")
	flags.StringVar(&dbFileName, "db", DefaultDbFilename, "Database file for persisting lambda configuration")

	err := flags.Parse(args)
	if err != nil {
		return nil, buf.String(), err
	}

	cfg.Database = DefaultDatabase()
	cfg.Database.Filename = dbFileName
	cfg.Networks = networks.networks

	return &cfg, buf.String(), err
}
