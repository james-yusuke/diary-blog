//go:build tinygo

package content

// LoadEmbedded returns the content snapshot generated from the repository's
// Markdown and YAML sources immediately before a Worker build.
func LoadEmbedded() *Store { return NewStore(embeddedSite, embeddedPosts) }
