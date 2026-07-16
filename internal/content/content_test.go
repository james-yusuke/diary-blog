package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFixture(t *testing.T, posts map[string]string) (string, string) {
	t.Helper()
	root := t.TempDir()
	configPath := filepath.Join(root, "site.yaml")
	postsDir := filepath.Join(root, "posts")
	if err := os.Mkdir(postsDir, 0o755); err != nil { t.Fatal(err) }
	config := "title: diary\nauthor_name: James Yusuke\ndescription: test site\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil { t.Fatal(err) }
	for name, body := range posts {
		if err := os.WriteFile(filepath.Join(postsDir, name), []byte(body), 0o644); err != nil { t.Fatal(err) }
	}
	return configPath, postsDir
}

func post(slug, date, tags, body string) string {
	return "---\nslug: " + slug + "\ntitle: " + slug + " title\nsummary: " + slug + " summary\npublished_at: " + date + "\ntags:\n" + tags + "---\n\n" + body
}

func TestLoadSortsPostsExcludesDraftsAndBuildsTags(t *testing.T) {
	posts := map[string]string{
		"older.md": post("older", "2026-01-02", "  - Go\n", "older body"),
		"newer.md": post("newer", "2026-01-03", "  - Go\n  - Web\n", "newer body"),
		"draft.md": "---\nslug: draft\ntitle: draft\nsummary: draft\npublished_at: 2026-01-04\ntags:\n  - Go\ndraft: true\n---\n\nnot published",
	}
	configPath, postsDir := writeFixture(t, posts)
	store, err := Load(configPath, postsDir)
	if err != nil { t.Fatal(err) }
	if len(store.Posts) != 2 { t.Fatalf("posts = %d, want 2", len(store.Posts)) }
	if store.Posts[0].Slug != "newer" { t.Fatalf("first post = %q, want newer", store.Posts[0].Slug) }
	if got := store.PostsForTag("Go"); len(got) != 2 { t.Fatalf("Go posts = %d, want 2", len(got)) }
	tags := store.Tags()
	if tags[0].Name != "Go" || tags[0].Count != 2 { t.Fatalf("first tag = %#v", tags[0]) }
}

func TestSearchFindsTitleTagsAndBody(t *testing.T) {
	configPath, postsDir := writeFixture(t, map[string]string{
		"alpha.md": post("alpha", "2026-01-02", "  - Go\n", "templ components are pleasant"),
		"beta.md": post("beta", "2026-01-01", "  - Design\n", "layout rhythm"),
	})
	store, err := Load(configPath, postsDir)
	if err != nil { t.Fatal(err) }
	if got := store.Search("templ"); len(got) != 1 || got[0].Slug != "alpha" { t.Fatalf("body search = %#v", got) }
	if got := store.Search("design"); len(got) != 1 || got[0].Slug != "beta" { t.Fatalf("tag search = %#v", got) }
	if got := store.Search("  "); len(got) != 2 { t.Fatalf("blank search = %d", len(got)) }
}

func TestLoadRejectsInvalidAndDuplicateMetadata(t *testing.T) {
	configPath, postsDir := writeFixture(t, map[string]string{
		"one.md": post("same-slug", "2026-01-02", "  - Go\n", "one"),
		"two.md": post("same-slug", "2026-01-01", "  - Web\n", "two"),
	})
	if _, err := Load(configPath, postsDir); err == nil || !strings.Contains(err.Error(), "duplicate") { t.Fatalf("duplicate error = %v", err) }

	configPath, postsDir = writeFixture(t, map[string]string{
		"bad.md": post("Not Allowed", "2026-01-01", "  - Go\n", "bad"),
	})
	if _, err := Load(configPath, postsDir); err == nil || !strings.Contains(err.Error(), "kebab-case") { t.Fatalf("invalid slug error = %v", err) }
}
