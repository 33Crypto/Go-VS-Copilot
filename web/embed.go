package web

import "embed"

//go:embed index.html app.css app.js manifest.webmanifest sw.js icons
var Files embed.FS
