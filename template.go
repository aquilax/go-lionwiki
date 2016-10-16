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

func (t *Template) Render(w http.ResponseWriter, s *Session) {
	content := t.content
	var val string
	for k, v := range *(NewTVFromSession(s)) {
		re := regexp.MustCompile(`\{(([^}{]*) )?` + k + `( ([^}]*))?\}`)
		val, _ = v.(string)
		repl := ""
		if len(val) > 0 {
			repl = "$2" + strings.Replace(strings.TrimSpace(val), "$", "&#36;", -1) + "$4"
		}
		content = re.ReplaceAll(content, []byte(repl))
		// content = strings.Replace(content, k, val, -1)
	}
	w.Write(content)
}

func NewTVFromSession(s *Session) *TemplateVars {
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

	// 'PAGE_TITLE' => h($page == $START_PAGE && $page == $TITLE ? $WIKI_TITLE : $TITLE),
	// 'PAGE_TITLE_HEAD' => h($TITLE),
	// 'PAGE_URL' => u($page),
	// 'EDIT' => !$action ? ("<a href=\"$self?page=".u($page)."&amp;action=edit".(is_writable("$PG_DIR$page.txt") ? "\">$T_EDIT</a>" : "&amp;showsource=1\">$T_SHOW_SOURCE</a>")) : "",
	// 'WIKI_TITLE' => h($WIKI_TITLE),
	// 'LAST_CHANGED_TEXT' => $last_changed_ts ? $T_LAST_CHANGED : "",
	// 'LAST_CHANGED' => $last_changed_ts ? date($DATE_FORMAT, $last_changed_ts + $LOCAL_HOUR * 3600) : "",
	// 'CONTENT' => $action != "edit" ? $CON : "",
	// 'TOC' => $TOC,
	// 'SYNTAX' => $action == "edit" || $preview ? "<a href=\"$SYNTAX_PAGE\">$T_SYNTAX</a>" : "",
	// 'SHOW_PAGE' => $action == "edit" || $preview ? "<a href=\"$self?page=".u($page)."\">$T_SHOW_PAGE</a>" : "",
	// 'COOKIE' => '<a href="'.$self.'?page='.u($page).'&amp;action='.u($action).'&amp;erasecookie=1">'.$T_ERASE_COOKIE.'</a>',
	// 'CONTENT_FORM' => $CON_FORM_BEGIN,
	// '\/CONTENT_FORM' => $CON_FORM_END,
	// 'CONTENT_TEXTAREA' => $CON_TEXTAREA,
	// 'CONTENT_SUBMIT' => $CON_SUBMIT,
	// 'CONTENT_PREVIEW' => $CON_PREVIEW,
	// 'RENAME_TEXT' => $RENAME_TEXT,
	// 'RENAME_INPUT' => $RENAME_INPUT,
	// 'EDIT_SUMMARY_TEXT' => $EDIT_SUMMARY_TEXT,
	// 'EDIT_SUMMARY_INPUT' => $EDIT_SUMMARY,
	// 'FORM_PASSWORD' => $FORM_PASSWORD,
	// 'FORM_PASSWORD_INPUT' => $FORM_PASSWORD_INPUT
	return tv

}

func (tv *TemplateVars) Set(name string, value interface{}) {
	(*tv)[name] = value
}
