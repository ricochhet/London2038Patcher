package jsonutil

import (
	"encoding/json"
	"os"

	"github.com/ricochhet/london2038patcher/pkg/errutil"
)

// Unmarshal parses a JSON file from the specified path.
func Unmarshal[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	var t T
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, errutil.WithFrame(err)
	}

	return &t, nil
}
