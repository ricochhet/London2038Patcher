package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ricochhet/london2038patcher/pkg/dlutil"
	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/fsutil"
	"github.com/ricochhet/london2038patcher/pkg/patchutil"
	"github.com/ricochhet/london2038patcher/pkg/xmlutil"
)

type Patcher struct {
	ChecksumURL  string
	PatchURL     string
	ChecksumFile string
	PatchDir     string
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

var (
	checksumURLFlag  string
	patchURLFlag     string
	checksumFileFlag string
	patchDirFlag     bool
)

var (
	buildDate string
	gitHash   string
	buildOn   string
)

func version() string {
	return fmt.Sprintf(
		"London2038Patcher\n\tBuild Date: %s\n\tGit Hash: %s\n\tBuilt On: %s\n",
		buildDate, gitHash, buildOn,
	)
}

//nolint:gochecknoinits // wontfix
func init() {
	flag.StringVar(
		&checksumURLFlag,
		"checksum-url",
		"https://auth.london2038.com/patcher/checksums.xml",
		"URL for checksum file",
	)
	flag.StringVar(
		&patchURLFlag,
		"patch-url",
		"https://auth.london2038.com/patcher/",
		"URL for patch files",
	)
	flag.StringVar(
		&checksumFileFlag,
		"checksum-file",
		"checksums.xml",
		"Path to save checksum file to",
	)
	flag.BoolVar(&patchDirFlag, "patch-dir", false, "Use patch directory for files")
}

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Fprint(os.Stdout, version())
		return
	}

	cmd, err := commands()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
		return
	}

	if cmd {
		return
	}

	p := Patcher{
		ChecksumURL:  checksumURLFlag,
		PatchURL:     patchURLFlag,
		ChecksumFile: checksumFileFlag,
	}

	if err := dlutil.Download(context.Background(), p.ChecksumFile, p.ChecksumURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading checksums: %v\n", err)
		return
	}

	files, err := xmlutil.Unmarshal[Files](p.ChecksumFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading checksums: %v\n", err)
		return
	}

	p.PatchDir, err = patchDir(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating patch folder: %v\n", err)
	}

	if err := p.process(files); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing files: %v\n", err)
	}
}

// process processes the files by downloading them to the correct directory.
func (p *Patcher) process(files *Files) error {
	for _, entry := range files.Entries {
		if strings.ToLower(entry.Download) != "true" {
			continue
		}

		path := entry.Name
		if patchDirFlag {
			path = filepath.Join(p.PatchDir, entry.Name)
		}

		url := p.PatchURL + strings.ReplaceAll(entry.Name, "\\", "/")

		if err := ensure(path); err != nil {
			return errutil.WithFrame(err)
		}

		if validate(path, entry.Hash) {
			fmt.Fprintf(os.Stdout, "Skipping: %s (already up-to-date)\n", path)
			continue
		}

		fmt.Fprintf(os.Stdout, "Downloading: %s to %s\n", url, path)

		if err := dlutil.Download(context.Background(), path, url); err != nil {
			return errutil.WithFrame(err)
		}
	}

	return nil
}

func commands() (bool, error) {
	cmd := flag.Args()[0]
	args := flag.Args()[1:]

	switch cmd {
	case "unpack":
		f, err := fsutil.Read(args[0])
		if err != nil {
			return true, errutil.WithFrame(err)
		}

		idx, err := patchutil.Parse(f)
		if err != nil {
			return true, errutil.WithFrame(err)
		}

		if err := idx.Unpack(args[1], args[2]); err != nil {
			return true, errutil.WithFrame(err)
		}

		return true, nil
	case "help":
		flag.Usage()
		return true, nil
	}

	return false, nil
}

// ensure ensures the file path, returning an error if it fails.
func ensure(path string) error {
	dir := filepath.Dir(path)
	if dir != "." {
		return os.MkdirAll(dir, 0o755)
	}

	return nil
}

// validate checks if a file exists and matches the given MD5 hash.
func validate(path, hash string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return false
	}

	sum := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	return sum == strings.ToUpper(hash)
}

// patchDir creates a top level patch folder name using CRC32 of all file hashes.
func patchDir(files *Files) (string, error) {
	var hash string

	for _, f := range files.Entries {
		if strings.ToLower(f.Download) == "true" {
			hash += f.Hash
		}
	}

	crc := crc32.ChecksumIEEE([]byte(hash))

	path := fmt.Sprintf("London2038Patch%08X", crc)

	return path, os.MkdirAll(path, 0o755)
}
