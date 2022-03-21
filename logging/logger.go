package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() *zap.SugaredLogger {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	t, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return t.Sugar()
}

type GooseLogger struct {
	Logger *zap.SugaredLogger
}

func (g GooseLogger) Fatal(v ...interface{}) {
	g.Logger.Fatal(v...)
}

func (g GooseLogger) Fatalf(format string, v ...interface{}) {
	g.Logger.Fatalf(format, v...)
}

func (g GooseLogger) Print(v ...interface{}) {
	g.Logger.Info(v...)
}

func (g GooseLogger) Println(v ...interface{}) {
	g.Logger.Info(v...)
}

func (g GooseLogger) Printf(format string, v ...interface{}) {
	g.Logger.Infof(format, v...)
}
