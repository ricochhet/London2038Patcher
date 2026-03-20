package server

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ricochhet/london2038patcher/pkg/cryptoutil"
	"github.com/ricochhet/london2038patcher/pkg/embedutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
	"github.com/ricochhet/london2038patcher/pkg/strutil"
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
	HighlightURL string `json:"highlightURL"`
	DownloadURL  string `json:"downloadURL"`
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
	BrowseURL   string `json:"browseURL"`
	DownloadURL string `json:"downloadURL"`
	InfoURL     string `json:"infoURL"`
	PreviewURL  string `json:"previewURL"`
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

// extSliceToJSObject converts a slice of file extensions into a JSON object
// literal suitable for direct injection into a <script> block, e.g.:
//
//	{".jpg":1,".png":1}
//
// This lets the browser use O(1) property lookups instead of Array.includes.
func extSliceToJSObject(exts []string) template.JS {
	m := make(map[string]int, len(exts))
	for _, e := range exts {
		m[e] = 1
	}

	b, err := json.Marshal(m)
	if err != nil {
		return template.JS("{}")
	}

	return template.JS(b)
}

// DirectoryBrowseHandler is a handler that supplies a file browser.
func DirectoryBrowseHandler(
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

	bytes := maybeRead(fs, "directory.html")
	tmpl := template.Must(template.New("dir").Parse(string(bytes)))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("search") {
			handleSearch(w, r, absBase, route, hidden)
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
			handleDirListing(
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

// handleDirListing handles creating and listing files and directories in the browser.
func handleDirListing(
	w http.ResponseWriter,
	_ *http.Request,
	tmpl *template.Template,
	absPath, route, subPath string,
	browseRoute string,
	hidden []string,
	imageExtsJSON, textExtsJSON template.JS,
	readmeCandidates []string,
) {
	rawEntries, err := os.ReadDir(absPath)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	filtered := rawEntries[:0]
	for _, e := range rawEntries {
		if !isHidden(e.Name(), hidden) {
			filtered = append(filtered, e)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		di, dj := filtered[i].IsDir(), filtered[j].IsDir()
		if di != dj {
			return di
		}

		return strings.ToLower(filtered[i].Name()) < strings.ToLower(filtered[j].Name())
	})

	trimmed := strings.Trim(subPath, "/")

	var totalBytes int64
	fileCount := 0

	entries := make([]dirEntry, 0, len(filtered))
	for _, e := range filtered {
		info, err := e.Info()
		if err != nil {
			continue
		}

		var entryURL string
		if trimmed == "" {
			entryURL = route + "/" + e.Name()
		} else {
			entryURL = route + "/" + trimmed + "/" + e.Name()
		}

		sizeStr := "—"
		var sizeBytes int64
		if !e.IsDir() {
			sizeBytes = info.Size()
			sizeStr = strutil.Size(sizeBytes)
			totalBytes += sizeBytes
			fileCount++
		}

		ext := ""
		previewURL := ""
		if !e.IsDir() {
			ext = strings.ToLower(filepath.Ext(e.Name()))
			previewURL = entryURL + "?preview"
		}

		entries = append(entries, dirEntry{
			Name:        e.Name(),
			IsDir:       e.IsDir(),
			SizeStr:     sizeStr,
			SizeBytes:   sizeBytes,
			ModStr:      info.ModTime().Format("2006-01-02  15:04"),
			ModUnix:     info.ModTime().Unix(),
			BrowseURL:   entryURL,
			DownloadURL: entryURL + "?download",
			InfoURL:     entryURL + "?info",
			PreviewURL:  previewURL,
			Ext:         ext,
		})
	}

	entriesRaw, err := json.Marshal(entries)
	if err != nil {
		entriesRaw = []byte("[]")
	}

	parent := ""
	if trimmed != "" {
		up := path.Dir("/" + trimmed)
		if up == "/" {
			parent = route + "/"
		} else {
			parent = route + up
		}
	}

	title := "/"
	if trimmed != "" {
		title = trimmed
	}

	readmeContent := ""
	for _, candidate := range readmeCandidates {
		if b, err := os.ReadFile(filepath.Join(absPath, candidate)); err == nil {
			readmeContent = string(b)
			break
		}
	}

	totalSizeStr := ""
	if fileCount > 0 {
		totalSizeStr = strutil.Size(totalBytes)
	}

	data := dirTemplateData{
		Title:         title,
		Breadcrumbs:   buildBreadcrumbs(route, trimmed),
		Parent:        parent,
		Entries:       entries,
		EntriesJSON:   template.JS(entriesRaw),
		IsEmpty:       len(entries) == 0,
		Readme:        readmeContent,
		HasReadme:     readmeContent != "",
		Route:         browseRoute,
		FileCount:     fileCount,
		TotalSize:     totalSizeStr,
		ImageExtsJSON: imageExtsJSON,
		TextExtsJSON:  textExtsJSON,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := tmpl.Execute(w, data); err != nil {
		logutil.Errorf(logutil.Get(), "dirTempl.Execute: %v\n", err)
	}
}

// buildBreadcrumbs handles building of file browser paths.
func buildBreadcrumbs(route, path string) []breadcrumb {
	root := breadcrumb{Name: "~", Link: route + "/"}

	if path == "" {
		root.IsLast = true
		return []breadcrumb{root}
	}

	parts := strings.Split(path, "/")
	crumbs := []breadcrumb{root}

	for i, part := range parts {
		isLast := i == len(parts)-1

		link := ""
		if !isLast {
			link = route + "/" + strings.Join(parts[:i+1], "/")
		}

		crumbs = append(crumbs, breadcrumb{Name: part, Link: link, IsLast: isLast})
	}

	return crumbs
}

// handlePreview serves a file with Content-Disposition: inline so the browser can display it.
func handlePreview(w http.ResponseWriter, r *http.Request, abs string, stat os.FileInfo) {
	if stat.IsDir() {
		http.Error(w, "Cannot preview a directory", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename=%q`, stat.Name()))
	http.ServeFile(w, r, abs)
}

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

// handleSearch walks abs and returns JSON results matching the search query.
//   - ?search=q              — filename match
//   - ?search=q&content=1    — also search inside file contents (text files ≤ 10 MB)
//   - ext:dat or ext:.dat    — restrict results to files with that extension
//   - extension:dat          — alias for ext:
func handleSearch(w http.ResponseWriter, r *http.Request, abs, route string, hidden []string) {
	raw := strings.TrimSpace(r.URL.Query().Get("search"))
	contentSearch := r.URL.Query().Has("content")

	baseQuery, extFilter := parseSearchQuery(raw)

	var results []searchResult

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if baseQuery == "" && extFilter == "" {
		_ = json.NewEncoder(w).Encode(results)
		return
	}

	_ = filepath.Walk(abs, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if isHidden(info.Name(), hidden) {
			if info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		if extFilter != "" && ext != extFilter {
			return nil
		}

		nameMatch := baseQuery == "" || strings.Contains(strings.ToLower(info.Name()), baseQuery)

		rel, err := filepath.Rel(abs, walkPath)
		if err != nil {
			return nil
		}

		rel = filepath.ToSlash(rel)
		dir := path.Dir(rel)

		var dirURL string
		if dir == "." {
			dirURL = route + "/"
		} else {
			dirURL = route + "/" + dir
		}

		if nameMatch {
			results = append(results, searchResult{
				Name:         info.Name(),
				RelPath:      rel,
				HighlightURL: dirURL + "?highlight=" + url.QueryEscape(info.Name()),
				DownloadURL:  route + "/" + rel + "?download",
				MatchType:    "name",
			})

			return nil
		}

		if contentSearch && baseQuery != "" {
			if snippet, ok := searchFileContent(walkPath, info, baseQuery); ok {
				results = append(results, searchResult{
					Name:         info.Name(),
					RelPath:      rel,
					HighlightURL: dirURL + "?highlight=" + url.QueryEscape(info.Name()),
					DownloadURL:  route + "/" + rel + "?download",
					MatchType:    "content",
					Snippet:      snippet,
				})
			}
		}

		return nil
	})

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(results)
}

// searchFileContent checks whether query appears inside the given file.
func searchFileContent(filePath string, info os.FileInfo, query string) (string, bool) {
	if info.Size() > maxContentSearchSize {
		return "", false
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", false
	}

	ct := http.DetectContentType(data)
	if !strings.HasPrefix(ct, "text/") &&
		!strings.Contains(ct, "json") &&
		!strings.Contains(ct, "xml") {
		return "", false
	}

	lower := strings.ToLower(string(data))
	idx := strings.Index(lower, query)
	if idx < 0 {
		return "", false
	}

	start := idx - 60
	if start < 0 {
		start = 0
	}

	end := idx + len(query) + 60
	if end > len(data) {
		end = len(data)
	}

	snippet := "…" + strings.TrimSpace(string(data[start:end])) + "…"

	return snippet, true
}

// parseSearchQuery splits a raw search string into a base query and an optional
// extension filter. Tokens of the form "ext:VALUE" or "extension:VALUE" (case-
// insensitive) are extracted as an extension filter; all remaining tokens form
// the base query. The extension is normalised so that both "dat" and ".dat"
// resolve to ".dat".
//
// Bare words "ext" or "extension" without a trailing colon are left in the
// base query unchanged, so you can still search for files containing those words.
func parseSearchQuery(raw string) (baseQuery, extFilter string) {
	tokens := strings.Fields(raw)
	rest := tokens[:0]

	for _, t := range tokens {
		lower := strings.ToLower(t)

		var val string
		var isTag bool

		if after, ok := strings.CutPrefix(lower, "extension:"); ok && after != "" {
			val, isTag = after, true
		} else if after, ok := strings.CutPrefix(lower, "ext:"); ok && after != "" {
			val, isTag = after, true
		}

		if isTag {
			if !strings.HasPrefix(val, ".") {
				val = "." + val
			}

			extFilter = val

			continue
		}

		rest = append(rest, t)
	}

	baseQuery = strings.ToLower(strings.Join(rest, " "))

	return baseQuery, extFilter
}

// isHidden reports whether name matches any of the given glob patterns.
func isHidden(name string, patterns []string) bool {
	for _, p := range patterns {
		matched, err := filepath.Match(p, name)
		if err == nil && matched {
			return true
		}
	}

	return false
}
