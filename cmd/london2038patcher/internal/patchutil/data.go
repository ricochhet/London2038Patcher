package patchutil

import (
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ricochhet/london2038patcher/pkg/errutil"
)

// Unpack unpacks the specified path with the provided index.
func (idx *Index) Unpack(path, output, localization string, appendLocale bool) error {
	if err := checkLocale(localization); err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return errutil.WithFrame(err)
	}
	defer f.Close()

	if err := os.MkdirAll(output, 0o755); err != nil {
		return errutil.WithFrame(err)
	}

	for _, entry := range idx.Files {
		if entry.FileSize <= 0 {
			continue
		}

		if entry.Localization != 0 && entry.Localization != locales[localization] {
			continue
		}

		target := filepath.Join(output, filepath.FromSlash(entry.FileName))

		if appendLocale && entry.Localization != 0 {
			ext := fmt.Sprintf(".%d", entry.Localization)
			target += ext
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return errutil.WithFrame(err)
		}

		fmt.Fprintf(os.Stdout, "Extracting: %s (%d bytes)\n", entry.FileName, entry.FileSize)

		if _, err := f.Seek(entry.DatOffset, io.SeekStart); err != nil {
			return errutil.WithFrame(err)
		}

		buf := make([]byte, entry.FileSize)
		if _, err := io.ReadFull(f, buf); err != nil {
			return errutil.WithFrame(err)
		}

		if err := os.WriteFile(target, buf, 0o644); err != nil {
			return errutil.WithFrame(err)
		}
	}

	return nil
}

// Pack packs the specified path with the provided index.
func (idx *Index) Pack(path, output, localization string, appendLocale bool) error {
	if err := checkLocale(localization); err != nil {
		return err
	}

	f, err := os.Create(output)
	if err != nil {
		return errutil.WithFrame(err)
	}
	defer f.Close()

	for _, entry := range idx.Files {
		if entry.Localization != 0 && entry.Localization != locales[localization] {
			continue
		}

		source := filepath.Join(path, filepath.FromSlash(entry.FileName))

		if appendLocale && entry.Localization != 0 {
			source += fmt.Sprintf(".%d", entry.Localization)
		}

		buf, err := os.ReadFile(source)
		switch {
		case err != nil:
			buf = make([]byte, entry.FileSize)

			fmt.Fprintf(os.Stdout, "Missing file, zeroing: %s\n", source)

		case int64(len(buf)) < entry.FileSize:
			padded := make([]byte, entry.FileSize)
			copy(padded, buf)
			buf = padded

		case int64(len(buf)) > entry.FileSize:
			buf = buf[:entry.FileSize]
		}

		if _, err := f.Seek(entry.DatOffset, io.SeekStart); err != nil {
			return errutil.WithFrame(err)
		}

		if _, err := f.Write(buf); err != nil {
			return errutil.WithFrame(err)
		}

		fmt.Fprintf(os.Stdout, "Packing: %s (%d bytes)\n", source, len(buf))
	}

	return nil
}

// PackWithIndex generates both a .dat and .idx file from the input folder.
func PackWithIndex(path, indexFile, datFile, localization string, appendLocale bool) error {
	if err := checkLocale(localization); err != nil {
		return err
	}

	var idx Index

	idx.Header.PatchType = 1
	idx.Header.EndToken = 1147496776

	var offset int64

	err := filepath.Walk(path, func(target string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		rel, err := filepath.Rel(path, target)
		if err != nil {
			return errutil.WithFrame(err)
		}

		file := filepath.ToSlash(rel)
		loc := int16(0)

		if appendLocale {
			file, loc = removeLocaleExt(file)
		}

		buf, err := os.ReadFile(target)
		if err != nil {
			return errutil.WithFrame(err)
		}

		entry := Entry{
			FileName:     file,
			UsedInX86:    true,
			UsedInX64:    true,
			Localization: loc,
			FileSize:     int64(len(buf)),
			DatOffset:    offset,
			Hash:         crc32.ChecksumIEEE(buf),
		}

		idx.Files = append(idx.Files, entry)
		offset += int64(len(buf))

		return nil
	})
	if err != nil {
		return err
	}

	f, err := os.Create(datFile)
	if err != nil {
		return errutil.WithFrame(err)
	}
	defer f.Close()

	for _, entry := range idx.Files {
		source := filepath.Join(path, entry.FileName)
		if appendLocale && entry.Localization != 0 {
			source += fmt.Sprintf(".%d", entry.Localization)
		}

		buf, err := os.ReadFile(source)
		if err != nil {
			buf = make([]byte, entry.FileSize)
		}

		if _, err := f.Write(buf); err != nil {
			return errutil.WithFrame(err)
		}

		fmt.Fprintf(os.Stdout, "Packing: %s (%d bytes)\n", source, len(buf))
	}

	buf, err := Encode(&idx)
	if err != nil {
		return err
	}

	if err := os.WriteFile(indexFile, buf, 0o644); err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// check if locale is a (known) validate locale.
func checkLocale(s string) error {
	if locales[s] != 0 {
		return nil
	}

	return errutil.WithFramef("Locale: %s does not exist in locales", s)
}

// removeLocaleExt removes the locale extension from a file name if it exists.
func removeLocaleExt(s string) (string, int16) {
	ext := filepath.Ext(s)
	if len(ext) <= 1 {
		return s, 0
	}

	n, err := strconv.Atoi(ext[1:])
	if err != nil {
		return s, 0
	}

	for _, code := range locales {
		if int16(n) == code {
			return s[:len(s)-len(ext)], int16(n)
		}
	}

	return s, 0
}
