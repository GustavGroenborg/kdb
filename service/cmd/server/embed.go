package main

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var embeddedAssets embed.FS

// getEmbeddedFS returns an fs.FS sub-rooted at "dist", or nil if not found.
func getEmbeddedFS() fs.FS {
	sub, err := fs.Sub(embeddedAssets, "dist")
	if err != nil {
		return nil
	}
	return sub
}
