package main

import (
	"embed"
	"io/fs"
)

//go:embed all:web/dist
var webAssets embed.FS

// GetWebFS returns the embedded web assets filesystem  
func GetWebFS() (fs.FS, error) {
	return fs.Sub(webAssets, "web/dist")
}
