package main

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

type LionWiki struct {
	s *Settings
	t *Template
}

func NewLionWiki(s *Settings) *LionWiki {
	return &LionWiki{s, NewTemplate()}
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

	if s.Action == ActionEdit || s.Preview {
		lw.Edit(s)
	}
	lw.t.Render(w, s)
	w.Write([]byte("Not Implemented"))
}

func (lw *LionWiki) Run() error {
	if err := lw.createDirectories(lw.s); err != nil {
		return err
	}
	// Load Plugins
	// plugin('pluginsLoaded');
	if err := lw.t.Load(lw.s.Template); err != nil {
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

func (lw *LionWiki) Edit(s *Session) {
	s.ConFormBegin = `<form action="" method="post">
	<input type="hidden" name="action" value="save"/>
	<input type="hidden" name="last_changed" value="$last_changed_ts"/>
	<input type="hidden" name="showsource" value="$showsource"/>
	<input type="hidden" name="par" value="".h($par).""/>
	<input type="hidden" name="page" value="".h($page).""/>`
	s.ConFormEnd = "</form>"
	s.ConTextarea = `<textarea class="contentTextarea" name="content" style="width:100%" cols="100" rows="30">'.h(str_replace("&lt;", "<", $CON)).'</textarea>`
	s.ConPreview = `<input class="submit" type="submit" name="preview" value="'.$T_PREVIEW.'"/>`

	if s.ShowSource {
		s.ConSubmit = `<input class="submit" type="submit" value="'.$T_DONE.'"/>`
		s.EditSummaryText = s.t.Get("T_EDIT_SUMMARY")
		s.EditSummary = `<input type="text" name="esum" value="'.h($esum).'"/>`

		// if(!authentified()) { // if not logged on, require password
		// 	$FORM_PASSWORD = $T_PASSWORD;
		// 	$FORM_PASSWORD_INPUT = '<input type="password" name="sc"/>';
		// }

		// if(!$par) {
		// 	$RENAME_TEXT = $T_MOVE_TEXT;
		// 	$RENAME_INPUT = '<input type="text" name="moveto" value="'.h($page).'"/>';
		// }
	}

	if s.Preview {
		s.Title = s.t.Get("T_PREVIEW") + s.Page
	}

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
