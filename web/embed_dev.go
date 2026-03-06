//go:build dev

package web

import "io/fs"

// Dev mode: no embedded assets, frontend served by Vite dev server
var Assets fs.FS
