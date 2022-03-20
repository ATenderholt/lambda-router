package http

import (
	"github.com/ATenderholt/lambda-router/internal/docker"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewChiMux(layerHandler LayerHandler, functionHandler FunctionHandler, eventHandler EventSourceHandler,
	docker *docker.Manager) *chi.Mux {

	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)

	r.Get("/2018-10-31/layers/{layerName}/versions/{layerVersion}", layerHandler.GetLayerVersion)
	r.Get("/2018-10-31/layers/{layerName}/versions", layerHandler.GetAllLayerVersions)
	r.Post("/2018-10-31/layers/{layerName}/versions", layerHandler.PostLayerVersions)

	r.Get("/2020-06-30/functions/{name}/code-signing-config", functionHandler.GetFunctionCodeSigning)
	r.Get("/2015-03-31/functions/{name}/versions", functionHandler.GetFunctionVersions)
	r.Put("/2015-03-31/functions/{name}/configuration", functionHandler.PutLambdaConfiguration)
	r.Get("/2015-03-31/functions/{name}", functionHandler.GetLambdaFunction)
	r.Post("/2015-03-31/functions", functionHandler.PostLambdaFunction)

	r.Post("/2015-03-31/functions/{name}/invocations", docker.Invoke)

	r.Post("/2015-03-31/event-source-mappings", eventHandler.PostEventSource)
	r.Get("/2015-03-31/event-source-mappings/{id}", eventHandler.GetEventSource)

	return r
}
