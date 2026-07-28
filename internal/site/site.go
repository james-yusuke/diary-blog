package site

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/james-yusuke/diary-blog/internal/content"
	"github.com/james-yusuke/diary-blog/internal/views"
)

type App struct{ store *content.Store }

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	registerStatic(mux)
	mux.HandleFunc("/", a.route)
	return mux
}

func (a *App) route(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/":
		a.home(w, r)
	case r.URL.Path == "/about":
		a.render(w, r, views.AboutPage(a.store.Site))
	case r.URL.Path == "/feed.xml":
		a.feed(w, r)
	case strings.HasPrefix(r.URL.Path, "/posts/"):
		a.post(w, r)
	case strings.HasPrefix(r.URL.Path, "/tags/"):
		a.tag(w, r)
	default:
		a.notFound(w, r)
	}
}

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	a.render(w, r, views.HomePage(a.store.Site, a.store.Search(query), a.store.Tags(), query))
}

func (a *App) post(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/posts/")
	isZenn := strings.HasPrefix(path, "zenn/")
	slug := path
	if isZenn {
		slug = strings.TrimPrefix(path, "zenn/")
	}
	if slug == "" || strings.Contains(slug, "/") {
		a.notFound(w, r)
		return
	}
	post, ok := a.store.Find(slug)
	if !ok {
		a.notFound(w, r)
		return
	}
	if post.IsZenn() != isZenn {
		if post.IsZenn() {
			http.Redirect(w, r, post.URL(), http.StatusMovedPermanently)
			return
		}
		a.notFound(w, r)
		return
	}
	a.render(w, r, views.PostPage(a.store.Site, post, a.store.Related(post)))
}

func (a *App) tag(w http.ResponseWriter, r *http.Request) {
	rawTag := strings.TrimPrefix(r.URL.Path, "/tags/")
	tag, err := url.PathUnescape(rawTag)
	if err != nil || tag == "" || strings.Contains(tag, "/") {
		a.notFound(w, r)
		return
	}
	posts := a.store.PostsForTag(tag)
	if len(posts) == 0 {
		a.notFound(w, r)
		return
	}
	a.render(w, r, views.TagPage(a.store.Site, tag, posts, a.store.Tags()))
}

func (a *App) feed(w http.ResponseWriter, r *http.Request) {
	baseURL := strings.TrimRight(a.store.Site.BaseURL, "/")
	if baseURL == "" {
		baseURL = "http://" + r.Host
	}
	items := make([]rssItem, 0, len(a.store.Posts))
	for _, post := range a.store.Posts {
		items = append(items, rssItem{Title: post.Title, Link: baseURL + post.URL(), GUID: baseURL + post.URL(), Description: post.Summary, PubDate: post.Published.Format(time.RFC1123Z)})
	}
	feed := rss{Version: "2.0", Channel: rssChannel{Title: a.store.Site.Title, Link: baseURL, Description: a.store.Site.Description, Items: items}}
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(feed)
}

func (a *App) notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	a.render(w, r, views.NotFoundPage(a.store.Site))
}

func (a *App) render(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(context.Background(), w); err != nil {
		fmt.Printf("render page: %v\n", err)
	}
}

type rss struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}
type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
}
type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}
