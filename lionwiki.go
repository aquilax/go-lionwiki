package main

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

type LionWiki struct {
	s *Settings
}

func NewLionWiki(s *Settings) *LionWiki {
	return &LionWiki{s}
}

func (lw *LionWiki) wikiHandler(w http.ResponseWriter, r *http.Request) {
	s := NewSession(r)
	if s.Erasecookie {
		eraseCookies(w, r)
	}
	// plugin('actionBegin');

	if len(s.Action) == 0 {
		if len(s.Page) == 0 {
			v := url.Values{}
			v.Set("page", lw.s.StartPage)
			http.Redirect(w, r, "/?"+v.Encode(), http.StatusMovedPermanently)
			return
		}
		// language variant
		if fileExists(lw.s.PgDir + s.Page + "." + s.Lang + ".txt") {
			v := url.Values{}
			v.Set("page", s.Page+"."+s.Lang)
			http.Redirect(w, r, "/?"+v.Encode(), http.StatusFound)
			return
		}
		// create page if it doesn't exist
		if !fileExists(lw.s.PgDir + s.Page + ".txt") {
			s.Action = ActionEdit
		}
	}

	w.Write([]byte("Not Implemented"))
}

func (lw *LionWiki) Run() error {
	if err := lw.createDirectories(lw.s); err != nil {
		return err
	}
	// Load Plugins
	// plugin('pluginsLoaded');
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

func fileExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}
