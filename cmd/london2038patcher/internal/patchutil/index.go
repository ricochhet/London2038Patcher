package patchutil

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"unicode/utf16"

	"github.com/ricochhet/london2038patcher/pkg/byteutil"
	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
)

type Header struct {
	PatchType            uint32 `json:"patchType"`
	PatchMajorVersion    uint32 `json:"patchMajor"`
	PatchMinorVersion    uint32 `json:"patchMinor"`
	PatchBuildVersion    uint32 `json:"patchBuild"`
	PatchPrivateVersion  uint32 `json:"patchPrivate"`
	RequiredMajorVersion uint32 `json:"requiredMajor"`
	RequiredMinorVersion uint32 `json:"requiredMinor"`
	RequiredBuildVersion uint32 `json:"requiredBuild"`
	RequiredPrivate      uint32 `json:"requiredPrivate"`
	Unknown1             uint32 `json:"unknown1"`
	Unknown2             uint32 `json:"unknown2"`
	Unknown3             uint32 `json:"unknown3"`
	Unknown4             uint32 `json:"unknown4"`
	Unknown5             uint32 `json:"unknown5"`
	EndToken             uint32 `json:"endToken"`
}

type Pattern struct {
	Pattern string `json:"pattern"`
}

type Entry struct {
	FileName     string `json:"fileName"`
	UsedInX86    bool   `json:"usedInX86"`
	UsedInX64    bool   `json:"usedInX64"`
	Localization int16  `json:"localization"`
	FileSize     int64  `json:"fileSize"`
	DatOffset    int64  `json:"datOffset"`
	Hash         uint32 `json:"hash"`
}

type Index struct {
	Header       Header    `json:"header"`
	Search       []Pattern `json:"searchPatterns"`
	Files        []Entry   `json:"files"`
	FullConsumed bool      `json:"fullConsumed"`
	OriginalSize int       `json:"originalSize"`
}

// DecodeFile decodes an index file to the specified output.
func DecodeFile(path, output string) ([]byte, error) {
	if !fsutil.Exists(path) {
		return nil, errutil.WithFramef("path does not exist: %s", path)
	}

	f, err := fsutil.Read(path)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	idx, err := Decode(f)
	if err != nil {
		return nil, err
	}

	outFile, err := os.Create(output)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}
	defer outFile.Close()

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	bw := bufio.NewWriterSize(outFile, 4*1024*1024)
	if _, err := bw.Write(data); err != nil {
		return nil, errutil.WithFrame(err)
	}

	if err := bw.Flush(); err != nil {
		return nil, errutil.WithFrame(err)
	}

	return data, nil
}

// EncodeFile encodes an index file to the specified output.
func EncodeFile(path, output string) error {
	if !fsutil.Exists(path) {
		return errutil.WithFramef("path does not exist: %s", path)
	}

	f, err := fsutil.Read(path)
	if err != nil {
		return errutil.WithFrame(err)
	}

	var idx Index
	if err := json.Unmarshal(f, &idx); err != nil {
		return errutil.WithFrame(err)
	}

	buf, err := Encode(&idx)
	if err != nil {
		return err
	}

	outFile, err := os.Create(output)
	if err != nil {
		return errutil.WithFrame(err)
	}
	defer outFile.Close()

	bw := bufio.NewWriterSize(outFile, 4*1024*1024)
	if _, err := bw.Write(buf); err != nil {
		return errutil.WithFrame(err)
	}

	return bw.Flush()
}

// Decode decodes the byte buffer into an Index.
func Decode(buf []byte) (*Index, error) {
	offset := 0

	version := byteutil.ReadU32(buf, &offset)
	if version > 4 {
		return &Index{}, errutil.WithFramef("not a patch index file (version=%d)", version)
	}

	var idx Index

	idx.OriginalSize = len(buf)

	idx.Header.PatchType = version
	idx.Header.PatchMajorVersion = byteutil.ReadU32(buf, &offset)
	idx.Header.PatchMinorVersion = byteutil.ReadU32(buf, &offset)
	idx.Header.PatchBuildVersion = byteutil.ReadU32(buf, &offset)
	idx.Header.PatchPrivateVersion = byteutil.ReadU32(buf, &offset)
	idx.Header.RequiredMajorVersion = byteutil.ReadU32(buf, &offset)
	idx.Header.RequiredMinorVersion = byteutil.ReadU32(buf, &offset)
	idx.Header.RequiredBuildVersion = byteutil.ReadU32(buf, &offset)
	idx.Header.RequiredPrivate = byteutil.ReadU32(buf, &offset)
	idx.Header.Unknown1 = byteutil.ReadU32(buf, &offset)
	idx.Header.Unknown2 = byteutil.ReadU32(buf, &offset)
	idx.Header.Unknown3 = byteutil.ReadU32(buf, &offset)
	idx.Header.Unknown4 = byteutil.ReadU32(buf, &offset)
	idx.Header.Unknown5 = byteutil.ReadU32(buf, &offset)
	idx.Header.EndToken = byteutil.ReadU32(buf, &offset)

	if idx.Header.EndToken != 1147496776 {
		return &Index{}, errutil.WithFrame(errors.New("invalid end token in header"))
	}

	for {
		check := byteutil.ReadI32(buf, &offset)
		if check == 0 {
			break
		}

		charCount := byteutil.ReadI32(buf, &offset)
		pattern := byteutil.ReadStringUnicode(buf, &offset, int(charCount))
		idx.Search = append(idx.Search, Pattern{Pattern: pattern})
	}

	for {
		charCount := byteutil.ReadI32(buf, &offset)
		filename := byteutil.ReadStringUnicode(buf, &offset, int(charCount))
		usedInX86 := byteutil.ReadI32(buf, &offset) != 0
		usedInX64 := byteutil.ReadI32(buf, &offset) != 0
		localization := int16(byteutil.ReadU16(buf, &offset))
		fileSize := byteutil.ReadI64(buf, &offset)
		datOffset := byteutil.ReadI64(buf, &offset)
		hash := byteutil.ReadU32(buf, &offset)
		more := byteutil.ReadI32(buf, &offset)

		idx.Files = append(idx.Files, Entry{
			FileName:     filename,
			UsedInX86:    usedInX86,
			UsedInX64:    usedInX64,
			Localization: localization,
			FileSize:     fileSize,
			DatOffset:    datOffset,
			Hash:         hash,
		})

		if more == -1 {
			break
		}
	}

	idx.FullConsumed = (offset == len(buf))

	return &idx, nil
}

// Encode encodes the Index into an HGL patch index.
func Encode(idx *Index) ([]byte, error) {
	size := indexSize(idx)
	buf := make([]byte, size)
	offset := 0

	byteutil.WriteU32(buf, &offset, idx.Header.PatchType)
	byteutil.WriteU32(buf, &offset, idx.Header.PatchMajorVersion)
	byteutil.WriteU32(buf, &offset, idx.Header.PatchMinorVersion)
	byteutil.WriteU32(buf, &offset, idx.Header.PatchBuildVersion)
	byteutil.WriteU32(buf, &offset, idx.Header.PatchPrivateVersion)
	byteutil.WriteU32(buf, &offset, idx.Header.RequiredMajorVersion)
	byteutil.WriteU32(buf, &offset, idx.Header.RequiredMinorVersion)
	byteutil.WriteU32(buf, &offset, idx.Header.RequiredBuildVersion)
	byteutil.WriteU32(buf, &offset, idx.Header.RequiredPrivate)
	byteutil.WriteU32(buf, &offset, idx.Header.Unknown1)
	byteutil.WriteU32(buf, &offset, idx.Header.Unknown2)
	byteutil.WriteU32(buf, &offset, idx.Header.Unknown3)
	byteutil.WriteU32(buf, &offset, idx.Header.Unknown4)
	byteutil.WriteU32(buf, &offset, idx.Header.Unknown5)
	byteutil.WriteU32(buf, &offset, idx.Header.EndToken)

	for _, p := range idx.Search {
		byteutil.WriteI32(buf, &offset, 1)
		byteutil.WriteI32(buf, &offset, int32(utf16Len(p.Pattern)))
		byteutil.WriteStringUnicode(buf, &offset, p.Pattern)
	}

	byteutil.WriteI32(buf, &offset, 0)

	for i, f := range idx.Files {
		byteutil.WriteI32(buf, &offset, int32(utf16Len(f.FileName)))
		byteutil.WriteStringUnicode(buf, &offset, f.FileName)

		if f.UsedInX86 {
			byteutil.WriteI32(buf, &offset, 1)
		} else {
			byteutil.WriteI32(buf, &offset, 0)
		}

		if f.UsedInX64 {
			byteutil.WriteI32(buf, &offset, 1)
		} else {
			byteutil.WriteI32(buf, &offset, 0)
		}

		byteutil.WriteU16(buf, &offset, uint16(f.Localization))
		byteutil.WriteI64(buf, &offset, f.FileSize)
		byteutil.WriteI64(buf, &offset, f.DatOffset)
		byteutil.WriteU32(buf, &offset, f.Hash)

		if i == len(idx.Files)-1 {
			byteutil.WriteI32(buf, &offset, -1)
		} else {
			byteutil.WriteI32(buf, &offset, 0)
		}
	}

	if offset != size {
		return nil, errutil.WithFramef("encode size mismatch: wrote %d, expected %d", offset, size)
	}

	return buf, nil
}

// indexSize computes the byte length of the encoded Index.
func indexSize(idx *Index) int {
	size := 0

	size += 15 * 4

	for _, p := range idx.Search {
		size += 4
		size += 4
		size += utf16Len(p.Pattern) * 2
	}

	size += 4

	for _, f := range idx.Files {
		size += 4
		size += utf16Len(f.FileName) * 2
		size += 4
		size += 4
		size += 2
		size += 8
		size += 8
		size += 4
		size += 4
	}

	return size
}

// utf16Len gets the utf16 encoded length of the string.
func utf16Len(s string) int {
	return len(utf16.Encode([]rune(s)))
}
