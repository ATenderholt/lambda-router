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
	layerRepo := repo.NewLayerRepository(db)

	layerHandler := LayerHandler{
		cfg:       cfg,
		layerRepo: layerRepo,
	}

	r := chi.NewRouter()
	r.Get("/2018-10-31/layers/{layerName}/versions", layerHandler.GetAllLayerVersions)

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
