package main

import "net/http"

type ActionType string

const (
	ActionNone ActionType = ""
	ActionEdit ActionType = "edit"
)

type Session struct {
	t           *Translation
	Action      ActionType
	Title       string
	Page        string
	Content     string
	MoveTo      string
	F1          string
	F2          string
	Lang        string
	Erasecookie bool
	Preview     bool
	ShowSource  bool
	Error       string

	Head            string
	ConFormBegin    string
	ConFormEnd      string
	ConPreview      string
	ConTextarea     string
	ConSubmit       string
	EditSummaryText string
	EditSummary     string
}

func NewSession(r *http.Request) *Session {
	q := r.URL.Query()
	page := clearPath(q.Get("page"))
	s := &Session{
		t:           NewTranslation(),
		Action:      ActionType(q.Get("action")),
		Lang:        clearPath(q.Get("lang")),
		Page:        page,
		MoveTo:      clearPath(q.Get("moveto")),
		F1:          clearPath(q.Get("f1")),
		F2:          clearPath(q.Get("f2")),
		Content:     r.PostForm.Get("content"),
		Erasecookie: len(q.Get("erasecookie")) > 0,
		Error:       q.Get("error"),
		Preview:     len(r.PostForm.Get("preview")) > 0,
		ShowSource:  len(r.PostForm.Get("showsource")) > 0,
		Title:       page,
	}
	return s
}

func clearPath(path string) string {
	// TODO: index.php:516
	return path
}
