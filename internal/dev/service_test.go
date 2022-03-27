package dev_test

import (
	"context"
	"github.com/ATenderholt/dockerlib"
	"github.com/ATenderholt/lambda-router/internal/dev"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallDependencies(t *testing.T) {
	docker, err := dockerlib.NewDockerController()
	if err != nil {
		t.Fatalf("Unable to get dockerlib controller: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to get current directory: %v", err)
	}

	service := dev.NewService(docker)
	dir, err := service.InstallDependencies(context.Background(), "python3.8", filepath.Join(cwd, "testdata"))
	if err != nil {
		t.Fatalf("Unable to install dependencies: %v", err)
	}
	t.Cleanup(func() {
		err := os.RemoveAll(dir)
		if err != nil {
			t.Logf("Unable to remove %s: %v", dir, err)
		}
	})

	expected := filepath.Join(dir, "requests-2.27.1.dist-info", "METADATA")
	_, err = os.Stat(expected)
	if err != nil {
		t.Fatalf("Unable to find expected installed files: %v", err)
	}

}
