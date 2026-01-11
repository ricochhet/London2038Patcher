package pathutil

import "strings"

// Normalize replaces all back slashes with forward slashes.
func Normalize(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
