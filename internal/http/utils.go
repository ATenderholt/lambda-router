package http

import (
	"encoding/json"
	"fmt"
	"github.com/ATenderholt/rainbow-functions/internal/domain"
	"github.com/ATenderholt/rainbow-functions/settings"
	aws "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"net/http"
	"os"
)

func createDirs(dirPath string) error {
	logger.Debugf("Creating directory if necessary %s ...", dirPath)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		e := fmt.Errorf("unable to create directory %s: %v", dirPath, err)
		logger.Error(e)
		return e
	}

	return nil
}

func layersToAwsLayers(layers []domain.LambdaLayer, cfg *settings.Config) []aws.LayerVersionsListItem {
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
