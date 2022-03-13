package logging

import "go.uber.org/zap"

func NewLogger() *zap.SugaredLogger {
	t, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	return t.Sugar()
}
