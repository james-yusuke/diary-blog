//go:build !tinygo

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/james-yusuke/diary-blog/internal/site"
)

func main() {
	app, err := site.New("config/site.yaml", "content")
	if err != nil {
		log.Fatal(err)
	}
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	fmt.Printf("diary is running at http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, app.Handler()))
}
