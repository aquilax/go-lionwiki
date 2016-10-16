package main

import (
	"net/http"
	"strings"
)

type LionWiki struct{}

func NewLionWiki() *LionWiki {
	return &LionWiki{}
}

func (lw *LionWiki) wikiHandler(w http.ResponseWriter, r *http.Request) {
	session := NewSession(r)
	if session.Erasecookie {
		eraseCookies(w, r)
	}
	w.Write([]byte("Not Implemented"))
}

func (lw *LionWiki) Run(s *Settings) error {
	if err := lw.createDirectories(s); err != nil {
		return err
	}
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", lw.wikiHandler)
	return http.ListenAndServe(":9090", nil)
}

func (lw *LionWiki) createDirectories(s *Settings) error {
	// TODO index.php:78
	return nil
}

func eraseCookies(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		if strings.HasPrefix(cookie.Name, CookiePrefix) {
			// Clear cookie
			cookie.MaxAge = -1
			cookie.Value = ""
			http.SetCookie(w, cookie)
		}
	}
}
