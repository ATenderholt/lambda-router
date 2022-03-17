package http

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ATenderholt/lambda-router/internal/domain"
	"github.com/ATenderholt/lambda-router/settings"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	aws "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/go-chi/chi/v5"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
)

type LayerHandler struct {
	cfg         *settings.Config
	layerRepo   domain.LayerRepository
	runtimeRepo domain.RuntimeRepository
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

func (h LayerHandler) GetLayerVersion(writer http.ResponseWriter, request *http.Request) {
	layerName := chi.URLParam(request, "layerName")
	layerVersionStr := chi.URLParam(request, "layerVersion")
	layerVersion, err := strconv.Atoi(layerVersionStr)
	if err != nil {
		msg := fmt.Sprintf("Unable to convert version %s to integer", layerVersionStr)
		logger.Errorf(msg)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	logger.Infof("Getting version %d of layer %s", layerVersion, layerName)
	ctx := request.Context()

	layer, err := h.layerRepo.GetLayerByNameAndVersion(ctx, layerName, layerVersion)
	if err != nil {
		logger.Error(err.Error())
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	logger.Info("Found %+v", layer)

	result := lambda.GetLayerVersionOutput{
		CompatibleArchitectures: []aws.Architecture{},
		CompatibleRuntimes:      layer.CompatibleRuntimes,
		Content: &aws.LayerVersionContentOutput{
			CodeSize:   layer.CodeSize,
			CodeSha256: &layer.CodeSha256,
		},
		CreatedDate:     &layer.CreatedOn,
		Description:     &layer.Description,
		LayerArn:        layer.GetArn(h.cfg),
		LayerVersionArn: layer.GetVersionArn(h.cfg),
		LicenseInfo:     nil,
		Version:         int64(layer.Version),
	}

	respondWithJson(writer, result)
}

func (h LayerHandler) PostLayerVersions(writer http.ResponseWriter, request *http.Request) {
	layerName := chi.URLParam(request, "layerName")
	dec := json.NewDecoder(request.Body)

	var body lambda.PublishLayerVersionInput
	err := dec.Decode(&body)

	if err != nil {
		msg := fmt.Sprintf("Problem parsing request for Lambda layer %s: %v", layerName, err)
		logger.Error(msg)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	logger.Infof("Layer description: %s", *body.Description)
	logger.Infof("Layer runtimes: %v", body.CompatibleRuntimes)

	ctx := request.Context()

	dbRuntimes, err := h.runtimeRepo.RuntimeIDsByNames(ctx, body.CompatibleRuntimes)
	switch {
	case err == sql.ErrNoRows:
		msg := fmt.Sprintf("Unable to find all expected runtimes: %v", body.CompatibleRuntimes)
		logger.Error(msg)
		http.Error(writer, msg, http.StatusNotFound)
	case err != nil:
		msg := fmt.Sprintf("Unable to query all runtimes: %v", err)
		logger.Error(msg)
		http.Error(writer, msg, http.StatusInternalServerError)
	}

	version, err := h.layerRepo.GetLatestLayerVersionByName(ctx, layerName)
	switch {
	case err == sql.ErrNoRows:
		version = -1
	case err != nil:
		msg := fmt.Sprintf("Unable to get latest Layer for %s: %v", layerName, err)
		logger.Error(msg)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	logger.Info("Found latest verion for layer %s: %v", layerName, version)
	logger.Info("Saving %d bytes from zipfile", len(body.Content.ZipFile))

	rawHash := sha256.Sum256(body.Content.ZipFile)
	hash := base64.StdEncoding.EncodeToString(rawHash[:])

	layer := domain.LambdaLayer{
		Name:               layerName,
		Version:            version + 1,
		Description:        *body.Description,
		CompatibleRuntimes: body.CompatibleRuntimes,
		CodeSize:           int64(len(body.Content.ZipFile)),
		CodeSha256:         hash,
	}

	destPath := layer.GetDestPath(h.cfg)
	logger.Info("Saving layer %s to %s...", layerName, destPath)
	err = createDirs(filepath.Dir(destPath))
	if err != nil {
		msg := fmt.Sprintf("unable to create parent directory for layer %s: %v", destPath, err)
		logger.Error(msg)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	err = ioutil.WriteFile(destPath, body.Content.ZipFile, 0644)
	if err != nil {
		msg := fmt.Sprintf("error when saving layer %s: %v", layerName, err)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	savedLayer, err := h.layerRepo.InsertLayer(ctx, layer, &dbRuntimes)
	if err != nil {
		logger.Error(err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	result := savedLayer.ToPublishLayerVersionOutput(h.cfg)

	respondWithJson(writer, result)
}
