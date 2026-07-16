//go:build !tinygo

package site

import "net/http"

func registerStatic(mux *http.ServeMux) {
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
}
