package internal

import (
	"encoding/json"
	"fmt"
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

// Convert map key type to string
func FixMap(m map[any]any) map[string]any {
	out := map[string]any{}
	for k, v := range m {
		if v, ok := any(v).(map[any]any); ok {
			out[fmt.Sprint(k)] = FixMap(v)
			continue
		}
		out[fmt.Sprint(k)] = v
	}
	return out
}
