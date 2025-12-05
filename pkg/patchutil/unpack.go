package patchutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ricochhet/london2038patcher/pkg/errutil"
)

// Unpack unpacks the specified path with the provided index.
func (idx *Index) Unpack(path, output string) error {
	f, err := os.Open(path)
	if err != nil {
		return errutil.WithFrame(err)
	}
	defer f.Close()

	err = os.MkdirAll(output, 0o755)
	if err != nil {
		return errutil.WithFrame(err)
	}

	for _, entry := range idx.Files {
		if entry.FileSize <= 0 {
			continue
		}

		target := filepath.Join(output, filepath.FromSlash(entry.FileName))

		err = os.MkdirAll(filepath.Dir(target), 0o755)
		if err != nil {
			return errutil.WithFrame(err)
		}

		fmt.Fprintf(os.Stdout, "Extracting: %s (%d bytes)\n", entry.FileName, entry.FileSize)

		_, err = f.Seek(entry.DatOffset, io.SeekStart)
		if err != nil {
			return errutil.WithFrame(err)
		}

		buf := make([]byte, entry.FileSize)

		_, err = io.ReadFull(f, buf)
		if err != nil {
			return errutil.WithFrame(err)
		}

		err = os.WriteFile(target, buf, 0o644)
		if err != nil {
			return errutil.WithFrame(err)
		}
	}

	return nil
}
