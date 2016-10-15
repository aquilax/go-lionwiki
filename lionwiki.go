package main

import (
	"net/http"
)

type LionWiki struct{}

func NewLionWiki() *LionWiki {
	return &LionWiki{}
}

func (lw *LionWiki) wikiHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Not Implemented"))
}

func (lw *LionWiki) Run() error {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", lw.wikiHandler)
	return http.ListenAndServe(":9090", nil)
}
