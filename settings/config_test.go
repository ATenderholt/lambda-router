package settings_test

import (
	"github.com/ATenderholt/rainbow-functions/settings"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg, output, err := settings.FromFlags("lambda-router", []string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.Empty(t, output)

	expected := settings.DefaultConfig()
	assert.Equal(t, cfg, expected)
}

func TestSetDebug(t *testing.T) {
	cfg, output, err := settings.FromFlags("lambda-router", []string{"-debug"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.Empty(t, output)

	expected := settings.DefaultConfig()
	expected.IsDebug = true
	assert.Equal(t, cfg, expected)
}

func TestSetContainer(t *testing.T) {
	cfg, output, err := settings.FromFlags("lambda-router", []string{"-local=false"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.Empty(t, output)

	expected := settings.DefaultConfig()
	expected.IsLocal = false
	assert.Equal(t, cfg, expected)
}

func TestSetPortSqsAndDevFile(t *testing.T) {
	cfg, output, err := settings.FromFlags("lambda-router", []string{
		"-port", "8000",
		"-sqs-endpoint", "http://sqs",
		"-config", "testdata/functions.yml",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.Empty(t, output)

	expected := settings.DefaultConfig()
	expected.BasePort = 8000
	expected.SqsEndpoint = "http://sqs"
	expected.DevConfigFile = "testdata/functions.yml"
	assert.Equal(t, cfg, expected)
}

func TestSetNetworks(t *testing.T) {
	cfg, output, err := settings.FromFlags("lambda-router", []string{
		"-networks", "sqs,s3,lambda",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	assert.Empty(t, output)

	expected := settings.DefaultConfig()
	expected.Networks = []string{"sqs", "s3", "lambda"}
	assert.Equal(t, cfg, expected)
}
