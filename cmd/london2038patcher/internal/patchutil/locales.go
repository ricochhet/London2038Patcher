package patchutil

import (
	"path/filepath"
	"strconv"

	"github.com/ricochhet/london2038patcher/pkg/errutil"
)

type LocaleRegistry struct {
	m map[string]int16
}

type LocaleFilter struct {
	allowed map[int16]struct{}
}

// NewLocaleFilter returns a new locale filter based on the local registry.
func NewLocaleFilter(lr *LocaleRegistry, codes []string) (*LocaleFilter, error) {
	if len(codes) == 0 {
		return &LocaleFilter{allowed: nil}, nil
	}

	set := make(map[int16]struct{}, len(codes))
	for _, code := range codes {
		loc, ok := lr.m[code]
		if !ok || loc == 0 {
			return nil, errutil.WithFramef(
				"Locale: %q does not exist in locales",
				code,
			)
		}

		set[loc] = struct{}{}
	}

	return &LocaleFilter{allowed: set}, nil
}

// Allowed returns true if the code is allowed by the location filter.
func (lf *LocaleFilter) Allowed(v int16) bool {
	if lf == nil || lf.allowed == nil || v == 0 {
		return true
	}

	_, ok := lf.allowed[v]

	return ok
}

// NewDefaultLocaleRegistry returns a new LocaleRegistry with default values.
func NewDefaultLocaleRegistry() *LocaleRegistry {
	return &LocaleRegistry{m: map[string]int16{
		"en":  17509,
		"unk": 26952,
	}}
}

// removeLocaleExt removes the locale extension from a file name if it exists.
func (lm *LocaleRegistry) removeLocaleExt(s string) (string, int16) {
	ext := filepath.Ext(s)
	if len(ext) <= 1 {
		return s, 0
	}

	n, err := strconv.Atoi(ext[1:])
	if err != nil {
		return s, 0
	}

	for _, code := range lm.m {
		if int16(n) == code {
			return s[:len(s)-len(ext)], int16(n)
		}
	}

	return s, 0
}
