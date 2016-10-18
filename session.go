package main

import "net/http"

type ActionType string

const (
	ActionNone     ActionType = ""
	ActionEdit     ActionType = "edit"
	ActionViewHTML ActionType = "view-html"
	ActionRev      ActionType = "rev"
	ActionHistory  ActionType = "history"
	ActionDiff     ActionType = "diff"
	ActionSearch   ActionType = "search"
	ActionRecent   ActionType = "recent"
)

type Session struct {
	Tr                *Translation
	Self              string
	Action            ActionType
	Title             string
	Page              string
	Content           string
	MoveTo            string
	F1                string
	F2                string
	Lang              string
	Erasecookie       bool
	Preview           bool
	ShowSource        bool
	Error             string
	Query             string
	IsWritable        bool
	LastChangedTs     int
	Par               string
	Esum              string
	Restore           bool
	Head              string
	ConFormBegin      string
	ConFormEnd        string
	ConPreview        string
	ConTextarea       string
	ConSubmit         string
	EditSummaryText   string
	EditSummary       string
	RenameText        string
	RenameInput       string
	FormPassword      string
	FormPasswordInput string
	TOC               string
}

func NewSession(r *http.Request) *Session {
	q := r.URL.Query()
	page := clearPath(q.Get("page"))
	s := &Session{
		Self:        "/",
		Tr:          NewTranslation(),
		Action:      ActionType(q.Get("action")),
		Lang:        clearPath(q.Get("lang")),
		Page:        page,
		MoveTo:      clearPath(q.Get("moveto")),
		F1:          clearPath(q.Get("f1")),
		F2:          clearPath(q.Get("f2")),
		Content:     r.PostForm.Get("content"),
		Restore:     r.PostForm.Get("restore") == "1",
		Erasecookie: len(q.Get("erasecookie")) > 0,
		Error:       q.Get("error"),
		Preview:     len(r.PostForm.Get("preview")) > 0,
		ShowSource:  len(r.PostForm.Get("showsource")) > 0,
		Title:       page,
		IsWritable:  false,
	}
	return s
}

func clearPath(path string) string {
	// TODO: index.php:516
	return path
}

func (s *Session) Authentified() bool {
	// TODO:
	return false
}
