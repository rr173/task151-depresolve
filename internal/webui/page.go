package webui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/index.html web/style.css web/app.js
var assets embed.FS

func Handler() http.Handler { sub, _ := fs.Sub(assets, "web"); return http.FileServer(http.FS(sub)) }
