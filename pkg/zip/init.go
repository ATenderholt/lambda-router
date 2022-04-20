package zip

import (
	"github.com/ATenderholt/rainbow-functions/logging"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func init() {
	logger = logging.NewLogger().Named("zip")
}
