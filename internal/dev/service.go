package dev

import (
	"context"
	"fmt"
	"github.com/ATenderholt/dockerlib"
	"github.com/ATenderholt/rainbow-functions/settings"
	"github.com/docker/docker/api/types/mount"
	"os"
	"path/filepath"
	"strings"
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
	cfg       *settings.Config
	docker    *dockerlib.DockerController
	tempPaths map[string]string
}

func NewService(cfg *settings.Config, docker *dockerlib.DockerController) *Service {
	return &Service{
		cfg:       cfg,
		docker:    docker,
		tempPaths: make(map[string]string),
	}
}

func mkTempDir() (string, error) {
	temp, err := os.MkdirTemp("", "lambda-build-*")
	if err != nil {
		e := Error{"unable to make temp directory " + temp, err}
		logger.Error(e)
		return "", e
	}

	// Mac returns things in /var which is symlinked to /private/var
	// Only /private/var seems exposed in Docker Desktop
	temp2, err := filepath.EvalSymlinks(temp)
	if err != nil {
		e := Error{"unable to resolve symlinks for temp directory " + temp, err}
		logger.Error(e)
		return temp, e
	}

	return temp2, nil
}

func (s *Service) InstallDependencies(ctx context.Context, runtime, basePath string) (string, error) {
	name := filepath.Base(basePath)
	temp, err := mkTempDir()
	if err != nil {
		return "", err
	}
	s.tempPaths[name] = temp

	path := filepath.Join(basePath, Requirements)
	stats, err := os.Stat(path)
	switch {
	case os.IsNotExist(err):
		logger.Infof("Requirements file not found in %s", basePath)
		return temp, nil
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

	err = s.docker.EnsureImage(ctx, imageMap[runtime])
	if err != nil {
		e := Error{"unable to ensure image " + imageMap[runtime] + " exists", err}
		logger.Error(e)
		return "", e
	}

	containerPath := basePath
	if !s.cfg.IsLocal {
		containerName := os.Getenv("NAME")
		logger.Infof("Getting source for mount %s in container %s", s.cfg.DataPath(), containerName)
		hostPath, err := s.docker.GetContainerHostPath(ctx, containerName, s.cfg.DataPath())

		if err != nil {
			e := fmt.Errorf("unable to get host path for %s: %v", s.cfg.DataPath(), err)
			logger.Error(e)
			return "", e
		}

		containerPath = strings.Replace(containerPath, s.cfg.DataPath(), hostPath, 1)
	}

	container := dockerlib.Container{
		Name:  name + "_deps",
		Image: imageMap[runtime],
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   containerPath,
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
		Command: []string{"pip", "install", "-r", "/work/requirements.txt", "-t", "/build/python"},
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

func (s *Service) CleanupAll() {
	logger.Info("Cleaning up all temporary directories")
	for _, path := range s.tempPaths {
		err := os.RemoveAll(path)
		if err != nil {
			logger.Warnf("Unable to remove %s: %v", path, err)
		}
	}
}
