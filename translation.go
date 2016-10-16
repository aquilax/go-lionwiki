package main

type Translation map[string]string

func NewTranslation() *Translation {
	t := make(Translation)
	return &t
}

func (t *Translation) Get(text string) string {
	return text
}
