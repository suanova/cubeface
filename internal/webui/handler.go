package webui

import (
	"bytes"
	"encoding/json"
	"html"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

const buildHint = "The frontend has not been built.\nRun: cd web && bun install && bun run build\n"

// New returns a handler that serves the embedded SPA under basePath. Missing
// non-asset paths fall back to index.html for client-side routing.
func New(dist fs.FS, basePath string) http.Handler {
	var sub fs.FS
	if dist != nil {
		if s, err := fs.Sub(dist, "web/dist"); err == nil {
			sub = s
		}
	}
	return &handler{dist: sub, basePath: basePath}
}

type handler struct {
	dist     fs.FS
	basePath string
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	requestPath := r.URL.Path
	if h.basePath != "" {
		if requestPath == h.basePath {
			requestPath = "/"
		} else if strings.HasPrefix(requestPath, h.basePath+"/") {
			requestPath = strings.TrimPrefix(requestPath, h.basePath)
		} else {
			http.NotFound(w, r)
			return
		}
	}
	if requestPath == "" {
		requestPath = "/"
	}

	if h.dist == nil {
		http.Error(w, buildHint, http.StatusServiceUnavailable)
		return
	}

	cleaned := path.Clean("/" + strings.TrimPrefix(requestPath, "/"))
	if cleaned != "/" {
		if data, ok := h.readFile(cleaned); ok {
			h.writeFile(w, r, cleaned, data)
			return
		}
		if path.Ext(cleaned) != "" {
			// Looks like a missing static asset; do not return the SPA shell.
			http.NotFound(w, r)
			return
		}
	}
	h.serveIndex(w, r)
}

func (h *handler) readFile(name string) ([]byte, bool) {
	file, err := h.dist.Open(strings.TrimPrefix(name, "/"))
	if err != nil {
		return nil, false
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(file, 64<<20))
	if err != nil {
		return nil, false
	}
	return data, true
}

func (h *handler) writeFile(w http.ResponseWriter, r *http.Request, name string, data []byte) {
	header := w.Header()
	header.Set("Content-Type", contentType(name))
	if strings.HasPrefix(name, "/static/") {
		header.Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		header.Set("Cache-Control", "no-cache")
	}
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
}

func (h *handler) serveIndex(w http.ResponseWriter, r *http.Request) {
	data, ok := h.readFile("/index.html")
	if !ok {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(buildHint))
		return
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	header.Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(injectBasePath(data, h.basePath))
	}
}

// injectBasePath writes a <base> tag and, behind a path prefix,
// window.__BASE_PATH__ into the SPA shell so the same frontend build works at
// the site root or behind a prefix. The <base> tag is always emitted: the
// built HTML references static assets relatively, so without it a deep link
// such as /datasets/demo-dataset would resolve them under /datasets/.
// json.Marshal escapes quotes, angle brackets and ampersands, so the value
// cannot terminate the inline script; the href is HTML-escaped too.
func injectBasePath(index []byte, basePath string) []byte {
	tag := "<base href=\"" + html.EscapeString(basePath+"/") + "\">"
	if basePath != "" {
		encoded, err := json.Marshal(basePath)
		if err != nil {
			return index
		}
		tag += "<script>window.__BASE_PATH__=" + string(encoded) + ";</script>"
	}
	snippet := []byte("<head>" + tag)
	if bytes.Contains(index, []byte("<head>")) {
		return bytes.Replace(index, []byte("<head>"), snippet, 1)
	}
	if i := bytes.Index(index, []byte("</head>")); i >= 0 {
		return append(append(append([]byte(nil), index[:i]...), []byte(tag)...), index[i:]...)
	}
	return index
}

func contentType(name string) string {
	switch strings.ToLower(path.Ext(name)) {
	case ".html":
		return "text/html; charset=utf-8"
	case ".js", ".mjs":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".ico":
		return "image/x-icon"
	case ".woff2":
		return "font/woff2"
	case ".woff":
		return "font/woff"
	case ".txt":
		return "text/plain; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}
