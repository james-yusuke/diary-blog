//go:build !tinygo

package site

import (
	_ "embed"
	"net/http"
)

//go:embed ads.txt
var adsTXT string

func registerStatic(mux *http.ServeMux) {
	mux.HandleFunc("/ads.txt", serveAdsTXT)
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
}

func serveAdsTXT(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(adsTXT))
}
