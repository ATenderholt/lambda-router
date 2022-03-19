package http

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ATenderholt/lambda-router/internal/domain"
	"github.com/ATenderholt/lambda-router/pkg/zip"
	"github.com/ATenderholt/lambda-router/settings"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	aws "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type FunctionHandler struct {
	cfg          *settings.Config
	functionRepo domain.FunctionRepository
	layerRepo    domain.LayerRepository
	runtimeRepo  domain.RuntimeRepository
}

func NewFunctionHandler(cfg *settings.Config, functionRepo domain.FunctionRepository, layerRepo domain.LayerRepository,
	runtimeRepo domain.RuntimeRepository) FunctionHandler {
	return FunctionHandler{
		cfg:          cfg,
		functionRepo: functionRepo,
		layerRepo:    layerRepo,
		runtimeRepo:  runtimeRepo,
	}
}

func (f FunctionHandler) PostLambdaFunction(writer http.ResponseWriter, request *http.Request) {
	dec := json.NewDecoder(request.Body)

	var body lambda.CreateFunctionInput
	err := dec.Decode(&body)

	if err != nil {
		msg := fmt.Sprintf("error decoding %s: %v", request.Body, err)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	// move Code to new variable so that most of response body can be printed
	code := body.Code
	body.Code = nil
	logger.Infof("Creating lambda function %+v", body)

	ctx := request.Context()

	runtimeExists, err := f.runtimeRepo.RuntimeExistsByName(ctx, body.Runtime)
	if err != nil {
		msg := fmt.Sprintf("Error when querying runtime %s for function %s", body.Runtime, *body.FunctionName)
		logger.Error(msg)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	if !runtimeExists {
		msg := fmt.Sprintf("Unable to find runtime %s for function %s", body.Runtime, *body.FunctionName)
		logger.Error(msg)
		http.Error(writer, msg, http.StatusNotFound)
		return
	}

	dbVersion, err := f.functionRepo.GetLatestVersionForFunctionName(ctx, *body.FunctionName)
	if err != nil {
		msg := fmt.Sprintf("Error when finding latest version of function %s", *body.FunctionName)
		logger.Error(msg)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	function := domain.CreateFunction(&body)
	function.Version = strconv.Itoa(dbVersion + 1)
	rawHash := sha256.Sum256(code.ZipFile)
	function.CodeSha256 = base64.StdEncoding.EncodeToString(rawHash[:])

	// TODO : validate Layer runtime support

	err = zip.UncompressZipFileBytes(code.ZipFile, function.GetDestPath(f.cfg))
	if err != nil {
		msg := fmt.Sprintf("error when saving function %s: %v", *body.FunctionName, err)
		logger.Errorf(msg)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	layerDestPath := function.GetLayerDestPath(f.cfg)
	err = createDirs(layerDestPath)
	if err != nil {
		msg := fmt.Sprintf("Unable to create Layer path for Function %s: %v", function.FunctionName, err)
		logger.Errorf(msg)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	for _, layer := range function.Layers {
		layerPath := layer.GetDestPath(f.cfg)
		err = zip.UncompressZipFile(layerPath, layerDestPath)
		if err != nil {
			msg := fmt.Sprintf("error when unpacking layer %s: %v", layer.Name, err)
			logger.Errorf(msg)
			http.Error(writer, msg, http.StatusInternalServerError)
			return
		}
	}

	saved, err := f.functionRepo.InsertFunction(ctx, function)
	result := saved.ToCreateFunctionOutput(f.cfg)

	respondWithJson(writer, result)
}

func (f FunctionHandler) PutLambdaConfiguration(response http.ResponseWriter, request *http.Request) {
	name := chi.URLParam(request, "name")

	logger.Infof("Setting configuration for Lambda Function %s ...", name)

	decoder := json.NewDecoder(request.Body)
	defer request.Body.Close()

	var body lambda.UpdateFunctionConfigurationInput
	err := decoder.Decode(&body)
	if err != nil {
		msg := fmt.Sprintf("Error when decoding body: %v", err)
		logger.Error(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}

	logger.Infof("Configuration: %+v", body)

	ctx := request.Context()

	function, err := f.functionRepo.GetLatestFunctionByName(ctx, name)

	switch {
	case err == sql.ErrNoRows:
		logger.Infof("Unable to find Function named %s", name)
		http.NotFound(response, request)
		return
	case err != nil:
		msg := fmt.Sprintf("Error when querying for Function %s: %v", name, err)
		logger.Error(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}

	if body.Environment != nil {
		err = f.functionRepo.UpsertFunctionEnvironment(ctx, function, body.Environment)
		if err != nil {
			msg := fmt.Sprintf("Error when upserting Environment for Function %s: %v", name, err)
			logger.Error(msg)
			http.Error(response, msg, http.StatusInternalServerError)
			return
		}
	}

	result := function.ToUpdateFunctionConfigurationOutput(f.cfg)
	respondWithJson(response, result)
}

func (f FunctionHandler) GetLambdaFunction(response http.ResponseWriter, request *http.Request) {
	name := chi.URLParam(request, "name")

	logger.Infof("Getting Lambda Function %s", name)

	ctx := request.Context()

	function, err := f.functionRepo.GetLatestFunctionByName(ctx, name)
	if err != nil {
		msg := fmt.Sprintf("Unable to get Lambda Function %s: %v", name, err)
		logger.Error(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}

	layers, err := f.functionRepo.GetLayersForFunction(ctx, *function)
	if err != nil {
		msg := fmt.Sprintf("Unable to load Layers for Function %s: %v", name, err)
		logger.Error(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}

	function.Layers = layers
	result := function.ToGetFunctionOutput(f.cfg)

	respondWithJson(response, result)
}

func (f FunctionHandler) GetFunctionVersions(response http.ResponseWriter, request *http.Request) {
	name := chi.URLParam(request, "name")

	logger.Infof("Getting Versions for Lambda Function %s", name)

	ctx := request.Context()

	functions, err := f.functionRepo.GetVersionsForFunctionName(ctx, name)

	if err != nil {
		msg := fmt.Sprintf("Unable to get versions for Lambda Function %s: %v", name, err)
		logger.Error(msg)
		http.Error(response, msg, http.StatusInternalServerError)
		return
	}

	configs := make([]aws.FunctionConfiguration, len(functions))
	for i, function := range functions {
		configs[i] = *function.ToFunctionConfiguration(f.cfg)
	}

	results := lambda.ListVersionsByFunctionOutput{
		NextMarker:     nil,
		Versions:       configs,
		ResultMetadata: middleware.Metadata{},
	}

	respondWithJson(response, results)
}

func (f FunctionHandler) GetFunctionCodeSigning(response http.ResponseWriter, request *http.Request) {
	result := lambda.GetFunctionCodeSigningConfigOutput{}
	respondWithJson(response, result)
}
