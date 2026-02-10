package ui

import (
	"embed"
	"io/fs"
)

//go:embed assets/*
var assetsEmbed embed.FS

// Assets returns the embedded filesystem for the UI assets
func Assets() (fs.FS, error) {
	return fs.Sub(assetsEmbed, "assets")
}
