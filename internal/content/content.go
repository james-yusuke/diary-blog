//go:build !tinygo

package content

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"gopkg.in/yaml.v3"
)

var (
	slugPattern         = regexp.MustCompile(`^[a-z0-9]+(?:[-_][a-z0-9]+)*$`)
	markdownLinkPattern = regexp.MustCompile(`!?\[([^\]]+)\]\([^)]*\)`)
)

type postMeta struct {
	Slug      string    `yaml:"slug"`
	Title     string    `yaml:"title"`
	Summary   string    `yaml:"summary"`
	Published time.Time `yaml:"published_at"`
	Tags      []string  `yaml:"tags"`
	Draft     bool      `yaml:"draft"`
}

type zennMeta struct {
	Title       string   `yaml:"title"`
	Topics      []string `yaml:"topics"`
	Published   *bool    `yaml:"published"`
	PublishedAt string   `yaml:"published_at"`
}

type ZennArticleState string

const (
	ZennArticlePublic    ZennArticleState = "public"
	ZennArticleScheduled ZennArticleState = "scheduled"
	ZennArticleDraft     ZennArticleState = "draft"
)

// ZennArticleAudit describes how one Markdown article under zenn/articles is
// expected to be handled by the diary. It is deliberately based on the same
// parser used by LoadAll so the CI report cannot disagree with the build.
type ZennArticleAudit struct {
	Slug      string
	Path      string
	State     ZennArticleState
	Published time.Time
}

// Load reads diary posts from contentDir. Markdown below zenn/articles uses
// Zenn's front matter; all other Markdown uses diary's front matter.
func Load(configPath, contentDir string) (*Store, error) {
	site, posts, err := loadSiteAndPosts(configPath, contentDir)
	if err != nil {
		return nil, err
	}
	return NewStore(site, posts), nil
}

// LoadAll reads every non-draft post, including scheduled posts. It is used by
// the Worker generator so a post can become public without another deploy.
func LoadAll(configPath, contentDir string) (*Store, error) {
	site, posts, err := loadSiteAndPosts(configPath, contentDir)
	if err != nil {
		return nil, err
	}
	return newStore(site, posts), nil
}

// AuditZenn classifies every Zenn article. A draft is intentionally absent
// from the Worker snapshot; published and scheduled articles must be present
// in it so they can become visible without another deployment.
func AuditZenn(contentDir string, now time.Time) ([]ZennArticleAudit, error) {
	articlesRoot := filepath.Join(contentDir, "zenn", "articles")
	if _, err := os.Stat(articlesRoot); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("read Zenn articles: %w", err)
	}

	var audit []ZennArticleAudit
	err := filepath.WalkDir(articlesRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			return nil
		}
		post, draft, err := parseZennPost(path)
		if err != nil {
			return err
		}
		slug := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		entryAudit := ZennArticleAudit{Slug: slug, Path: path}
		if draft {
			entryAudit.State = ZennArticleDraft
		} else {
			entryAudit.Published = post.Published
			if post.Published.After(now) {
				entryAudit.State = ZennArticleScheduled
			} else {
				entryAudit.State = ZennArticlePublic
			}
		}
		audit = append(audit, entryAudit)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("audit Zenn content: %w", err)
	}
	sort.Slice(audit, func(i, j int) bool { return audit[i].Slug < audit[j].Slug })
	return audit, nil
}

func loadSiteAndPosts(configPath, contentDir string) (SiteConfig, []Post, error) {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return SiteConfig{}, nil, fmt.Errorf("read site configuration: %w", err)
	}
	var site SiteConfig
	if err := yaml.Unmarshal(configBytes, &site); err != nil {
		return SiteConfig{}, nil, fmt.Errorf("parse site configuration: %w", err)
	}
	if site.Title == "" || site.AuthorName == "" || site.Description == "" {
		return SiteConfig{}, nil, errors.New("site configuration requires title, author_name, and description")
	}

	posts, err := loadPosts(contentDir)
	if err != nil {
		return SiteConfig{}, nil, err
	}
	return site, posts, nil
}

func loadPosts(contentDir string) ([]Post, error) {
	var posts []Post
	seenSlugs := make(map[string]bool)
	zennRoot := filepath.Join("zenn")
	zennArticles := filepath.Join(zennRoot, "articles")
	err := filepath.WalkDir(contentDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			return nil
		}
		rel, err := filepath.Rel(contentDir, path)
		if err != nil {
			return err
		}
		isZenn := rel == zennArticles || strings.HasPrefix(rel, zennArticles+string(filepath.Separator))
		if strings.HasPrefix(rel, zennRoot+string(filepath.Separator)) && !isZenn {
			return nil
		}

		var post Post
		var draft bool
		if isZenn {
			post, draft, err = parseZennPost(path)
		} else {
			post, draft, err = parsePost(path)
		}
		if err != nil {
			return err
		}
		if draft {
			return nil
		}
		if seenSlugs[post.Slug] {
			return fmt.Errorf("duplicate post slug %q", post.Slug)
		}
		seenSlugs[post.Slug] = true
		posts = append(posts, post)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read content: %w", err)
	}
	return posts, nil
}

func parsePost(path string) (Post, bool, error) {
	metaRaw, body, err := readFrontMatter(path)
	if err != nil {
		return Post{}, false, err
	}
	var meta postMeta
	if err := yaml.Unmarshal([]byte(metaRaw), &meta); err != nil {
		return Post{}, false, fmt.Errorf("parse post metadata %s: %w", path, err)
	}
	if err := validateMeta(meta); err != nil {
		return Post{}, false, fmt.Errorf("validate post %s: %w", path, err)
	}
	post, err := renderPost(meta.Slug, meta.Title, meta.Summary, meta.Published, meta.Tags, body)
	post.Source = SourceDiary
	return post, meta.Draft, err
}

func parseZennPost(path string) (Post, bool, error) {
	metaRaw, body, err := readFrontMatter(path)
	if err != nil {
		return Post{}, false, err
	}
	var meta zennMeta
	if err := yaml.Unmarshal([]byte(metaRaw), &meta); err != nil {
		return Post{}, false, fmt.Errorf("parse Zenn metadata %s: %w", path, err)
	}
	if err := validateZennMeta(meta, false); err != nil {
		return Post{}, false, fmt.Errorf("validate Zenn post %s: %w", path, err)
	}
	if !*meta.Published {
		return Post{}, true, nil
	}
	if err := validateZennMeta(meta, true); err != nil {
		return Post{}, false, fmt.Errorf("validate Zenn post %s: %w", path, err)
	}
	published, err := parseZennPublishedAt(meta.PublishedAt)
	if err != nil {
		return Post{}, false, fmt.Errorf("validate Zenn post %s: %w", path, err)
	}
	summary, err := zennSummary(body)
	if err != nil {
		return Post{}, false, fmt.Errorf("validate Zenn post %s: %w", path, err)
	}
	slug := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if !slugPattern.MatchString(slug) {
		return Post{}, false, fmt.Errorf("validate Zenn post %s: filename must be lowercase letters, digits, hyphens, or underscores", path)
	}
	post, err := renderPost(slug, meta.Title, summary, published, meta.Topics, body)
	post.Source = SourceZenn
	return post, false, err
}

func readFrontMatter(path string) (string, string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("read post %s: %w", path, err)
	}
	metaRaw, body, err := splitFrontMatter(string(contents))
	if err != nil {
		return "", "", fmt.Errorf("parse post %s: %w", path, err)
	}
	return metaRaw, body, nil
}

func renderPost(slug, title, summary string, published time.Time, tags []string, body string) (Post, error) {
	var rendered bytes.Buffer
	markdown := goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Footnote), goldmark.WithParserOptions(parser.WithAutoHeadingID()))
	if err := markdown.Convert([]byte(body), &rendered); err != nil {
		return Post{}, fmt.Errorf("render post %s: %w", slug, err)
	}
	return Post{
		Slug: slug, Title: title, Summary: summary, Published: published,
		Tags: tags, Body: body, HTML: rendered.String(), ReadingMins: readingMins(body),
	}, nil
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
		return errors.New("slug must use lowercase letters, digits, hyphens, or underscores")
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
	return validateTags(meta.Tags)
}

func validateZennMeta(meta zennMeta, requirePublishedAt bool) error {
	if strings.TrimSpace(meta.Title) == "" {
		return errors.New("title is required")
	}
	if meta.Published == nil {
		return errors.New("published is required")
	}
	if requirePublishedAt && strings.TrimSpace(meta.PublishedAt) == "" {
		return errors.New("published_at is required")
	}
	return validateTags(meta.Topics)
}

func validateTags(tags []string) error {
	seen := make(map[string]bool)
	for _, tag := range tags {
		if strings.TrimSpace(tag) == "" || seen[tag] {
			return errors.New("tags must be non-empty and unique")
		}
		seen[tag] = true
	}
	return nil
}

func parseZennPublishedAt(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Time{}, fmt.Errorf("load JST location: %w", err)
	}
	for _, layout := range []string{"2006-01-02", "2006-01-02 15:04"} {
		if parsed, err := time.ParseInLocation(layout, value, location); err == nil {
			return parsed, nil
		}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, errors.New("published_at must be YYYY-MM-DD, YYYY-MM-DD HH:MM (JST), or RFC3339")
	}
	return parsed, nil
}

func zennSummary(body string) (string, error) {
	for _, paragraph := range strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n\n") {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" || strings.HasPrefix(paragraph, "#") || strings.HasPrefix(paragraph, "```") || strings.HasPrefix(paragraph, ":::") {
			continue
		}
		summary := markdownLinkPattern.ReplaceAllString(paragraph, "$1")
		summary = strings.NewReplacer("**", "", "__", "", "`", "", "*", "").Replace(summary)
		summary = strings.Join(strings.Fields(summary), " ")
		if summary == "" {
			continue
		}
		runes := []rune(summary)
		if len(runes) > 160 {
			return string(runes[:157]) + "...", nil
		}
		return summary, nil
	}
	return "", errors.New("body requires a non-heading paragraph for summary")
}

func readingMins(body string) int {
	count := utf8.RuneCountInString(strings.TrimSpace(body))
	if count == 0 {
		return 1
	}
	return max(1, (count+499)/500)
}
