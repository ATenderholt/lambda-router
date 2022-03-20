//go:build wireinject
// +build wireinject

package main

import (
	"fmt"
	"github.com/ATenderholt/lambda-router/internal/docker"
	"github.com/ATenderholt/lambda-router/internal/domain"
	handler "github.com/ATenderholt/lambda-router/internal/http"
	"github.com/ATenderholt/lambda-router/internal/repo"
	"github.com/ATenderholt/lambda-router/pkg/database"
	"github.com/ATenderholt/lambda-router/settings"
	"github.com/go-chi/chi/v5"
	"github.com/google/wire"
	"net/http"
)

func NewApp(cfg *settings.Config, mux *chi.Mux, docker *docker.Manager, functionRepo domain.FunctionRepository) App {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.BasePort),
		Handler: mux,
	}

	return App{
		port:         cfg.BasePort,
		srv:          srv,
		docker:       docker,
		functionRepo: functionRepo,
	}
}

func RealDatabase(cfg *settings.Config) database.Database {
	return database.RealDatabase{
		Wrapped: cfg.CreateDatabase(),
	}
}

var db = wire.NewSet(
	RealDatabase,
	repo.NewFunctionRepository,
	repo.NewLayerRepository,
	repo.NewRuntimeRepository,
	repo.NewEventSourceRepository,
	// have to tell wire how to map interface to concrete type
	wire.Bind(new(domain.FunctionRepository), new(*repo.FunctionRepository)),
	wire.Bind(new(domain.LayerRepository), new(*repo.LayerRepository)),
	wire.Bind(new(domain.RuntimeRepository), new(*repo.RuntimeRepository)),
	wire.Bind(new(domain.EventSourceRepository), new(*repo.EventSourceRepository)),
)

var api = wire.NewSet(
	handler.NewFunctionHandler,
	handler.NewLayerHandler,
	handler.NewEventSourceHandler,
	handler.NewChiMux,
)

func InjectApp(cfg *settings.Config) (App, error) {
	wire.Build(
		NewApp,
		db,
		api,
		docker.NewManager,
	)
	return App{}, nil
}
