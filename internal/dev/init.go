package dev

import (
	"github.com/ATenderholt/lambda-router/logging"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func init() {
	logger = logging.NewLogger().Named("dev")
}
