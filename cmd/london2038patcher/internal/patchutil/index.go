package patchutil

import (
	"errors"

	"github.com/ricochhet/london2038patcher/pkg/byteutil"
	"github.com/ricochhet/london2038patcher/pkg/errutil"
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

// Parse parses the byte buffer into an Index.
func Parse(buf []byte) (*Index, error) {
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
