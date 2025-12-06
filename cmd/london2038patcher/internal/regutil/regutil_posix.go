//go:build !windows
// +build !windows

package regutil

// Regedit edits the registry and adds keys created by Hellgate: London setup.
func Regedit(_, _ string) error {
	return nil
}
