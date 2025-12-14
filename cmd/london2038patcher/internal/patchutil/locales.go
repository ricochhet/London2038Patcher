package patchutil

import (
	"path/filepath"
	"strconv"

	"github.com/ricochhet/london2038patcher/pkg/errutil"
)

type LocaleMap struct {
	m map[string]int16
}

// NewDefaultLocales returns a new LocaleMap with default values.
func NewDefaultLocales() *LocaleMap {
	return &LocaleMap{m: map[string]int16{
		"en":  17509,
		"unk": 26952,
	}}
}

// toInt converts a slice of strings into an int16 slice based on the LocaleMap.
func (lm *LocaleMap) toInt(m []string) ([]int16, error) {
	if len(m) == 0 {
		return nil, nil
	}

	out := make([]int16, len(m))
	for i, code := range m {
		loc, ok := lm.m[code]
		if !ok || loc == 0 {
			return nil, errutil.WithFramef(
				"Locale: %q does not exist in locales",
				code,
			)
		}

		out[i] = loc
	}

	return out, nil
}

// localeSet converts a slice of locale ints to a map.
func localeSet(m []int16) map[int16]struct{} {
	if len(m) == 0 {
		return nil
	}

	set := make(map[int16]struct{}, len(m))
	for _, l := range m {
		set[l] = struct{}{}
	}

	return set
}

// localeAllowed returns true if the locale is in the map.
func localeAllowed(m map[int16]struct{}, v int16) bool {
	if m == nil || v == 0 {
		return true
	}

	_, ok := m[v]

	return ok
}

// removeLocaleExt removes the locale extension from a file name if it exists.
func (lm *LocaleMap) removeLocaleExt(s string) (string, int16) {
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
