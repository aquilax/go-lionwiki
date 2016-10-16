package main

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type LionWiki struct {
	st *Settings
	t  *Template
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
			http.Redirect(w, r, s.Self+"?page="+u(lw.st.StartPage), http.StatusMovedPermanently)
			return
		}
		// language variant
		if fileExists(lw.st.PgDir + s.Page + "." + s.Lang + ".txt") {
			v := url.Values{}
			v.Set("page", s.Page+"."+s.Lang)
			http.Redirect(w, r, "/?"+v.Encode(), http.StatusFound)
			return
		}
		// create page if it doesn't exist
		if !fileExists(lw.st.PgDir + s.Page + ".txt") {
			s.Action = ActionEdit
		}
	}

	if s.Action == ActionEdit || s.Preview {
		lw.Edit(s)
	}
	lw.t.Render(w, s, lw.st)
}

func (lw *LionWiki) Run() error {
	if err := lw.createDirectories(lw.st); err != nil {
		return err
	}
	// Load Plugins
	// plugin('pluginsLoaded');
	if err := lw.t.Load(lw.st.Template); err != nil {
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
	showSource := 0
	if s.ShowSource {
		showSource = 1
	}
	s.ConFormBegin = fmt.Sprintf(`<form action="%s" method="post">
	<input type="hidden" name="action" value="save"/>
	<input type="hidden" name="last_changed" value="%s"/>
	<input type="hidden" name="showsource" value="%d"/>
	<input type="hidden" name="par" value="%s"/>
	<input type="hidden" name="page" value="%s"/>`, s.Self, s.LastChangedTs, showSource, h(s.Par), h(s.Page))
	s.ConFormEnd = "</form>"
	s.ConTextarea = fmt.Sprintf(`<textarea class="contentTextarea" name="content" style="width:100%%" cols="100" rows="30">%s</textarea>`, strings.Replace(s.Content, "&lt;", "<", -1))
	s.ConPreview = fmt.Sprintf(`<input class="submit" type="submit" name="preview" value="%s"/>`, s.Tr.Get("T_PREVIEW"))

	if s.ShowSource {
		s.ConSubmit = fmt.Sprintf(`<input class="submit" type="submit" value="%s"/>`, s.Tr.Get("T_DONE"))
		s.EditSummaryText = s.Tr.Get("T_EDIT_SUMMARY")
		s.EditSummary = fmt.Sprintf(`<input type="text" name="esum" value="%s"/>`, h(s.Esum))

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
		s.Title = s.Tr.Get("T_PREVIEW") + s.Page
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

func h(text string) string {
	return html.EscapeString(text)
}
func u(text string) string {
	// TODO: index.php:511
	return text
}
