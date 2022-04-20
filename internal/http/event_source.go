package http

import (
	"encoding/json"
	"fmt"
	"github.com/ATenderholt/rainbow-functions/internal/domain"
	"github.com/ATenderholt/rainbow-functions/settings"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
	"time"
)

type EventSourceHandler struct {
	cfg          *settings.Config
	eventRepo    domain.EventSourceRepository
	functionRepo domain.FunctionRepository
}

func NewEventSourceHandler(cfg *settings.Config, eventRepo domain.EventSourceRepository, functionRepo domain.FunctionRepository) EventSourceHandler {
	return EventSourceHandler{
		cfg:          cfg,
		eventRepo:    eventRepo,
		functionRepo: functionRepo,
	}
}

func (e EventSourceHandler) PostEventSource(writer http.ResponseWriter, request *http.Request) {
	var requestBodyBuilder strings.Builder
	reader := io.TeeReader(request.Body, &requestBodyBuilder)

	var payload lambda.CreateEventSourceMappingInput
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&payload)
	if err != nil {
		msg := fmt.Sprintf("unable to decode body for creating an Event Source: %v", err)
		logger.Error(msg)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	ctx := request.Context()

	function, err := e.functionRepo.GetLatestFunctionByName(ctx, *payload.FunctionName)
	if err != nil {
		msg := fmt.Sprintf("unable to load Function %s: %v", *payload.FunctionName, err)
		logger.Error(msg)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	eventSource := domain.EventSource{
		UUID:         uuid.New(),
		Enabled:      true,
		Arn:          *payload.EventSourceArn,
		Function:     function,
		BatchSize:    *payload.BatchSize,
		LastModified: time.Now().UnixMilli(),
	}

	logger.Infof("Saving Event Source: %+v", eventSource)

	err = e.eventRepo.InsertEventSource(ctx, eventSource)
	if err != nil {
		msg := fmt.Sprintf("unable to save Event Source %+v: %v", eventSource, err)
		logger.Error(msg)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	body := eventSource.ToCreateEventSourceMappingOutput(e.cfg)

	respondWithJson(writer, body)
}

func (e EventSourceHandler) GetEventSource(writer http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")

	logger.Infof("Getting event source %s", id)

	ctx := request.Context()

	eventSource, err := e.eventRepo.GetEventSource(ctx, id)
	if err != nil {
		msg := fmt.Sprintf("Unable to load Event Source %+v: %v", eventSource, err)
		logger.Error(msg)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	if eventSource == nil {
		logger.Infof("Event Source %s not found", id)
		http.NotFound(writer, request)
		return
	}

	body := eventSource.ToGetEventSourceMappingOutput(e.cfg)

	respondWithJson(writer, body)
}
