package content

import (
	"sort"
	"strings"
	"time"
)

type SiteConfig struct {
	Title         string `yaml:"title"`
	AuthorName    string `yaml:"author_name"`
	Description   string `yaml:"description"`
	Bio           string `yaml:"bio"`
	GitHubURL     string `yaml:"github_url"`
	RepositoryURL string `yaml:"repository_url"`
	BaseURL       string `yaml:"base_url"`
}

type Post struct {
	Source      string
	Slug        string
	Title       string
	Summary     string
	Published   time.Time
	Tags        []string
	Body        string
	HTML        string
	ReadingMins int
}

const (
	SourceDiary = "diary"
	SourceZenn  = "zenn"
)

func (p Post) URL() string {
	if p.Source == SourceZenn {
		return "/posts/zenn/" + p.Slug
	}
	return "/posts/" + p.Slug
}

func (p Post) IsZenn() bool { return p.Source == SourceZenn }
func (p Post) SourceLabel() string {
	if p.IsZenn() {
		return "Zenn"
	}
	return "独自記事"
}
func (p Post) DisplayDate() string { return p.Published.Format("2006.01.02") }

type Tag struct {
	Name  string
	Count int
}

type Store struct {
	Site   SiteConfig
	Posts  []Post
	bySlug map[string]Post
	byTag  map[string][]Post
}

func NewStore(site SiteConfig, posts []Post) *Store {
	return NewStoreAt(site, posts, time.Now())
}

// NewStoreAt builds the public post indexes as of now. It is explicit so
// scheduled-post behavior can be tested without relying on the system clock.
func NewStoreAt(site SiteConfig, posts []Post, now time.Time) *Store {
	visiblePosts := make([]Post, 0, len(posts))
	for _, post := range posts {
		if !post.Published.After(now) {
			visiblePosts = append(visiblePosts, post)
		}
	}
	return newStore(site, visiblePosts)
}

func newStore(site SiteConfig, posts []Post) *Store {
	store := &Store{Site: site, Posts: posts, bySlug: make(map[string]Post), byTag: make(map[string][]Post)}
	sort.Slice(store.Posts, func(i, j int) bool { return store.Posts[i].Published.After(store.Posts[j].Published) })
	for _, post := range store.Posts {
		store.bySlug[post.Slug] = post
		for _, tag := range post.Tags {
			store.byTag[tag] = append(store.byTag[tag], post)
		}
	}
	for tag := range store.byTag {
		sort.Slice(store.byTag[tag], func(i, j int) bool { return store.byTag[tag][i].Published.After(store.byTag[tag][j].Published) })
	}
	return store
}

func (s *Store) Find(slug string) (Post, bool) { post, ok := s.bySlug[slug]; return post, ok }
func (s *Store) Search(query string) []Post {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return s.Posts
	}
	var matches []Post
	for _, post := range s.Posts {
		haystack := strings.ToLower(post.Title + " " + post.Summary + " " + strings.Join(post.Tags, " ") + " " + post.Body)
		if strings.Contains(haystack, query) {
			matches = append(matches, post)
		}
	}
	return matches
}
func (s *Store) PostsForTag(tag string) []Post { return s.byTag[tag] }
func (s *Store) Tags() []Tag {
	tags := make([]Tag, 0, len(s.byTag))
	for name, posts := range s.byTag {
		tags = append(tags, Tag{Name: name, Count: len(posts)})
	}
	sort.Slice(tags, func(i, j int) bool {
		if tags[i].Count == tags[j].Count {
			return tags[i].Name < tags[j].Name
		}
		return tags[i].Count > tags[j].Count
	})
	return tags
}
func (s *Store) Related(post Post) []Post {
	var related []Post
	for _, candidate := range s.Posts {
		if candidate.Slug != post.Slug && sharesTag(post, candidate) {
			related = append(related, candidate)
			if len(related) == 3 {
				break
			}
		}
	}
	return related
}
func sharesTag(a, b Post) bool {
	for _, tagA := range a.Tags {
		for _, tagB := range b.Tags {
			if tagA == tagB {
				return true
			}
		}
	}
	return false
}
