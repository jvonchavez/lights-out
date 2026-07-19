// Package web embeds the built frontend so the binary ships with its own
// assets and the deployable unit stays a single file.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// all: is required. Without it embed skips files whose names begin with an
// underscore or dot, and Vite emits hashed asset names that can.
//
//go:embed all:dist
var dist embed.FS

// FS returns the built frontend rooted at dist/.
func FS() (fs.FS, error) { return fs.Sub(dist, "dist") }

// Built reports whether a real frontend build is embedded, as opposed to
// the placeholder that keeps this package compiling before the first
// `make web`.
func Built() bool {
	_, err := dist.ReadFile("dist/index.html")
	return err == nil
}

// Handler serves the embedded frontend with an SPA fallback: any path that
// is not a real file returns index.html so client-side routing works.
func Handler() (http.Handler, error) {
	sub, err := FS()
	if err != nil {
		return nil, err
	}
	files := http.FileServerFS(sub)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := fs.Stat(sub, p); err != nil {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}
		// Go's mime table does not always know .wasm, and a wrong type
		// makes instantiateStreaming refuse the module.
		if strings.HasSuffix(p, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		}
		if strings.HasPrefix(p, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		files.ServeHTTP(w, r)
	}), nil
}
