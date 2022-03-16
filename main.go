package main

import (
	"context"
	"embed"
	"github.com/ATenderholt/lambda-router/internal/http"
	"github.com/ATenderholt/lambda-router/logging"
	"github.com/ATenderholt/lambda-router/settings"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var logger *zap.SugaredLogger

//go:embed migrations/*.sql
var embedMigrations embed.FS

func init() {
	logger = logging.NewLogger()
}

func main() {
	cfg := settings.DefaultConfig()
	mainCtx := context.Background()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(mainCtx)
	go func() {
		s := <-c
		logger.Infof("Received signal %v", s)
		cancel()
	}()

	if err := start(ctx, cfg); err != nil {
		logger.Errorf("Failed to start: %v", err)
	}
}

func start(ctx context.Context, config *settings.Config) error {
	logger.Info("Starting up ...")

	initializeDb(config)

	//initializeDocker(ctx)
	server, err := http.Serve(config)
	if err != nil {
		panic(err)
	}

	<-ctx.Done()

	logger.Info("Shutting down ...")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer func() {
		cancel()
	}()

	err = server.Shutdown(ctxShutDown)
	if err != nil {
		logger.Error("Error when shutting down HTTP server")
	}
	//
	//err = docker.ShutdownAll(ctxShutDown)
	//if err != nil {
	//	logger.Error("Errors when shutting down docker containers: %v", err)
	//}

	return nil
}

func initializeDb(config *settings.Config) {
	db := config.CreateDatabase()
	defer db.Close()

	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(logging.GooseLogger{logger})

	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}
}
