package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeFixture(t *testing.T, files map[string]string) (string, string) {
	t.Helper()
	root := t.TempDir()
	configPath := filepath.Join(root, "site.yaml")
	contentDir := filepath.Join(root, "content")
	if err := os.MkdirAll(contentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	config := "title: diary\nauthor_name: James Yusuke\ndescription: test site\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, body := range files {
		path := filepath.Join(contentDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return configPath, contentDir
}

func post(slug, date, tags, body string) string {
	return "---\nslug: " + slug + "\ntitle: " + slug + " title\nsummary: " + slug + " summary\npublished_at: " + date + "\ntags:\n" + tags + "---\n\n" + body
}

func zennPost(title, publishedAt, topics, body string, published bool) string {
	return "---\ntitle: " + title + "\ntopics:\n" + topics + "published: " + map[bool]string{true: "true", false: "false"}[published] + "\npublished_at: \"" + publishedAt + "\"\n---\n\n" + body
}

func TestLoadRecursivelyReadsDiaryAndZennArticles(t *testing.T) {
	configPath, contentDir := writeFixture(t, map[string]string{
		"posts/nested/older.md":                post("older", "2026-01-02", "  - Go\n", "older body"),
		"journal/newer.md":                     post("newer", "2026-01-03", "  - Web\n", "newer body"),
		"zenn/articles/nested/zenn_article.md": zennPost("Zenn article", "2026-01-04 09:00", "  - Go\n  - Zenn\n", "First **Zenn** paragraph.\n\nSecond paragraph.", true),
		"zenn/README.md":                       "this must be ignored",
		"posts/draft.md":                       "---\nslug: draft\ntitle: draft\nsummary: draft\npublished_at: 2026-01-05\ntags:\n  - Go\ndraft: true\n---\n\ndraft body",
	})

	store, err := Load(configPath, contentDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(store.Posts) != 3 {
		t.Fatalf("posts = %d, want 3", len(store.Posts))
	}
	if store.Posts[0].Slug != "zenn_article" {
		t.Fatalf("first post = %q, want zenn_article", store.Posts[0].Slug)
	}
	zenn, ok := store.Find("zenn_article")
	if !ok {
		t.Fatal("Zenn article was not loaded")
	}
	if zenn.Summary != "First Zenn paragraph." {
		t.Fatalf("summary = %q", zenn.Summary)
	}
	if !zenn.IsZenn() || zenn.URL() != "/posts/zenn/zenn_article" {
		t.Fatalf("Zenn URL = %q, source = %q", zenn.URL(), zenn.Source)
	}
	if got := store.PostsForTag("Zenn"); len(got) != 1 || got[0].Slug != "zenn_article" {
		t.Fatalf("Zenn posts = %#v", got)
	}
}

func TestLoadZennAllowsNoTopicsAndRequiresPublishedAt(t *testing.T) {
	configPath, contentDir := writeFixture(t, map[string]string{
		"zenn/articles/no_topics.md": zennPost("No topics", "2026-01-03", "", "A paragraph.", true),
	})
	store, err := Load(configPath, contentDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(store.Posts) != 1 || len(store.Posts[0].Tags) != 0 {
		t.Fatalf("posts = %#v", store.Posts)
	}

	configPath, contentDir = writeFixture(t, map[string]string{
		"zenn/articles/missing-date.md": "---\ntitle: Missing date\npublished: true\n---\n\nA paragraph.",
	})
	if _, err := Load(configPath, contentDir); err == nil || !strings.Contains(err.Error(), "published_at is required") {
		t.Fatalf("missing published_at error = %v", err)
	}

	configPath, contentDir = writeFixture(t, map[string]string{
		"zenn/articles/draft.md": "---\ntitle: Draft\npublished: false\n---\n\nA draft paragraph.",
	})
	store, err = Load(configPath, contentDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(store.Posts) != 0 {
		t.Fatalf("draft posts = %#v", store.Posts)
	}
}

func TestLoadRejectsInvalidAndDuplicateMetadata(t *testing.T) {
	configPath, contentDir := writeFixture(t, map[string]string{
		"posts/one.md":               post("same-slug", "2026-01-02", "  - Go\n", "one"),
		"zenn/articles/same-slug.md": zennPost("Same", "2026-01-01", "  - Web\n", "two", true),
	})
	if _, err := Load(configPath, contentDir); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate error = %v", err)
	}

	configPath, contentDir = writeFixture(t, map[string]string{
		"posts/bad.md": post("Not Allowed", "2026-01-01", "  - Go\n", "bad"),
	})
	if _, err := Load(configPath, contentDir); err == nil || !strings.Contains(err.Error(), "lowercase") {
		t.Fatalf("invalid slug error = %v", err)
	}
}

func TestNewStoreAtExcludesFuturePosts(t *testing.T) {
	now := time.Date(2026, time.January, 2, 9, 0, 0, 0, time.UTC)
	past := Post{Slug: "past", Title: "Past", Summary: "Past", Published: now.Add(-time.Minute)}
	future := Post{Slug: "future", Title: "Future", Summary: "Future", Published: now.Add(time.Minute)}
	store := NewStoreAt(SiteConfig{}, []Post{future, past}, now)
	if len(store.Posts) != 1 || store.Posts[0].Slug != "past" {
		t.Fatalf("visible posts = %#v", store.Posts)
	}
	if _, ok := store.Find("future"); ok {
		t.Fatal("future post should not be reachable")
	}

	store = NewStoreAt(SiteConfig{}, []Post{future}, future.Published)
	if _, ok := store.Find("future"); !ok {
		t.Fatal("post should be visible at published_at")
	}
}

func TestZennSummaryTruncatesTo160Runes(t *testing.T) {
	summary, err := zennSummary(strings.Repeat("あ", 161))
	if err != nil {
		t.Fatal(err)
	}
	if got := len([]rune(summary)); got != 160 || !strings.HasSuffix(summary, "...") {
		t.Fatalf("summary = %q (%d runes)", summary, got)
	}
}

func TestRenderPostSupportsFootnoteReferences(t *testing.T) {
	post, err := renderPost("footnotes", "Footnotes", "Summary", time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), []string{"Go"}, "本文の参照[^go-asm]。\n\n[^go-asm]: [Go assembler guide](https://go.dev/doc/asm)")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(post.HTML, "[^go-asm]") {
		t.Fatalf("footnote reference was left as literal text: %s", post.HTML)
	}
	if !strings.Contains(post.HTML, `class="footnote-ref"`) || !strings.Contains(post.HTML, "Go assembler guide") {
		t.Fatalf("footnote was not rendered as a reference and definition: %s", post.HTML)
	}
}
