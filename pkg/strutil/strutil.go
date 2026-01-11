package strutil

// Empty returns s if it's not empty, otherwise returns "<nil>".
func Empty(s string) string {
	if s != "" {
		return s
	}

	return "<empty>"
}
