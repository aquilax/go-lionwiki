package main

import (
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

func (t *Template) Render(w http.ResponseWriter, s *Session, settings *Settings) {
	content := t.content
	var val string
	for k, v := range *(NewTVFromSession(s, settings)) {
		re := regexp.MustCompile(`\{(([^}{]*) )?` + k + `( ([^}]*))?\}`)
		val, _ = v.(string)
		repl := ""
		if len(val) > 0 {
			repl = "$2" + strings.Replace(strings.TrimSpace(val), "$", "&#36;", -1) + "$4"
		}
		content = re.ReplaceAll(content, []byte(repl))
	}
	w.Write(content)
}

func NewTVFromSession(s *Session, settings *Settings) *TemplateVars {
	tv := NewTemplateVars()
	if s.Action != ActionNone {
		tv.Set("HEAD", s.Head+`<meta name="robots" content="noindex, nofollow"/>`)
	} else {
		tv.Set("HEAD", s.Head)
	}
	tv.Set("SEARCH_FORM", `<form action="'.$self.'" method="get"><span><input type="hidden" name="action" value="search"/><input type="submit" style="display:none;"/>`)
	tv.Set(`/SEARCH_FORM`, "</span></form>")
	tv.Set("SEARCH_INPUT", `<input type="text" name="query" value="'.h($query).'"/>`)
	tv.Set("SEARCH_SUBMIT", `<input class="submit" type="submit" value="$T_SEARCH"/>`)
	tv.Set("HOME", `<a href=\"$self?page=".u($START_PAGE)."\">$T_HOME</a>`)
	tv.Set("RECENT_CHANGES", `<a href=\"$self?action=recent\">$T_RECENT_CHANGES</a>`)
	tv.Set("ERROR", s.Error)
	tv.Set("HISTORY", "")
	if len(s.Page) > 0 {
		tv.Set("HISTORY", `<a href=\"$self?page=".u($page)."&amp;action=history\">$T_HISTORY</a>`)
	}

	if s.Page == settings.StartPage && s.Page == s.Title {
		tv.Set("PAGE_TITLE", settings.WikiTitle)
	} else {
		tv.Set("PAGE_TITLE", s.Title)
	}
	tv.Set("PAGE_TITLE_HEAD", h(s.Title))
	tv.Set("PAGE_URL", u(s.Page))

	// 'EDIT' => !$action ? ("<a href=\"$self?page=".u($page)."&amp;action=edit".(is_writable("$PG_DIR$page.txt") ? "\">$T_EDIT</a>" : "&amp;showsource=1\">$T_SHOW_SOURCE</a>")) : "",
	//
	tv.Set("WIKI_TITLE", h(settings.WikiTitle))
	// 'LAST_CHANGED_TEXT' => $last_changed_ts ? $T_LAST_CHANGED : "",
	// 'LAST_CHANGED' => $last_changed_ts ? date($DATE_FORMAT, $last_changed_ts + $LOCAL_HOUR * 3600) : "",
	// 'CONTENT' => $action != "edit" ? $CON : "",
	tv.Set("TOC", s.TOC)
	// 'SYNTAX' => $action == "edit" || $preview ? "<a href=\"$SYNTAX_PAGE\">$T_SYNTAX</a>" : "",
	// 'SHOW_PAGE' => $action == "edit" || $preview ? "<a href=\"$self?page=".u($page)."\">$T_SHOW_PAGE</a>" : "",
	// 'COOKIE' => '<a href="'.$self.'?page='.u($page).'&amp;action='.u($action).'&amp;erasecookie=1">'.$T_ERASE_COOKIE.'</a>',
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
