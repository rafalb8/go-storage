package internal

import (
	"encoding/json"
	"os"

	"go.uber.org/zap"
)

var (
	logger, _ = zap.NewProduction()
)

func Logger() *zap.SugaredLogger {
	return logger.Sugar()
}

func PrintJSON(v interface{}) {
	out, _ := json.MarshalIndent(v, "", "  ")
	logger.Info(string(out))
}

func Must[T any](v T, err error) T {
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	return v
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
