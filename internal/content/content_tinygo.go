//go:build tinygo

package content

import "time"

// LoadEmbedded returns the content snapshot generated from the repository's
// Markdown and YAML sources immediately before a Worker build.
func LoadEmbedded() *Store { return LoadEmbeddedAt(time.Now()) }

// LoadEmbeddedAt returns the public view of the generated snapshot at now.
// The Worker passes its JavaScript wall clock here because TinyGo's wasm
// time.Now is monotonic time since startup, not a Unix wall clock.
func LoadEmbeddedAt(now time.Time) *Store { return NewStoreAt(embeddedSite, embeddedPosts, now) }
