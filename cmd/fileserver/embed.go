package main

import (
	"embed"

	"github.com/ricochhet/london2038patcher/pkg/embedutil"
)

//go:embed web/*
var webFS embed.FS

func Embed() *embedutil.EmbeddedFileSystem {
	return &embedutil.EmbeddedFileSystem{
		Initial: "web",
		FS:      webFS,
	}
}
