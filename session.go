package main

import "net/http"

type Session struct {
	Action      string
	Page        string
	Content     string
	MoveTo      string
	F1          string
	F2          string
	Lang        string
	Erasecookie bool
}

func NewSession(r *http.Request) *Session {
	q := r.URL.Query()
	s := &Session{
		Action:      q.Get("action"),
		Lang:        clearPath(q.Get("lang")),
		Page:        clearPath(q.Get("page")),
		MoveTo:      clearPath(q.Get("moveto")),
		F1:          clearPath(q.Get("f1")),
		F2:          clearPath(q.Get("f2")),
		Content:     r.PostForm.Get("content"),
		Erasecookie: len(q.Get("erasecookie")) > 0,
	}
	return s
}

func (s *Session) GetTitle() string {
	return s.Page
}

func clearPath(path string) string {
	// TODO: index.php:516
	return path
}
