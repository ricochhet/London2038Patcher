package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ricochhet/london2038patcher/pkg/dlutil"
	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/xmlutil"
)

type Patcher struct {
	XMLUrl  string
	BaseURL string
	XMLFile string
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
	xmlURLFlag  string
	baseURLFlag string
	xmlFileFlag string
)

var (
	buildDate string
	gitHash   string
	buildOn   string
)

func version() string {
	return fmt.Sprintf(
		"London2038Patcher\n  Build Date: %s\n  Git Hash:  %s\n  Built On:  %s\n",
		buildDate, gitHash, buildOn,
	)
}

//nolint:gochecknoinits // wontfix
func init() {
	flag.StringVar(
		&xmlURLFlag,
		"xmlurl",
		"https://auth.london2038.com/patcher/checksums.xml",
		"URL to the checksums XML file",
	)
	flag.StringVar(
		&baseURLFlag,
		"baseurl",
		"https://auth.london2038.com/patcher/",
		"Base URL for patch files",
	)
	flag.StringVar(&xmlFileFlag, "xmlfile", "checksums.xml", "Local path to save XML file")
}

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Fprint(os.Stdout, version())
		return
	}

	p := Patcher{
		XMLUrl:  xmlURLFlag,
		BaseURL: baseURLFlag,
		XMLFile: xmlFileFlag,
	}

	if err := dlutil.Download(context.Background(), p.XMLFile, p.XMLUrl); err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading checksums: %v\n", err)
		return
	}

	files, err := xmlutil.Unmarshal[Files](p.XMLFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading checksums: %v\n", err)
		return
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

		name := entry.Name
		path := strings.ReplaceAll(entry.Name, "\\", "/")
		url := p.BaseURL + path

		if err := ensure(name); err != nil {
			return errutil.WithFrame(err)
		}

		if validate(name, entry.Hash) {
			fmt.Fprintf(os.Stdout, "Skipping: %s (already up-to-date)\n", name)
			continue
		}

		fmt.Fprintf(os.Stdout, "Downloading: %s to %s\n", url, name)

		if err := dlutil.Download(context.Background(), name, url); err != nil {
			return errutil.WithFrame(err)
		}
	}

	return nil
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
