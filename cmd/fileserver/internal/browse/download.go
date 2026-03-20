package browse

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/httputil"
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
			errutil.HTTPInternalServerErrorf(w, "Could not open file: %v\n", err)
			return
		}
		defer f.Close()

		http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)

		return
	}

	name := stat.Name() + ".zip"

	httputil.ContentType(w, httputil.ContentTypeZip)
	w.Header().Set(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename=%q`, name),
	)

	if err := writeZipArchive(w, root); err != nil {
		logutil.Errorf(logutil.Get(), "handleDownload zip walk: %v\n", err)
	}
}
