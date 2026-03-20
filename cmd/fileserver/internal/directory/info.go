package directory

import (
	"encoding/json"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ricochhet/london2038patcher/pkg/cryptoutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
)

// handleInfo handles per-file information.
func handleInfo(
	w http.ResponseWriter,
	_ *http.Request,
	filePath, base string,
	stat os.FileInfo,
) {
	rel, _ := filepath.Rel(base, filePath)
	rel = filepath.ToSlash(rel)

	res := fileInfoResponse{
		Name:        stat.Name(),
		Path:        "/" + rel,
		FullPath:    filePath,
		Size:        stat.Size(),
		Modified:    stat.ModTime().UTC(),
		IsDirectory: stat.IsDir(),
	}

	if !stat.IsDir() {
		ext := filepath.Ext(stat.Name())
		res.Extension = ext
		res.MimeType = mime.TypeByExtension(ext)

		if hash, err := cryptoutil.MD5(filePath); err != nil {
			logutil.Errorf(logutil.Get(), "handleInfo md5 %q: %v\n", filePath, err)
		} else {
			res.MD5 = hash
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(res); err != nil {
		logutil.Errorf(logutil.Get(), "handleInfo encode: %v\n", err)
	}
}
