package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateEmbedsDiaryAndZennContent(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"config", "content/posts", "content/zenn/articles", "assets", "internal/content", "internal/site"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"config/site.yaml":                   "title: diary\nauthor_name: James\ndescription: test\n",
		"assets/site.css":                    "body {}",
		"assets/tokens.css":                  ":root {}",
		"assets/mermaid.js":                  "export default {};",
		"assets/site.js":                     "export default {};",
		"internal/site/ads.txt":              "adm.shinobi.jp,231582,DIRECT\n",
		"content/posts/diary.md":             "---\nslug: diary-post\ntitle: Diary\nsummary: Summary\npublished_at: 2026-01-01\ntags:\n  - Go\n---\n\nDiary body.",
		"content/posts/rfc3339.md":           "---\nslug: rfc3339-post\ntitle: RFC3339\nsummary: Summary\npublished_at: \"2026-07-28T14:38:00+09:00\"\ntags:\n  - Go\n---\n\nRFC3339 body.",
		"content/zenn/articles/zenn_post.md": "---\ntitle: Zenn\ntopics:\n  - Zenn\npublished: true\npublished_at: \"2999-01-02 09:00\"\n---\n\nZenn body.",
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := generate(root); err != nil {
		t.Fatal(err)
	}
	generated, err := os.ReadFile(filepath.Join(root, "internal", "content", "embedded_tinygo_gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{`Source: "diary", Slug: "diary-post"`, `Source: "diary", Slug: "rfc3339-post"`, `Source: "zenn", Slug: "zenn_post"`, `time.FixedZone("UTC+09:00", 32400)`} {
		if !strings.Contains(string(generated), fragment) {
			t.Fatalf("generated content missing %q: %s", fragment, generated)
		}
	}
	assets, err := os.ReadFile(filepath.Join(root, "internal", "site", "assets_tinygo_gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{`workerTokensCSS`, `workerMermaidJS`, `workerSiteJS`, `workerAdsTXT`, `adm.shinobi.jp,231582,DIRECT`} {
		if !strings.Contains(string(assets), fragment) {
			t.Fatalf("generated Worker assets do not include Mermaid JavaScript: %s", assets)
		}
	}
}

func TestCanonicalZoneName(t *testing.T) {
	for _, tc := range []struct {
		offset int
		want   string
	}{
		{offset: 0, want: "UTC"},
		{offset: 9 * 60 * 60, want: "UTC+09:00"},
		{offset: -(3*60 + 30) * 60, want: "UTC-03:30"},
	} {
		if got := canonicalZoneName(tc.offset); got != tc.want {
			t.Errorf("canonicalZoneName(%d) = %q, want %q", tc.offset, got, tc.want)
		}
	}
}
