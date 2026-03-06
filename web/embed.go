//go:build !dev

package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var assets embed.FS

// Assets exposes embedded files as fs.FS interface (nil-able in dev mode).
var Assets fs.FS

func init() {
	Assets = assets
}
