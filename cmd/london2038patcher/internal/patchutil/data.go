package patchutil

import (
	"bufio"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/ricochhet/london2038patcher/pkg/errutil"
)

// Unpack unpacks the specified path with the provided index.
func (idx *Index) Unpack(path, output string, locales []int16, archs []string) error {
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
		if entry.FileSize <= 0 || !localeAllowed(allowed, entry.Localization) ||
			skipArch(entry, archs) {
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

		outFile, err := os.Create(target)
		if err != nil {
			return errutil.WithFrame(err)
		}

		bw := bufio.NewWriterSize(outFile, 4*1024*1024)
		if _, err := bw.Write(buf); err != nil {
			outFile.Close()
			return errutil.WithFrame(err)
		}

		if err := bw.Flush(); err != nil {
			outFile.Close()
			return errutil.WithFrame(err)
		}

		outFile.Close()
	}

	return nil
}

// Pack packs the specified path with the provided index.
func (idx *Index) Pack(path, output string, locales []int16, archs []string) error {
	allowed := localeSet(locales)

	f, err := os.Create(output)
	if err != nil {
		return errutil.WithFrame(err)
	}
	defer f.Close()

	bw := bufio.NewWriterSize(f, 4*1024*1024)
	defer bw.Flush()

	for _, entry := range idx.Files {
		if !localeAllowed(allowed, entry.Localization) || skipArch(entry, archs) {
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

		if _, err := bw.Write(buf); err != nil {
			return errutil.WithFrame(err)
		}

		fmt.Fprintf(os.Stdout, "Packing: %s (%d bytes)\n", source, len(buf))
	}

	return nil
}

// PackWithIndex generates both a .dat and .idx file from the input folder.
func (lm *LocaleMap) PackWithIndex(
	path, indexFile, datFile string,
	locales []int16,
	archs []string,
) error {
	allowed := localeSet(locales)

	var idx Index

	idx.Header.PatchType = 1
	idx.Header.EndToken = 1147496776

	var offset int64

	if err := lm.readIntoIndex(&idx, path, offset, allowed); err != nil {
		return err
	}

	f, err := os.Create(datFile)
	if err != nil {
		return errutil.WithFrame(err)
	}
	defer f.Close()

	bw := bufio.NewWriterSize(f, 4*1024*1024)
	defer bw.Flush()

	for _, entry := range idx.Files {
		if skipArch(entry, archs) {
			continue
		}

		source := filepath.Join(path, entry.FileName)
		if entry.Localization != 0 {
			source += fmt.Sprintf(".%d", entry.Localization)
		}

		buf, err := os.ReadFile(source)
		if err != nil {
			buf = make([]byte, entry.FileSize)
		}

		if _, err := bw.Write(buf); err != nil {
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

// readIntoIndex reads all files in the path and adds them to the index.
func (lm *LocaleMap) readIntoIndex(
	idx *Index,
	path string,
	offset int64,
	allowed map[int16]struct{},
) error {
	return filepath.Walk(path, func(target string, info os.FileInfo, err error) error {
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
}

// skipArch returns true if the architecture should be skipped.
func skipArch(entry Entry, archs []string) bool {
	if entry.UsedInX64 && entry.UsedInX86 {
		return false
	}

	if entry.UsedInX64 && !allowX64(archs) {
		return true
	}

	return entry.UsedInX86 && !allowX86(archs)
}

// allowX86 returns true if x86 arch is allowed.
func allowX86(s []string) bool {
	return slices.Contains(s, "x86")
}

// allowX64 returns true if x64 arch is allowed.
func allowX64(s []string) bool {
	return slices.Contains(s, "x64")
}
