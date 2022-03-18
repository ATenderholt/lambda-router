package domain

import (
	"context"
	"net/http"
)

type DockerManager interface {
	StartFunction(ctx context.Context, function *Function) error
	Invoke(writer http.ResponseWriter, request *http.Request)
}
