package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCheckAcceptsPublishedScheduledAndDraftZennArticles(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"config", "content/zenn/articles"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "config", "site.yaml"), []byte("title: diary\nauthor_name: James\ndescription: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	posts := map[string]string{
		"public.md":    "---\ntitle: Public\npublished: true\npublished_at: \"2026-01-02 09:00\"\n---\n\nPublic article.",
		"scheduled.md": "---\ntitle: Scheduled\npublished: true\npublished_at: \"2026-01-04 09:00\"\n---\n\nScheduled article.",
		"draft.md":     "---\ntitle: Draft\npublished: false\n---\n\nDraft article.",
	}
	for name, body := range posts {
		if err := os.WriteFile(filepath.Join(root, "content", "zenn", "articles", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := check(root, time.Date(2026, time.January, 3, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
}

func TestCheckRejectsInvalidPublishedZennArticle(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"config", "content/zenn/articles"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "config", "site.yaml"), []byte("title: diary\nauthor_name: James\ndescription: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "content", "zenn", "articles", "missing-date.md"), []byte("---\ntitle: Missing date\npublished: true\n---\n\nArticle."), 0o644); err != nil {
		t.Fatal(err)
	}

	err := check(root, time.Now())
	if err == nil || !strings.Contains(err.Error(), "published_at is required") {
		t.Fatalf("check error = %v", err)
	}
}
