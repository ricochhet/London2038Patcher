package patchutil

import (
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"

	"github.com/ricochhet/london2038patcher/pkg/errutil"
)

// Unpack unpacks the specified path with the provided index.
func (idx *Index) Unpack(path, output string, locales []int16) error {
	allowed := localeSet(locales)

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

		if !localeAllowed(allowed, entry.Localization) {
			continue
		}

		target := filepath.Join(output, filepath.FromSlash(entry.FileName))
		if entry.Localization != 0 {
			target += fmt.Sprintf(".%d", entry.Localization)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return errutil.WithFrame(err)
		}

		fmt.Fprintf(os.Stdout, "Extracting: %s (%d bytes)\n", target, entry.FileSize)

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
func (idx *Index) Pack(path, output string, locales []int16) error {
	allowed := localeSet(locales)

	f, err := os.Create(output)
	if err != nil {
		return errutil.WithFrame(err)
	}
	defer f.Close()

	for _, entry := range idx.Files {
		if !localeAllowed(allowed, entry.Localization) {
			continue
		}

		source := filepath.Join(path, filepath.FromSlash(entry.FileName))
		if entry.Localization != 0 {
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
func (lm *LocaleMap) PackWithIndex(path, indexFile, datFile string, locales []int16) error {
	allowed := localeSet(locales)

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
		file, loc := lm.removeLocaleExt(file)

		if !localeAllowed(allowed, loc) {
			return nil
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
		if entry.Localization != 0 {
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

	data, err := Encode(&idx)
	if err != nil {
		return err
	}

	return os.WriteFile(indexFile, data, 0o644)
}
