package domain_test

import (
	"github.com/ATenderholt/rainbow-functions/internal/domain"
	"github.com/ATenderholt/rainbow-functions/settings"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func stringEndsWith(s string, suffix string) assert.Comparison {
	return func() bool {
		return strings.HasSuffix(s, suffix)
	}
}

func TestLocalFunctionDestPaths(t *testing.T) {
	cfg := settings.DefaultConfig()
	f := domain.Function{
		FunctionName: "test-function",
		Version:      "1.2.3",
	}

	destPath := f.GetDestPath(cfg)

	assert.Condition(t, stringEndsWith(destPath, "data/lambda/functions/test-function/1.2.3/content"))
}

func TestLocalLayerDestPaths(t *testing.T) {
	cfg := settings.DefaultConfig()
	f := domain.Function{
		FunctionName: "test-function",
		Version:      "1.2.3",
	}

	destPath := f.GetLayerDestPath(cfg)

	assert.Condition(t, stringEndsWith(destPath, "data/lambda/functions/test-function/1.2.3/layers"))
}
