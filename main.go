package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"github.com/ATenderholt/dockerlib"
	"github.com/ATenderholt/lambda-router/internal/dev"
	"github.com/ATenderholt/lambda-router/internal/docker"
	"github.com/ATenderholt/lambda-router/internal/domain"
	"github.com/ATenderholt/lambda-router/internal/sqs"
	"github.com/ATenderholt/lambda-router/logging"
	"github.com/ATenderholt/lambda-router/settings"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var logger *zap.SugaredLogger

//go:embed migrations/*.sql
var embedMigrations embed.FS

func init() {
	logger = logging.NewLogger()
}

type App struct {
	cfg          *settings.Config
	srv          *http.Server
	functionRepo domain.FunctionRepository
	docker       *docker.Manager
	sqs          *sqs.Manager
	devService   *dev.Service
}

func (app App) Start() (err error) {
	ctx := context.Background()

	functions, err := app.functionRepo.GetAllLatestFunctions(ctx)
	if err != nil {
		logger.Error("Unable to query for Runtimes")
		return
	}

	for _, function := range functions {
		environment, e := app.functionRepo.GetEnvironmentForFunction(ctx, function)
		if e != nil {
			logger.Errorf("Unable to get Environment for Function %s: %v", function.FunctionName, e)
			err = e
			return
		}
		function.Environment = environment

		err = app.docker.StartFunction(ctx, &function)
		if err != nil {
			logger.Errorf("Unable to start Function %s: %v", function.FunctionName, err)
			return
		}
	}

	err = app.StartDevFunctions(ctx)
	if err != nil {
		logger.Errorf("Unable to start Dev functions: %v", err)
	}

	err = app.sqs.StartAllEventSources(ctx)
	if err != nil {
		logger.Errorf("Unable to start Event sources: %v", err)
		return
	}

	go func() {
		e := app.srv.ListenAndServe()
		if e != nil && e != http.ErrServerClosed {
			logger.Errorf("Problem starting HTTP server: %v", e)
			err = e
		}
	}()

	logger.Infof("Finished starting HTTP server on port %d", app.cfg.BasePort)
	return
}

func (app App) StartDevFunctions(ctx context.Context) error {
	stats, err := os.Stat(app.cfg.DevConfigFile)
	switch {
	case errors.Is(err, os.ErrNotExist):
		logger.Infof("Config file for Dev Functions doesn't exist, so not starting any.")
		return nil
	case err != nil:
		logger.Errorf("Unexpected error when getting info about Dev Functions file: %v", err)
		return err
	}

	if stats.IsDir() {
		err := fmt.Errorf("config file specified for Dev Functions is a directory: %s", app.cfg.DevConfigFile)
		logger.Error(err)
		return err
	}

	functions, err := dev.ParseFile(app.cfg.DevConfigFile)
	if err != nil {
		e := fmt.Errorf("unable to parse Dev Functions file: %v", err)
		logger.Error(e)
		return e
	}

	for _, function := range functions {
		var basePath string
		if filepath.IsAbs(function.BasePath) {
			basePath = function.BasePath
		} else {
			configPath, err := filepath.Abs(app.cfg.DevConfigFile)
			if err != nil {
				logger.Errorf("unable to get absolute path of config file %s: %v", app.cfg.DevConfigFile, err)
				continue
			}

			basePath = filepath.Join(filepath.Dir(configPath), function.BasePath)
		}

		dir, err := app.devService.InstallDependencies(ctx, function.Runtime, basePath)
		if err != nil {
			logger.Errorf("unable to install dependencies for Dev Function %s: %v", function.Name(), err)
			continue
		}

		function.DepPath = dir
		err = app.docker.StartFunction(ctx, function)
		if err != nil {
			logger.Errorf("unable to start Dev Function %s: %v", function.Name(), err)
		}
	}

	return nil
}

func (app App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	err := app.docker.ShutdownAll(ctx)
	if err != nil {
		logger.Error("Unable to shutdown Docker containers: %v", err)
	}

	err = app.srv.Shutdown(ctx)
	if err != nil {
		logger.Error("Unable to shutdown HTTP server: %v", err)
	}

	return err
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
