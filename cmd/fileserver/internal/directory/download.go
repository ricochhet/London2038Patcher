package directory

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ricochhet/london2038patcher/pkg/logutil"
)

// handleDownload handles downloading of files and directories.
func handleDownload(w http.ResponseWriter, r *http.Request, root string, stat os.FileInfo) {
	if !stat.IsDir() {
		w.Header().Set(
			"Content-Disposition",
			fmt.Sprintf(`attachment; filename=%q`, stat.Name()),
		)

		f, err := os.Open(root)
		if err != nil {
			http.Error(w, "Could not open file", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)

		return
	}

	name := stat.Name() + ".zip"

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename=%q`, name),
	)

	zw := zip.NewWriter(w)
	defer zw.Close()

	err := filepath.Walk(root, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		rel, err := filepath.Rel(root, walkPath)
		if err != nil {
			return err
		}

		fh, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		fh.Name = filepath.ToSlash(rel)
		fh.Method = zip.Deflate

		fw, err := zw.CreateHeader(fh)
		if err != nil {
			return err
		}

		f, err := os.Open(walkPath)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(fw, f)

		return err
	})
	if err != nil {
		logutil.Errorf(logutil.Get(), "handleDownload zip walk: %v\n", err)
	}
}
