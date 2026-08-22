package webui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/index.html web/style.css web/app.js
var assets embed.FS

// Handler serves the embedded static assets (index.html, style.css, app.js)
// so engineers can open the application page at /static/ and watch resolution
// status. The web/ prefix is stripped so files live at the FS root.
func Handler() http.Handler {
	sub, err := fs.Sub(assets, "web")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}
