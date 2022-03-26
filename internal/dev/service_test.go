package dev_test

import (
	"context"
	"fmt"
	"github.com/ATenderholt/dockerlib"
	"github.com/ATenderholt/lambda-router/internal/dev"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallDependencies(t *testing.T) {
	docker, err := dockerlib.NewDockerController()
	if err != nil {
		t.Errorf("Unable to get dockerlib controller: %v", err)
		t.FailNow()
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Unable to get current directory: %v", err)
		t.FailNow()
	}

	service := dev.NewService(docker)
	dir, err := service.InstallDependencies(context.Background(), "python3.8", filepath.Join(cwd, "testdata"))
	if err != nil {
		t.Errorf("Unable to install dependencies: %v", err)
		t.Fail()
	}

	fmt.Printf("Deps installed to %s", dir)
}
