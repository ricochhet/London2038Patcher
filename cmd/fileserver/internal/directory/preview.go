package directory

import (
	"fmt"
	"net/http"
	"os"
)

// handlePreview serves a file with Content-Disposition: inline so the browser can display it.
func handlePreview(w http.ResponseWriter, r *http.Request, abs string, stat os.FileInfo) {
	if stat.IsDir() {
		http.Error(w, "Cannot preview a directory", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename=%q`, stat.Name()))
	http.ServeFile(w, r, abs)
}
