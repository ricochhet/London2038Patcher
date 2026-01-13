package strutil

import "strings"

// ToSlice converts a string to a slice based on the separator provided.
func ToSlice(s, sep string) []string {
	return strings.Split(s, sep)
}

// Empty returns s if it's not empty, otherwise returns "<empty>".
func Empty(s string) string {
	if s != "" {
		return s
	}

	return "<empty>"
}
