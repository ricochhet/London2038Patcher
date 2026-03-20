package browse

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ricochhet/london2038patcher/pkg/errutil"
)

// handleSearch walks abs and returns JSON results matching the search query.
//   - ?search=q              — filename match
//   - ?search=q&content=1    — also search inside file contents (text files ≤ 10 MB)
//   - ext:dat or ext:.dat    — restrict results to files with that extension
//   - extension:dat          — alias for ext:
func handleSearch(
	w http.ResponseWriter,
	r *http.Request,
	abs, route string,
	hidden []string,
) error {
	raw := strings.TrimSpace(r.URL.Query().Get("search"))
	contentSearch := r.URL.Query().Has("content")

	query, filter := parseSearchQuery(raw)

	var results []searchResult

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if query == "" && filter == "" {
		return errutil.WithFrame(json.NewEncoder(w).Encode(results))
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
		if filter != "" && ext != filter {
			return nil
		}

		rel, err := filepath.Rel(abs, walkPath)
		if err != nil {
			return errutil.WithFrame(err)
		}

		rel = filepath.ToSlash(rel)
		dir := path.Dir(rel)

		var dirURL string
		if dir == "." {
			dirURL = route + "/"
		} else {
			dirURL = route + "/" + dir
		}

		if query == "" || strings.Contains(strings.ToLower(info.Name()), query) {
			results = append(results, searchResult{
				Name:         info.Name(),
				RelPath:      rel,
				HighlightURL: dirURL + "?highlight=" + url.QueryEscape(info.Name()),
				DownloadURL:  route + "/" + rel + "?download",
				MatchType:    "name",
			})

			return nil
		}

		if contentSearch && query != "" {
			if snippet, ok := searchFileContent(walkPath, info, query); ok {
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

	return enc.Encode(results)
}

// searchFileContent checks whether query appears inside the given file.
func searchFileContent(path string, info os.FileInfo, query string) (string, bool) {
	if info.Size() > maxContentSearchSize {
		return "", false
	}

	data, err := os.ReadFile(path)
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

	start := max(idx-60, 0)
	end := min(idx+len(query)+60, len(data))
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

		var (
			val   string
			isTag bool
		)

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
