package http

import (
	"github.com/ATenderholt/lambda-router/internal/repo"
	"github.com/ATenderholt/lambda-router/settings"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type LayerHandler struct {
	cfg         *settings.Config
	layerRepo   repo.LayerRepository
	runtimeRepo repo.RuntimeRepository
}

func (h LayerHandler) GetAllLayerVersions(response http.ResponseWriter, request *http.Request) {
	layerName := chi.URLParam(request, "layerName")

	ctx := request.Context()

	layers, err := h.layerRepo.GetLayerByName(ctx, layerName)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}

	result := lambda.ListLayerVersionsOutput{
		LayerVersions: layersToAwsLayers(layers, h.cfg),
		NextMarker:    nil,
	}

	respondWithJson(response, result)
}
