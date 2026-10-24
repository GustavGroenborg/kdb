package static

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Handler serves static files from an fs.FS (e.g. embed.FS) with fallback to index.html for SPA routing.
func Handler(staticFS fs.FS, fallbackDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Try embedded FS first if available
		if staticFS != nil {
			f, err := staticFS.Open(path)
			if err == nil {
				defer f.Close()
				fi, err := f.Stat()
				if err == nil && !fi.IsDir() {
					http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
					return
				}
			}

			// SPA fallback to index.html for non-API routes
			if !strings.HasPrefix(r.URL.Path, "/api") {
				indexFile, err := staticFS.Open("index.html")
				if err == nil {
					defer indexFile.Close()
					var modTime time.Time
					if indexStat, err := indexFile.Stat(); err == nil {
						modTime = indexStat.ModTime()
					}
					if rs, ok := indexFile.(io.ReadSeeker); ok {
						http.ServeContent(w, r, "index.html", modTime, rs)
						return
					}
				}
			}
		}

		// If embedded FS is not available, check filesystem directory
		if fallbackDir != "" {
			filePath := filepath.Join(fallbackDir, path)
			if fi, err := os.Stat(filePath); err == nil && !fi.IsDir() {
				http.ServeFile(w, r, filePath)
				return
			}

			// SPA fallback
			indexPath := filepath.Join(fallbackDir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
				return
			}
		}

		// If neither exists, show default greeting message
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!DOCTYPE html><html><head><title>KDB Befordringsfradrag</title></head>
		<body style="font-family:sans-serif;padding:40px;text-align:center;">
		<h1>KDB Befordringsfradrag Backend Active</h1>
		<p>Backend API is running at <code>/api</code>. Frontend build will be served here once compiled.</p>
		</body></html>`))
	}
}
