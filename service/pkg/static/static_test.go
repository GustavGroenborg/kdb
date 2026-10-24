package static

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestStaticServerHeadersAndCacheControl(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html":              {Data: []byte("<!DOCTYPE html><html><body>KDB UI</body></html>")},
		"assets/index-12345.js":   {Data: []byte("console.log('kdb')")},
		"assets/style-67890.css":  {Data: []byte("body { margin: 0; }")},
		"favicon.ico":             {Data: []byte("icon")},
	}

	handler := Handler(mockFS, "")

	tests := []struct {
		name                 string
		path                 string
		expectedStatus       int
		expectedCacheControl string
	}{
		{
			name:                 "Index HTML root",
			path:                 "/",
			expectedStatus:       http.StatusOK,
			expectedCacheControl: "no-cache, must-revalidate",
		},
		{
			name:                 "Index HTML explicit redirect",
			path:                 "/index.html",
			expectedStatus:       http.StatusMovedPermanently,
			expectedCacheControl: "no-cache, must-revalidate",
		},
		{
			name:                 "SPA fallback route (e.g. /overview)",
			path:                 "/overview",
			expectedStatus:       http.StatusOK,
			expectedCacheControl: "no-cache, must-revalidate",
		},
		{
			name:                 "Hashed JavaScript asset",
			path:                 "/assets/index-12345.js",
			expectedStatus:       http.StatusOK,
			expectedCacheControl: "public, max-age=31536000, immutable",
		},
		{
			name:                 "Hashed CSS asset",
			path:                 "/assets/style-67890.css",
			expectedStatus:       http.StatusOK,
			expectedCacheControl: "public, max-age=31536000, immutable",
		},
		{
			name:                 "Regular static asset",
			path:                 "/favicon.ico",
			expectedStatus:       http.StatusOK,
			expectedCacheControl: "public, max-age=3600",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rec.Code)
			}

			// Security headers check
			if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Errorf("missing or invalid X-Content-Type-Options")
			}
			if rec.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
				t.Errorf("missing or invalid X-Frame-Options")
			}
			if rec.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
				t.Errorf("missing or invalid Referrer-Policy")
			}

			// Cache-Control check
			if rec.Header().Get("Cache-Control") != tc.expectedCacheControl {
				t.Errorf("path %s: expected Cache-Control %q, got %q", tc.path, tc.expectedCacheControl, rec.Header().Get("Cache-Control"))
			}
		})
	}
}

func TestStaticServerFallbackDir(t *testing.T) {
	tempDir := t.TempDir()
	assetsDir := filepath.Join(tempDir, "assets")
	_ = os.MkdirAll(assetsDir, 0755)

	_ = os.WriteFile(filepath.Join(tempDir, "index.html"), []byte("fallback index"), 0644)
	_ = os.WriteFile(filepath.Join(assetsDir, "bundle-abc.js"), []byte("bundle code"), 0644)

	handler := Handler(nil, tempDir)

	// Test fallback asset
	req := httptest.NewRequest(http.MethodGet, "/assets/bundle-abc.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Errorf("expected immutable cache control, got %s", rec.Header().Get("Cache-Control"))
	}

	// Test fallback SPA route
	reqSPA := httptest.NewRequest(http.MethodGet, "/settings", nil)
	recSPA := httptest.NewRecorder()
	handler.ServeHTTP(recSPA, reqSPA)

	if recSPA.Code != http.StatusOK {
		t.Fatalf("expected 200 for SPA route, got %d", recSPA.Code)
	}
	if recSPA.Header().Get("Cache-Control") != "no-cache, must-revalidate" {
		t.Errorf("expected no-cache for SPA route, got %s", recSPA.Header().Get("Cache-Control"))
	}
}
