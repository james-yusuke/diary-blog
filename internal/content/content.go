//go:build !tinygo

package content

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"gopkg.in/yaml.v3"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type postMeta struct {
	Slug      string    `yaml:"slug"`
	Title     string    `yaml:"title"`
	Summary   string    `yaml:"summary"`
	Published time.Time `yaml:"published_at"`
	Tags      []string  `yaml:"tags"`
	Draft     bool      `yaml:"draft"`
}

func Load(configPath, postsDir string) (*Store, error) {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read site configuration: %w", err)
	}
	var site SiteConfig
	if err := yaml.Unmarshal(configBytes, &site); err != nil {
		return nil, fmt.Errorf("parse site configuration: %w", err)
	}
	if site.Title == "" || site.AuthorName == "" || site.Description == "" {
		return nil, errors.New("site configuration requires title, author_name, and description")
	}

	entries, err := os.ReadDir(postsDir)
	if err != nil {
		return nil, fmt.Errorf("read posts directory: %w", err)
	}
	var posts []Post
	seenSlugs := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		post, draft, err := parsePost(filepath.Join(postsDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		if draft {
			continue
		}
		if seenSlugs[post.Slug] {
			return nil, fmt.Errorf("duplicate post slug %q", post.Slug)
		}
		seenSlugs[post.Slug] = true
		posts = append(posts, post)
	}
	return NewStore(site, posts), nil
}

func parsePost(path string) (Post, bool, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Post{}, false, fmt.Errorf("read post %s: %w", path, err)
	}
	metaRaw, body, err := splitFrontMatter(string(contents))
	if err != nil {
		return Post{}, false, fmt.Errorf("parse post %s: %w", path, err)
	}
	var meta postMeta
	if err := yaml.Unmarshal([]byte(metaRaw), &meta); err != nil {
		return Post{}, false, fmt.Errorf("parse post metadata %s: %w", path, err)
	}
	if err := validateMeta(meta); err != nil {
		return Post{}, false, fmt.Errorf("validate post %s: %w", path, err)
	}
	var rendered bytes.Buffer
	markdown := goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithParserOptions(parser.WithAutoHeadingID()))
	if err := markdown.Convert([]byte(body), &rendered); err != nil {
		return Post{}, false, fmt.Errorf("render post %s: %w", path, err)
	}
	return Post{
		Slug: meta.Slug, Title: meta.Title, Summary: meta.Summary, Published: meta.Published,
		Tags: meta.Tags, Body: body, HTML: rendered.String(), ReadingMins: readingMins(body),
	}, meta.Draft, nil
}

func splitFrontMatter(contents string) (string, string, error) {
	contents = strings.ReplaceAll(contents, "\r\n", "\n")
	if !strings.HasPrefix(contents, "---\n") {
		return "", "", errors.New("front matter must start with ---")
	}
	end := strings.Index(contents[4:], "\n---\n")
	if end < 0 {
		return "", "", errors.New("front matter must end with ---")
	}
	end += 4
	return contents[4:end], contents[end+5:], nil
}

func validateMeta(meta postMeta) error {
	if !slugPattern.MatchString(meta.Slug) {
		return errors.New("slug must be lowercase kebab-case")
	}
	if strings.TrimSpace(meta.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(meta.Summary) == "" {
		return errors.New("summary is required")
	}
	if meta.Published.IsZero() {
		return errors.New("published_at is required")
	}
	if len(meta.Tags) == 0 {
		return errors.New("at least one tag is required")
	}
	seen := make(map[string]bool)
	for _, tag := range meta.Tags {
		if strings.TrimSpace(tag) == "" || seen[tag] {
			return errors.New("tags must be non-empty and unique")
		}
		seen[tag] = true
	}
	return nil
}

func readingMins(body string) int {
	count := utf8.RuneCountInString(strings.TrimSpace(body))
	if count == 0 {
		return 1
	}
	return max(1, (count+499)/500)
}
