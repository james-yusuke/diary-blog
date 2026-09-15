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
	contentDir := filepath.Join(root, "content")
	postsDir := filepath.Join(contentDir, "posts")
	zennDir := filepath.Join(contentDir, "zenn", "articles")
	if err := os.MkdirAll(postsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(zennDir, 0o755); err != nil {
		t.Fatal(err)
	}
	config := "title: diary\nauthor_name: James Yusuke\ndescription: test description\nbio: test bio\ngithub_url: https://github.com/james-yusuke\nrepository_url: https://github.com/james-yusuke\nbase_url: https://example.test\n"
	post := "---\nslug: hello-templ\ntitle: Hello templ\nsummary: A note about components\npublished_at: 2026-01-02\ntags:\n  - Go\n  - templ\n---\n\n# Hello\n\nA useful body."
	zennPost := "---\ntitle: Zenn note\ntopics:\n  - Go\npublished: true\npublished_at: \"2026-01-03 09:00\"\n---\n\n# Zenn\n\nA Zenn body."
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(postsDir, "hello.md"), []byte(post), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(zennDir, "zenn-note.md"), []byte(zennPost), 0o644); err != nil {
		t.Fatal(err)
	}
	app, err := New(configPath, contentDir)
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func TestRoutesRenderExpectedPages(t *testing.T) {
	app := testApp(t)
	for _, tc := range []struct{ path, contains string }{
		{"/", "Hello templ"},
		{"/", "https://avatars.githubusercontent.com/u/238946603?v=4"},
		{"/", "data-theme-toggle"},
		{"/", "data-theme-label"},
		{"/", "独自記事"},
		{"/", "Zenn"},
		{"/?q=components", "検索結果"},
		{"/posts/hello-templ", "A useful body"},
		{"/posts/hello-templ", "独自記事"},
		{"/posts/zenn/zenn-note", "A Zenn body"},
		{"/posts/zenn/zenn-note", "Zenn"},
		{"/tags/Go", "ほかのテーマ"},
		{"/about", "James Yusuke"},
		{"/about", "contact@yecov.com"},
		{"/about", "Web開発やGoを中心とした開発案件のご依頼を受け付けています"},
		{"/posts/hello-templ", "/assets/mermaid.js"},
		{"/posts/hello-templ", "https://adm.shinobi.jp/s/aec20509a6f514d325d9e767b7573e90"},
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
	for _, path := range []string{"/posts/nope", "/posts/zenn/hello-templ", "/tags/nope"} {
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
	if !strings.Contains(rec.Body.String(), "/posts/zenn/zenn-note") {
		t.Fatalf("RSS does not use the Zenn URL: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "<category>独自記事</category>") || !strings.Contains(rec.Body.String(), "<category>Zenn</category>") {
		t.Fatalf("RSS does not identify post sources: %s", rec.Body.String())
	}
}

func TestLegacyZennURLRedirects(t *testing.T) {
	app := testApp(t)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/posts/zenn-note", nil))
	if rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != "/posts/zenn/zenn-note" {
		t.Fatalf("legacy Zenn URL = %d, Location = %q", rec.Code, rec.Header().Get("Location"))
	}
}
