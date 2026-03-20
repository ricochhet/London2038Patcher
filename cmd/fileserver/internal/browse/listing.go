package browse

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ricochhet/london2038patcher/pkg/logutil"
	"github.com/ricochhet/london2038patcher/pkg/strutil"
)

// handleListing handles creating and listing files and directories in the browser.
func handleListing(
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
