package site

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testApp(t *testing.T) *App {
	t.Helper()
	root := t.TempDir()
	configPath := filepath.Join(root, "site.yaml")
	postsDir := filepath.Join(root, "posts")
	if err := os.Mkdir(postsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	config := "title: diary\nauthor_name: James Yusuke\ndescription: test description\nbio: test bio\ngithub_url: https://github.com/james-yusuke\nrepository_url: https://github.com/james-yusuke\nbase_url: https://example.test\n"
	post := "---\nslug: hello-templ\ntitle: Hello templ\nsummary: A note about components\npublished_at: 2026-01-02\ntags:\n  - Go\n  - templ\n---\n\n# Hello\n\nA useful body."
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(postsDir, "hello.md"), []byte(post), 0o644); err != nil {
		t.Fatal(err)
	}
	app, err := New(configPath, postsDir)
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func TestRoutesRenderExpectedPages(t *testing.T) {
	app := testApp(t)
	for _, tc := range []struct{ path, contains string }{
		{"/", "Hello templ"},
		{"/?q=components", "検索結果"},
		{"/posts/hello-templ", "A useful body"},
		{"/tags/Go", "ほかのテーマ"},
		{"/about", "James Yusuke"},
		{"/posts/hello-templ", "/assets/mermaid.js"},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		rec := httptest.NewRecorder()
		app.Handler().ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s status = %d", tc.path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), tc.contains) {
			t.Errorf("GET %s missing %q", tc.path, tc.contains)
		}
	}
}

func TestNotFoundAndRSS(t *testing.T) {
	app := testApp(t)
	for _, path := range []string{"/posts/nope", "/tags/nope"} {
		rec := httptest.NewRecorder()
		app.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s status = %d, want 404", path, rec.Code)
		}
	}
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/feed.xml", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Header().Get("Content-Type"), "application/rss+xml") {
		t.Fatalf("RSS response = %d, %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(rec.Body.String(), "<rss version=\"2.0\"") || !strings.Contains(rec.Body.String(), "Hello templ") {
		t.Fatalf("unexpected RSS: %s", rec.Body.String())
	}
}
