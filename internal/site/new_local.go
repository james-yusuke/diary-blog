//go:build !tinygo

package site

import "github.com/james-yusuke/diary-blog/internal/content"

func New(configPath, postsDir string) (*App, error) {
	store, err := content.Load(configPath, postsDir)
	if err != nil {
		return nil, err
	}
	return &App{store: store}, nil
}
