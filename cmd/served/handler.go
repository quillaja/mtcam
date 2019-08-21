package main

import (
	"net/http"
)

func CreateHandler(cfg *ServerdConfig) http.Handler {
	mux := http.NewServeMux()
	frontend := http.NewServeMux()
	api := http.NewServeMux()

	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from the api: " + r.URL.String()))
	})

	frontend.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world: " + r.RemoteAddr + r.RequestURI))
	})

	mux.Handle("/api/", api)
	mux.Handle("/", http.FileServer(http.Dir(cfg.StaticRoot)))

	return mux
}
