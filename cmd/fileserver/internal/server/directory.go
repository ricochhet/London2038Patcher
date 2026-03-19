package server

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime"
	"net/http"
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

type FileInfoResponse struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Extension   string    `json:"extension,omitempty"`
	MimeType    string    `json:"mimeType,omitempty"`
	Size        int64     `json:"size"`
	Modified    time.Time `json:"modified"`
	IsDirectory bool      `json:"isDirectory"`
	MD5         string    `json:"md5,omitempty"`
}

type breadcrumb struct {
	Name   string
	Link   string
	IsLast bool
}

type dirEntry struct {
	Name        string
	IsDir       bool
	SizeStr     string
	ModStr      string
	BrowseURL   string
	DownloadURL string
	InfoURL     string
}

type dirTemplateData struct {
	Title       string
	Breadcrumbs []breadcrumb
	Parent      string
	Entries     []dirEntry
	IsEmpty     bool
	Readme      string
}

// DirectoryBrowseHandler is a handler that supplies a file browser.
func DirectoryBrowseHandler(
	fs *embedutil.EmbeddedFileSystem,
	path, route string,
) http.Handler {
	route = strings.TrimSuffix(route, "/")

	absBase, err := filepath.Abs(path)
	if err != nil {
		panic(fmt.Sprintf("DirectoryBrowseHandler: cannot resolve basePath %q: %v", path, err))
	}

	bytes := maybeRead(fs, "directory.html")
	tmpl := template.Must(template.New("dir").Parse(string(bytes)))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		case r.URL.Query().Has("download"):
			handleDownload(w, r, abs, stat)
		case r.URL.Query().Has("info"):
			handleInfo(w, r, abs, absBase, stat)
		case stat.IsDir():
			handleDirListing(w, r, tmpl, abs, route, filepath.ToSlash(sub))
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
	absPath, baseRoute, subPath string,
) {
	dirEntries, err := os.ReadDir(absPath)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	sort.Slice(dirEntries, func(i, j int) bool {
		di, dj := dirEntries[i].IsDir(), dirEntries[j].IsDir()
		if di != dj {
			return di
		}

		return strings.ToLower(dirEntries[i].Name()) < strings.ToLower(dirEntries[j].Name())
	})

	trimmed := strings.Trim(subPath, "/")

	entries := make([]dirEntry, 0, len(dirEntries))
	for _, e := range dirEntries {
		info, err := e.Info()
		if err != nil {
			continue
		}

		var entryURL string
		if trimmed == "" {
			entryURL = baseRoute + "/" + e.Name()
		} else {
			entryURL = baseRoute + "/" + trimmed + "/" + e.Name()
		}

		sizeStr := "—"
		if !e.IsDir() {
			sizeStr = strutil.Size(info.Size())
		}

		entries = append(entries, dirEntry{
			Name:        e.Name(),
			IsDir:       e.IsDir(),
			SizeStr:     sizeStr,
			ModStr:      info.ModTime().Format("2006-01-02  15:04"),
			BrowseURL:   entryURL,
			DownloadURL: entryURL + "?download",
			InfoURL:     entryURL + "?info",
		})
	}

	parent := ""

	if trimmed != "" {
		up := path.Dir("/" + trimmed)
		if up == "/" {
			parent = baseRoute + "/"
		} else {
			parent = baseRoute + up
		}
	}

	title := "/"
	if trimmed != "" {
		title = trimmed
	}

	readme := ""
	if b, err := os.ReadFile(filepath.Join(absPath, "README.md")); err == nil {
		readme = string(b)
	}

	data := dirTemplateData{
		Title:       title,
		Breadcrumbs: buildBreadcrumbs(baseRoute, trimmed),
		Parent:      parent,
		Entries:     entries,
		IsEmpty:     len(entries) == 0,
		Readme:      readme,
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
	path, base string,
	stat os.FileInfo,
) {
	rel, _ := filepath.Rel(base, path)
	rel = filepath.ToSlash(rel)

	relPath := "/" + rel

	res := FileInfoResponse{
		Name:        stat.Name(),
		Path:        relPath,
		Size:        stat.Size(),
		Modified:    stat.ModTime().UTC(),
		IsDirectory: stat.IsDir(),
	}

	if !stat.IsDir() {
		ext := filepath.Ext(stat.Name())
		res.Extension = ext
		res.MimeType = mime.TypeByExtension(ext)

		if hash, err := cryptoutil.MD5(path); err != nil {
			logutil.Errorf(logutil.Get(), "handleInfo md5 %q: %v\n", path, err)
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
