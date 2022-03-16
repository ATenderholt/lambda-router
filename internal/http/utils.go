package http

import (
	"encoding/json"
	"fmt"
	"github.com/ATenderholt/lambda-router/internal/repo/types"
	"github.com/ATenderholt/lambda-router/settings"
	aws "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"net/http"
)

func layersToAwsLayers(layers []types.LambdaLayer, cfg *settings.Config) []aws.LayerVersionsListItem {
	results := make([]aws.LayerVersionsListItem, len(layers))
	for i, layer := range layers {
		results[i] = layer.ToLayerVersionsListItem(cfg)
	}

	return results
}

func respondWithJson(response http.ResponseWriter, value interface{}) {
	logger.Infof("Response: %+v", value)

	err := json.NewEncoder(response).Encode(value)
	if err != nil {
		msg := fmt.Sprintf("unable to return mashalled response for %+v: %v", value, err)
		http.Error(response, msg, http.StatusInternalServerError)
	}
}
