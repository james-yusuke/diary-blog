//go:build tinygo

package site

import (
	"syscall/js"

	"github.com/james-yusuke/diary-blog/internal/content"
)

func NewWorker() *App {
	now := timeFromUnixMilliseconds(js.Global().Get("Date").Call("now").Float())
	return &App{store: content.LoadEmbeddedAt(now)}
}
