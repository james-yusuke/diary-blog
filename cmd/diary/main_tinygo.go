//go:build tinygo

package main

import (
	"github.com/james-yusuke/diary-blog/internal/site"
	"github.com/syumai/workers"
)

func main() {
	workers.Serve(site.NewWorker().Handler())
}
