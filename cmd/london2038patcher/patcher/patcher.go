package patcher

import (
	"context"
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"strings"

	"github.com/ricochhet/london2038patcher/pkg/dlutil"
	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
	"github.com/ricochhet/london2038patcher/pkg/xmlutil"
)

type Patcher struct {
	HTTPClient dlutil.HTTPClient

	ChecksumURL  string
	PatchURL     string
	ChecksumFile string

	HellgateCUKey string
	HellgateKey   string

	UsePatchDir bool
	PatchDir    string
}

type FileEntry struct {
	Name     string `xml:"name,attr"`
	Hash     string `xml:"hash,attr"`
	Filesize string `xml:"filesize,attr"`
	Download string `xml:"download,attr"`
}

type Files struct {
	Entries []FileEntry `xml:"file"`
}

// Download downloads the checksums and files for London 2038.
func (p *Patcher) Download() error {
	files, err := p.downloadChecksums()
	if err != nil {
		return errutil.New("p.downloadChecksums", err)
	}

	if p.UsePatchDir {
		path, err := patchDir(files)
		if err != nil {
			return errutil.New("patchDir", err)
		}

		p.PatchDir = path
	}

	if err := p.downloadFiles(files); err != nil {
		return errutil.New("p.downloadFiles", err)
	}

	return nil
}

// downloadChecksums downloads the checksum file and unmarshals it into a Files struct.
func (p *Patcher) downloadChecksums() (*Files, error) {
	if err := p.HTTPClient.Download(
		context.Background(),
		p.ChecksumFile,
		p.ChecksumURL,
	); err != nil {
		return &Files{}, errutil.New("p.HTTPClient.Download", err)
	}

	files, err := xmlutil.ReadAndUnmarshal[Files](p.ChecksumFile)
	if err != nil {
		return &Files{}, errutil.New("xmlutil.ReadAndUnmarshal", err)
	}

	return files, nil
}

// downloadFiles processes the files by downloading them to the correct directory.
func (p *Patcher) downloadFiles(files *Files) error {
	for _, entry := range files.Entries {
		if strings.ToLower(entry.Download) != "true" {
			continue
		}

		path := entry.Name
		if p.UsePatchDir {
			path = filepath.Join(p.PatchDir, entry.Name)
		}

		url := p.PatchURL + strings.ReplaceAll(entry.Name, "\\", "/")

		if err := fsutil.Ensure(path); err != nil {
			return errutil.New("fsutil.Ensure", err)
		}

		if fsutil.Validate(path, entry.Hash, md5.New()) {
			logutil.Infof(logutil.Get(), "Skipping: %s (already up-to-date)\n", path)
			continue
		}

		logutil.Infof(logutil.Get(), "Downloading: %s to %s\n", url, path)

		if err := p.HTTPClient.Download(context.Background(), path, url); err != nil {
			return errutil.New("p.HTTPClient.Download", err)
		}
	}

	return nil
}

// patchDir creates a top level patch folder name using CRC32 of all file hashes.
func patchDir(files *Files) (string, error) {
	var hash string

	var sb strings.Builder

	for _, f := range files.Entries {
		if strings.ToLower(f.Download) == "true" {
			sb.WriteString(f.Hash)
		}
	}

	hash += sb.String()
	crc := crc32.ChecksumIEEE([]byte(hash))
	path := fmt.Sprintf("London2038Patcher/%08X", crc)

	return path, os.MkdirAll(path, 0o755)
}
