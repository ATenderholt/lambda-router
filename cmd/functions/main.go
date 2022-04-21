package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"github.com/ATenderholt/dockerlib"
	"github.com/ATenderholt/rainbow-functions/logging"
	"github.com/ATenderholt/rainbow-functions/settings"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"os"
	"os/signal"
)

var logger *zap.SugaredLogger

//go:embed migrations/*.sql
var embedMigrations embed.FS

func init() {
	logger = logging.NewLogger()
}

func main() {
	cfg, output, err := settings.FromFlags(os.Args[0], os.Args[1:])
	if err == flag.ErrHelp {
		fmt.Println(output)
		os.Exit(2)
	} else if err != nil {
		fmt.Println("got error:", err)
		fmt.Println("output:\n", output)
		os.Exit(1)
	}

	mainCtx := context.Background()

	dockerlib.SetLogger(logging.NewLogger().Desugar().Named("dockerlib"))

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

	err := os.MkdirAll(config.DataPath(), 0755)
	if err != nil {
		logger.Errorf("Unable to make data directory: %v", err)
		return err
	}
	
	initializeDb(config)

	app, err := InjectApp(config)
	if err != nil {
		logger.Errorf("Unable to initialize application: %v", err)
		return err
	}

	err = app.Start()
	if err != nil {
		logger.Errorf("Unable to start application: %v", err)
		return err
	}

	//initializeDocker(ctx)

	<-ctx.Done()

	logger.Info("Shutting down ...")
	err = app.Shutdown()
	if err != nil {
		logger.Error("Error when shutting down app")
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
