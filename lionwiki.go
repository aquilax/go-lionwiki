package main

import (
	"fmt"
	"html"
	"io/ioutil"
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

	if lw.st.ProtectedRead && !s.Authentified() {
		// does user need password to read content of site. If yes, ask for it.
		s.Content = fmt.Sprintf(`<form action=\"%s?page=%s" method="post"><p>%s <input type="password" name="sc"/> <input class="submit" type="submit"/></p></form>`, s.Self, u(s.Page), s.Tr.Get("T_PROTECTED_READ"))
		s.Action = ActionViewHTML
	} else if s.Restore || s.Action == ActionRev { // Show old revision
		// TODO: s.Content = @file_get_contents("$HIST_DIR$page/$f1");
		if s.Action == ActionRev {
			//revRestore := fmt.Sprintf(`[%s|./%s?page=%s&amp;action=edit&amp;f1=%s&amp;restore=1]`, s.Tr.Get("T_RESTORE"), s.Self, u(s.Page), s.F1)
			// TODO: $CON = strtr($T_REVISION, array('{TIME}' => rev_time($f1), '{RESTORE}' => $rev_restore)) . $CON;
			s.Action = ActionNone
		}
	} else if len(s.Page) > 0 && (s.Action == ActionNone || s.Action == ActionEdit) {
		// TODO: Handle err
		b, _ := ioutil.ReadFile(lw.st.PgDir + s.Page + ".txt")
		s.Content = string(b)
		// TODO: $CON = $par ? get_paragraph($CON, $par) : $CON;

		// if(!$action && substr($CON, 0, 10) == '{redirect:' && $_REQUEST['redirect'] != 'no')
		// 	die(header("Location:$self?page=".u(clear_path(substr($CON, 10, strpos($CON, '}') - 10)))));
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
