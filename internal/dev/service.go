package dev

import (
	"context"
	"fmt"
	"github.com/ATenderholt/dockerlib"
	"github.com/docker/docker/api/types/mount"
	"os"
	"path/filepath"
	"time"
)

const Requirements = "requirements.txt"

var imageMap = map[string]string{
	"python3.6":  "python:3.6-alpine",
	"python3.7":  "python:3.7-alpine",
	"python3.8":  "python:3.8-alpine",
	"python3.9":  "python:3.9-alpine",
	"python3.10": "python:3.10-alpine",
}

type Service struct {
	docker    *dockerlib.DockerController
	tempPaths map[string]string
}

func NewService(docker *dockerlib.DockerController) *Service {
	return &Service{
		docker:    docker,
		tempPaths: make(map[string]string),
	}
}

func (s *Service) InstallDependencies(ctx context.Context, runtime, basePath string) (string, error) {
	path := filepath.Join(basePath, Requirements)
	stats, err := os.Stat(path)
	switch {
	case os.IsExist(err):
		logger.Infof("Requirements file not found in %s", basePath)
		return "", nil
	case err != nil:
		e := Error{"unable to determine if Requirements file exists", err}
		logger.Error(e)
		return "", e
	}

	if stats.IsDir() {
		err := fmt.Errorf("path to requirements file (%s) is a directory", path)
		logger.Error(err)
		return "", err
	}

	temp, err := os.MkdirTemp("", "lambda-build-*")
	err = os.MkdirAll(temp, 0755)
	if err != nil {
		e := Error{"unable to make temp directory " + temp, err}
		logger.Error(e)
		return "", e
	}

	err = s.docker.EnsureImage(ctx, imageMap[runtime])
	if err != nil {
		e := Error{"unable to ensure image " + imageMap[runtime] + " exists", err}
		logger.Error(e)
		return "", e
	}

	container := dockerlib.Container{
		Name:  filepath.Base(basePath) + "_deps",
		Image: imageMap[runtime],
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   basePath,
				Target:   "/work",
				ReadOnly: true,
			},
			{
				Type:     mount.TypeBind,
				Source:   temp,
				Target:   "/build",
				ReadOnly: false,
			},
		},
		Ports:   nil,
		Command: []string{"pip", "install", "-r", "/work/requirements.txt", "-t", "/build"},
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	ready, err := s.docker.Start(timeoutCtx, &container, "Successfully installed")
	if err != nil {
		e := Error{"unable to start container to install dependencies for " + basePath, err}
		logger.Error(e)
		return "", e
	}

	// wait until container shows Successfully installed
	<-ready

	// wait until container actually is shut down since it doesn't seem to be immediate
	err = s.docker.WaitForShutdown(ctx, container, 10*time.Second)
	if err != nil {
		logger.Warnf("Unable to wait for %s to shutdown: %v", container.Name, err)
	}

	err = s.docker.Remove(ctx, container)
	if err != nil {
		logger.Warnf("Unable to remove %s: %v", container.Name, err)
	}

	return temp, nil
}
