package docker

import (
	"context"
	"errors"
	"fmt"
	"github.com/ATenderholt/dockerlib"
	"github.com/ATenderholt/lambda-router/internal/domain"
	"github.com/ATenderholt/lambda-router/settings"
	"github.com/docker/docker/api/types/mount"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

// Manager is responsible for launching Docker containers hosting Lambda functions & their invocation
type Manager struct {
	cfg *settings.Config

	// pool of ports available for use
	ports IntPool

	// map of running lambdas (name) and their hostname:port
	running map[string]string

	docker dockerlib.Controller
}

func NewManager(cfg *settings.Config) (*Manager, error) {
	ports := NewIntPool(cfg.BasePort+1, cfg.BasePort+51)
	running := make(map[string]string)
	docker, err := dockerlib.NewDockerController()
	if err != nil {
		return nil, err
	}

	return &Manager{
		cfg:     cfg,
		docker:  *docker,
		ports:   ports,
		running: running,
	}, nil
}

func (m Manager) StartFunction(ctx context.Context, function *domain.Function) error {
	port, err := m.ports.Get(ctx)
	if err != nil {
		msg := fmt.Sprintf("Unable to start Function %s: %v", function.FunctionName, err)
		logger.Error(msg)
		return errors.New(msg)
	}

	logger.Infof("Starting Function %s on port %d using handler %s", function.FunctionName, port, function.Handler)

	container := dockerlib.Container{
		Name:    function.FunctionName,
		Image:   "mlupin/docker-lambda:" + string(function.Runtime),
		Command: []string{function.Handler},
		Mounts: []mount.Mount{
			{
				Source:      function.GetDestPath(m.cfg),
				Target:      "/var/task",
				Type:        mount.TypeBind,
				ReadOnly:    true,
				Consistency: mount.ConsistencyDelegated,
			},
			{
				Source:      function.GetLayerDestPath(m.cfg),
				Target:      "/opt",
				Type:        mount.TypeBind,
				ReadOnly:    true,
				Consistency: mount.ConsistencyDelegated,
			},
		},
		Environment: []string{
			"DOCKER_LAMBDA_STAY_OPEN=1",
			"DOCKER_LAMBDA_WATCH=1",
		},
		Ports: map[int]int{
			9001: port,
		},
	}

	_, err = m.docker.Start(ctx, container, "")
	if err != nil {
		msg := fmt.Sprintf("Unable to start Function %s: %v", function.FunctionName, err)
		logger.Error(msg)
		return errors.New(msg)
	}

	var uri string
	if m.cfg.IsLocal {
		uri = fmt.Sprintf("http://localhost:%d", port)
	} else {
		uri = fmt.Sprintf("http://%s:9001", function.FunctionName)
	}

	m.running[function.FunctionName] = uri

	return nil
}

func (m Manager) Invoke(writer http.ResponseWriter, request *http.Request) {
	name := chi.URLParam(request, "name")
	logger.Infof("Invoking Function %s", name)

	host, ok := m.running[name]
	if !ok {
		msg := fmt.Sprintf("Function %s is not running", name)
		logger.Errorf(msg)
		http.Error(writer, msg, http.StatusNotFound)
		return
	}

	proxyReq, _ := http.NewRequest(request.Method, host+request.URL.Path, request.Body)

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		msg := fmt.Sprintf("Unable to invoke Function %s: %v", name, err)
		logger.Error(msg)
		http.Error(writer, msg, http.StatusInternalServerError)
		return
	}

	logger.Debugf("Got following response when invoking Function %s: %+v", name, resp)

	for key, value := range resp.Header {
		for _, v := range value {
			writer.Header().Add(key, v)
		}

	}

	writer.WriteHeader(resp.StatusCode)

	io.Copy(writer, resp.Body)
	resp.Body.Close()
}
