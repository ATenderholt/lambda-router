package http

import (
	"fmt"
	"github.com/ATenderholt/lambda-router/internal/repo"
	"github.com/ATenderholt/lambda-router/pkg/database"
	"github.com/ATenderholt/lambda-router/settings"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func Serve(cfg *settings.Config) (srv *http.Server, err error) {

	db := database.RealDatabase{Wrapped: cfg.CreateDatabase()}
	functionRepo := repo.NewFunctionRepository(db)
	layerRepo := repo.NewLayerRepository(db)
	runtimeRepo := repo.NewRuntimeRepository(db)

	functionHandler := FunctionHandler{
		cfg:          cfg,
		functionRepo: functionRepo,
		layerRepo:    layerRepo,
		runtimeRepo:  runtimeRepo,
	}

	layerHandler := LayerHandler{
		cfg:       cfg,
		layerRepo: layerRepo,
	}

	r := chi.NewRouter()
	r.Get("/2018-10-31/layers/{layerName}/versions/{layerVersion}", layerHandler.GetLayerVersion)
	r.Get("/2018-10-31/layers/{layerName}/versions", layerHandler.GetAllLayerVersions)
	r.Post("/2018-10-31/layers/{layerName}/versions", layerHandler.PostLayerVersions)

	r.Get("/2020-06-30/functions/{name}/code-signing-config", functionHandler.GetFunctionCodeSigning)
	r.Get("/2015-03-31/functions/{name}/versions", functionHandler.GetFunctionVersions)
	r.Put("/2015-03-31/functions/{name}/configuration", functionHandler.PutLambdaConfiguration)
	r.Get("/2015-03-31/functions/{name}", functionHandler.GetLambdaFunction)
	r.Post("/2015-03-31/functions", functionHandler.PostLambdaFunction)

	srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.BasePort),
		Handler: r,
	}

	go func() {
		e := srv.ListenAndServe()
		if e != nil && e != http.ErrServerClosed {
			logger.Errorf("Problem starting HTTP server: %v", e)
			err = e
		}
	}()

	logger.Infof("Finished starting HTTP server on port %d", cfg.BasePort)

	return
}
