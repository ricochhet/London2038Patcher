package directory

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ricochhet/london2038patcher/pkg/embedutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
)

const maxContentSearchSize = 10 << 20 // 10 MB

var defaultImageExts = []string{
	".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp", ".ico",
}

var defaultTextExts = []string{
	".txt", ".md", ".json", ".yaml", ".yml", ".toml", ".xml",
	".html", ".htm", ".js", ".ts", ".css", ".go", ".py", ".rs",
	".java", ".c", ".cpp", ".h", ".hpp", ".sh", ".bash", ".zsh",
	".env", ".log", ".csv", ".ini", ".conf", ".cfg", ".rb", ".php",
	".sql", ".tf", ".hcl", ".lua", ".vim", ".diff", ".patch",
}

var defaultReadmeCandidates = []string{"README.md", "INDEX.md", "index.md"}

type fileInfoResponse struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	FullPath    string    `json:"fullPath"`
	Extension   string    `json:"extension,omitempty"`
	MimeType    string    `json:"mimeType,omitempty"`
	Size        int64     `json:"size"`
	Modified    time.Time `json:"modified"`
	IsDirectory bool      `json:"isDirectory"`
	MD5         string    `json:"md5,omitempty"`
}

type searchResult struct {
	Name         string `json:"name"`
	RelPath      string `json:"relPath"`
	HighlightURL string `json:"highlightUrl"`
	DownloadURL  string `json:"downloadUrl"`
	MatchType    string `json:"matchType"`
	Snippet      string `json:"snippet,omitempty"`
}

type breadcrumb struct {
	Name   string
	Link   string
	IsLast bool
}

type dirEntry struct {
	Name        string `json:"name"`
	IsDir       bool   `json:"isDir"`
	SizeStr     string `json:"sizeStr"`
	SizeBytes   int64  `json:"sizeBytes"`
	ModStr      string `json:"modStr"`
	ModUnix     int64  `json:"modUnix"`
	BrowseURL   string `json:"browseUrl"`
	DownloadURL string `json:"downloadUrl"`
	InfoURL     string `json:"infoUrl"`
	PreviewURL  string `json:"previewUrl"`
	Ext         string `json:"ext"`
}

type dirTemplateData struct {
	Title       string
	Breadcrumbs []breadcrumb
	Parent      string
	Entries     []dirEntry
	EntriesJSON template.JS
	IsEmpty     bool
	Readme      string
	HasReadme   bool
	Route       string
	FileCount   int
	TotalSize   string

	ImageExtsJSON template.JS
	TextExtsJSON  template.JS
}

// BrowseHandler is a handler that supplies a file browser.
func BrowseHandler(
	fs *embedutil.EmbeddedFileSystem,
	dirPath, route string,
	hidden []string,
	imageExts, textExts, readmeCandidates []string,
) http.Handler {
	route = strings.TrimSuffix(route, "/")

	if len(imageExts) == 0 {
		imageExts = defaultImageExts
	}

	if len(textExts) == 0 {
		textExts = defaultTextExts
	}

	if len(readmeCandidates) == 0 {
		readmeCandidates = defaultReadmeCandidates
	}

	imageExtsJSON := extSliceToJSObject(imageExts)
	textExtsJSON := extSliceToJSObject(textExts)

	absBase, err := filepath.Abs(dirPath)
	if err != nil {
		panic(fmt.Sprintf("DirectoryBrowseHandler: cannot resolve basePath %q: %v", dirPath, err))
	}

	bytes := embedutil.MaybeRead(fs, "directory.html")
	tmpl := template.Must(template.New("dir").Parse(string(bytes)))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("search") {
			if err := handleSearch(w, r, absBase, route, hidden); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}

			return
		}

		sub := filepath.FromSlash(chi.URLParam(r, "*"))

		abs, err := fsutil.SafeJoin(absBase, sub)
		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		stat, err := os.Stat(abs)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}

			return
		}

		switch {
		case r.URL.Query().Has("preview"):
			handlePreview(w, r, abs, stat)
		case r.URL.Query().Has("download"):
			handleDownload(w, r, abs, stat)
		case r.URL.Query().Has("info"):
			handleInfo(w, r, abs, absBase, stat)
		case stat.IsDir():
			handleListing(
				w, r, tmpl,
				abs, route, filepath.ToSlash(sub), route,
				hidden,
				imageExtsJSON, textExtsJSON,
				readmeCandidates,
			)
		default:
			http.ServeFile(w, r, abs)
		}
	})
}
