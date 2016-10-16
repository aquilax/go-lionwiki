package main

const (
	CookiePrefix = "LW_"
)

type Settings struct {
	WikiTitle string // name of the site
	Password  string // SHA1 hash

	Template      string // presentation template
	ProtectedRead bool   // if true, you need to fill password for reading pages too
	NoHTML        bool   // XSS protection

	StartPage  string // Which page should be default (start page)?
	SyntaxPage string

	DateFormat string
	LocalHour  int

	RealPath       string
	VarDir         string
	PgDir          string
	HistDir        string
	PluginsDir     string
	PluginsDataDir string
	LangDir        string
}

func NewSettings() *Settings {
	varDir := "var/"
	return &Settings{
		WikiTitle:     "My new wiki",
		Password:      "",
		Template:      "templates/dandelion.html",
		ProtectedRead: false,
		NoHTML:        true,
		StartPage:     "Main page",
		SyntaxPage:    "http://lionwiki.0o.cz/?page=Syntax+reference",
		DateFormat:    "Y/m/d H:i",
		LocalHour:     0,

		RealPath:       "/",
		VarDir:         varDir,
		PgDir:          varDir + "pages/",
		HistDir:        varDir + "history/",
		PluginsDir:     "plugins/",
		PluginsDataDir: varDir + "plugins/",
		LangDir:        "lang/",
	}
}
