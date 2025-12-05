package fsutil

import (
	"os"
	"path/filepath"

	"github.com/ricochhet/london2038patcher/pkg/errutil"
)

// Read reads a file from the specified path.
func Read(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return data, nil
}

// Write writes to the specified path with the provided data.
func Write(path string, data []byte) error {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return errutil.WithFrame(err)
	}

	err = os.WriteFile(path, data, 0o644)
	if err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}
