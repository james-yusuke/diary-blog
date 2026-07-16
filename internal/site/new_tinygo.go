//go:build tinygo

package site

import "github.com/james-yusuke/diary-blog/internal/content"

func NewWorker() *App { return &App{store: content.LoadEmbedded()} }
