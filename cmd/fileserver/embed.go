package main

import (
	"embed"
	"path/filepath"

	"github.com/ricochhet/london2038patcher/pkg/embedutil"
)

//go:embed wwwroot/fileserver/*
var webFS embed.FS

func Embed() *embedutil.EmbeddedFileSystem {
	return &embedutil.EmbeddedFileSystem{
		Initial: filepath.ToSlash(filepath.Join("wwwroot", "fileserver")),
		FS:      webFS,
	}
}
