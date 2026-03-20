package browse

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ricochhet/london2038patcher/pkg/errutil"
)

// handlePreview serves a file with Content-Disposition: inline so the browser can display it.
func handlePreview(w http.ResponseWriter, r *http.Request, abs string, stat os.FileInfo) {
	if stat.IsDir() {
		errutil.HTTPBadRequestf(w, "Cannot preview a directory")
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename=%q`, stat.Name()))
	http.ServeFile(w, r, abs)
}
