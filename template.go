package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type Template struct {
	content []byte
}

type TemplateVars map[string]interface{}

func NewTemplate() *Template {
	return &Template{}
}

func (t *Template) Load(fileName string) error {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	t.content = b
	return nil
}

func NewTemplateVars() *TemplateVars {
	tv := make(TemplateVars)
	return &tv
}

func (t *Template) Render(w http.ResponseWriter, s *Session, st *Settings) {
	content := string(t.content)

	re := regexp.MustCompile(`\{([^}]* )?plugin:.+( [^}]*)?\}`) // get rid of absent plugin tags
	content = re.ReplaceAllString(content, "")

	var val string
	for k, v := range *(NewTVFromSession(s, st)) {
		re := regexp.MustCompile(`\{(([^}{]*) )?` + k + `( ([^}]*))?\}`)
		val, _ = v.(string)
		repl := ""
		if len(val) > 0 {
			repl = "${2}" + strings.Replace(strings.TrimSpace(val), "$", "&#36;", -1) + "${4}"
		}
		content = re.ReplaceAllString(content, repl)
	}
	w.Write([]byte(content))
}

func NewTVFromSession(s *Session, st *Settings) *TemplateVars {
	tv := NewTemplateVars()
	if s.Action != ActionNone {
		tv.Set("HEAD", s.Head+`<meta name="robots" content="noindex, nofollow"/>`)
	} else {
		tv.Set("HEAD", s.Head)
	}
	tv.Set("SEARCH_FORM", fmt.Sprintf(`<form action="%s" method="get"><span><input type="hidden" name="action" value="search"/><input type="submit" style="display:none;"/>`, s.Self))
	tv.Set(`/SEARCH_FORM`, "</span></form>")
	tv.Set("SEARCH_INPUT", fmt.Sprintf(`<input type="text" name="query" value="%s"/>`, h(s.Query)))
	tv.Set("SEARCH_SUBMIT", fmt.Sprintf(`<input class="submit" type="submit" value="%s"/>`, s.Tr.Get("T_SEARCH")))
	tv.Set("HOME", fmt.Sprintf(`<a href="%s?page=%s">%s</a>`, s.Self, u(st.StartPage), s.Tr.Get("T_HOME")))
	tv.Set("RECENT_CHANGES", fmt.Sprintf(`<a href="%s?action=recent">%s</a>`, s.Self, s.Tr.Get("T_RECENT_CHANGES")))
	tv.Set("ERROR", s.Error)

	tv.Set("HISTORY", "")
	if len(s.Page) > 0 {
		tv.Set("HISTORY", fmt.Sprintf(`<a href="%s?page=%s&amp;action=history">%s</a>`, s.Self, u(s.Page), s.Tr.Get("T_HISTORY")))
	}

	if s.Page == st.StartPage && s.Page == s.Title {
		tv.Set("PAGE_TITLE", st.WikiTitle)
	} else {
		tv.Set("PAGE_TITLE", s.Title)
	}
	tv.Set("PAGE_TITLE_HEAD", h(s.Title))
	tv.Set("PAGE_URL", u(s.Page))

	if s.Action == ActionNone {
		label := s.Tr.Get("T_EDIT")
		extra := ""
		if !s.IsWritable {
			label = s.Tr.Get("T_SHOW_SOURCE")
			extra = "&amp;showsource=1"
		}
		tv.Set("EDIT", fmt.Sprintf(`<a href="%s?page=%s&amp;action=edit%s">%s</a>`, s.Self, u(s.Page), extra, label))
	} else {
		tv.Set("EDIT", "")
	}

	tv.Set("WIKI_TITLE", h(st.WikiTitle))
	if s.LastChangedTs > 0 {
		// TODO: 'LAST_CHANGED' => $last_changed_ts ? date($DATE_FORMAT, $last_changed_ts + $LOCAL_HOUR * 3600) : "",
		tv.Set("LAST_CHANGED_TEXT", string(s.LastChangedTs))
	} else {
		tv.Set("LAST_CHANGED_TEXT", "")
	}
	if s.Action != ActionEdit {
		tv.Set("CONTENT", s.Content)
	} else {
		tv.Set("CONTENT", "")
	}
	tv.Set("TOC", s.TOC)

	if s.Action == ActionEdit || s.Preview {
		tv.Set("SYNTAX", fmt.Sprintf(`<a href="%s">%s</a>`, st.SyntaxPage, s.Tr.Get("T_SYNTAX")))
	} else {
		tv.Set("SYNTAX", "")
	}

	if s.Action == ActionEdit || s.Preview {
		tv.Set("SHOW_PAGE", fmt.Sprintf(`<a href="%s?page=%s">%s</a>`, s.Self, u(s.Page), s.Tr.Get("T_SHOW_PAGE")))
	} else {
		tv.Set("SHOW_PAGE", "")
	}

	tv.Set("COOKIE", fmt.Sprintf(`<a href="%s?page=%s&amp;action=%s&amp;erasecookie=1">%s</a>`, s.Self, u(s.Page), u(string(s.Action)), s.Tr.Get("T_ERASE_COOKIE")))
	tv.Set("CONTENT_FORM", s.ConFormBegin)
	tv.Set("/CONTENT_FORM", s.ConFormEnd)
	tv.Set("CONTENT_TEXTAREA", s.ConTextarea)
	tv.Set("CONTENT_SUBMIT", s.ConSubmit)
	tv.Set("CONTENT_PREVIEW", s.ConPreview)
	tv.Set("RENAME_TEXT", s.RenameText)
	tv.Set("RENAME_INPUT", s.RenameInput)
	tv.Set("EDIT_SUMMARY_TEXT", s.EditSummaryText)
	tv.Set("EDIT_SUMMARY_INPUT", s.EditSummary)
	tv.Set("FORM_PASSWORD", s.FormPassword)
	tv.Set("FORM_PASSWORD_INPUT", s.FormPasswordInput)
	return tv
}

func (tv *TemplateVars) Set(name string, value interface{}) {
	(*tv)[name] = value
}
