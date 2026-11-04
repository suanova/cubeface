package webui

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func testDist() fstest.MapFS {
	return fstest.MapFS{
		"web/dist/index.html": &fstest.MapFile{
			Data: []byte("<!doctype html><html><head><title>CubeFace</title></head><body><div id=\"root\"></div></body></html>"),
		},
		"web/dist/static/js/index.abc123.js":   &fstest.MapFile{Data: []byte("console.log('ok')")},
		"web/dist/static/css/index.abc123.css": &fstest.MapFile{Data: []byte("body{}")},
		"web/dist/logo.svg":                    &fstest.MapFile{Data: []byte("<svg></svg>")},
	}
}

func do(t *testing.T, h http.Handler, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

func TestServesSPAAndAssets(t *testing.T) {
	h := New(testDist(), "")

	rec := do(t, h, http.MethodGet, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("index content type = %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("index cache control = %q", cc)
	}
	if got := rec.Body.String(); !strings.Contains(got, "CubeFace") || strings.Contains(got, "<base") {
		t.Fatalf("unexpected index body: %s", got)
	}

	// Deep links fall back to the SPA shell.
	rec = do(t, h, http.MethodGet, "/dashboard/settings")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "CubeFace") {
		t.Fatalf("SPA fallback = %d %s", rec.Code, rec.Body.String())
	}

	// Fingerprinted assets are cached immutably.
	rec = do(t, h, http.MethodGet, "/static/js/index.abc123.js")
	if rec.Code != http.StatusOK {
		t.Fatalf("asset = %d", rec.Code)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Fatalf("asset cache control = %q", cc)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/javascript; charset=utf-8" {
		t.Fatalf("asset content type = %q", ct)
	}

	// PUBLIC assets are served but revalidated.
	rec = do(t, h, http.MethodGet, "/logo.svg")
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatalf("logo = %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Fatalf("logo cache control = %q", cc)
	}

	// Missing assets are 404 rather than the SPA shell.
	rec = do(t, h, http.MethodGet, "/static/js/missing.js")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing asset = %d", rec.Code)
	}

	if rec = do(t, h, http.MethodPost, "/"); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST / = %d", rec.Code)
	}
}

func TestInjectsBasePath(t *testing.T) {
	h := New(testDist(), "/cubeface")

	if rec := do(t, h, http.MethodGet, "/"); rec.Code != http.StatusNotFound {
		t.Fatalf("root outside base path = %d", rec.Code)
	}

	rec := do(t, h, http.MethodGet, "/cubeface")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /cubeface = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `<base href="/cubeface/">`) {
		t.Fatalf("missing base tag: %s", body)
	}
	if !strings.Contains(body, `window.__BASE_PATH__="/cubeface";`) {
		t.Fatalf("missing base path global: %s", body)
	}

	rec = do(t, h, http.MethodGet, "/cubeface/dashboard")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `<base href="/cubeface/">`) {
		t.Fatalf("prefixed SPA fallback = %d %s", rec.Code, rec.Body.String())
	}

	rec = do(t, h, http.MethodGet, "/cubeface/static/js/index.abc123.js")
	if rec.Code != http.StatusOK || rec.Body.String() != "console.log('ok')" {
		t.Fatalf("prefixed asset = %d %q", rec.Code, rec.Body.String())
	}

	if rec = do(t, h, http.MethodGet, "/cubefacex/static/js/index.abc123.js"); rec.Code != http.StatusNotFound {
		t.Fatalf("prefix lookalike = %d", rec.Code)
	}
}

func TestHeadAndMissingBuild(t *testing.T) {
	h := New(testDist(), "")
	rec := do(t, h, http.MethodHead, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("HEAD / = %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("HEAD body = %q", rec.Body.String())
	}

	// A filesystem without the embedded build reports how to build it.
	h = New(fstest.MapFS{"other/file.txt": &fstest.MapFile{Data: []byte("x")}}, "")
	rec = do(t, h, http.MethodGet, "/")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("missing build = %d", rec.Code)
	}
	if body, _ := io.ReadAll(rec.Body); !strings.Contains(string(body), "bun run build") {
		t.Fatalf("missing build hint: %s", body)
	}
}

func TestInjectBasePathEdgeCases(t *testing.T) {
	// No literal <head>: inject before </head>.
	got := string(injectBasePath([]byte("<html></head><body></body></html>"), "/x"))
	if !strings.Contains(got, `<base href="/x/">`) || strings.Contains(got, "<head><base") {
		t.Fatalf("injection without head: %s", got)
	}

	// A document without head is left alone.
	if got := injectBasePath([]byte("<html></html>"), "/x"); string(got) != "<html></html>" {
		t.Fatalf("unexpected injection: %s", got)
	}

	// Empty base path is a no-op.
	if got := injectBasePath([]byte("<head></head>"), ""); string(got) != "<head></head>" {
		t.Fatalf("unexpected empty base injection: %s", got)
	}

	// Attribute and script values are escaped.
	got = string(injectBasePath([]byte("<head></head>"), `/a"><script>alert(1)</script>`))
	if strings.Contains(got, `href="/a"><script>`) {
		t.Fatalf("base href was not escaped: %s", got)
	}
	if strings.Count(got, "</script>") != 1 || !strings.Contains(got, `</script>`) {
		t.Fatalf("script value was not escaped: %s", got)
	}
}
