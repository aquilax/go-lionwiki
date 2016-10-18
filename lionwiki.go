package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
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
	} else if s.Action == ActionHistory { // show whole history of page
		lw.History(s)
	} else if s.Action == ActionDiff {
		lw.Diff(s)
	} else if s.Action == ActionSearch {
		lw.Search(s)
	} else if s.Action == ActionRecent { // recent changes
		lw.Recent(s)
	} else {
		//plugin('action', $action);
	}

	if s.Action == ActionNone || s.Preview { // page parsing
		lw.View(s)
	}
	// plugin('formatFinished');

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

func (lw *LionWiki) History(s *Session) {
	// for($files = array(), $dir = @opendir("$HIST_DIR$page/"); $f = @readdir($dir);)
	// 	if(substr($f, -4) == '.bak')
	// 		$files[] = $f;

	// rsort($files);
	// $CON = '<form action="'.$self.'" method="get"><input type="hidden" name="action" value="diff"/><input type="hidden" name="page" value="'.h($page).'"/><input type="submit" class="submit" value="'.$T_DIFF.'"/><br/>';
	// $meta = @fopen("$HIST_DIR$page/meta.dat", "rb");

	// for($i = 0, $mi = 1, $c = count($files); $i < $c; $i++) {
	// 	if(($m = meta_getline($meta, $mi)) && !strcmp(basename($files[$i], ".bak"), $m[0]))
	// 		$mi++;

	// 	$CON .= '<input type="radio" name="f1" value="'.h($files[$i]).'"/><input type="radio" name="f2" value="'.h($files[$i]).'"/>';
	// 	$CON .= "<a href=\"$self?page=".u($page)."&amp;action=rev&amp;f1=".$files[$i]."\">".rev_time($files[$i])."</a> - ($m[2] B) $m[1] <i>".h($m[3])."</i><br/>";
	// }

	// $CON .= '</form>';
}

func (lw *LionWiki) Diff(s *Session) {
	// if(!$f1 && $dir = @opendir("$HIST_DIR$page/")) { // diff is made on two last revisions
	// 	while($f = @readdir($dir))
	// 		if(substr($f, -4) == '.bak')
	// 			$files[] = clear_path($f);

	// 	rsort($files);

	// 	die(header("Location:$self?action=diff&page=".u($page)."&f1=".u($files[0])."&f2=".u($files[1])));
	// }

	// $r1 = "<a href=\"$self?page=".u($page)."&amp;action=rev&amp;f1=$f1\">".rev_time($f1)."</a>";
	// $r2 = "<a href=\"$self?page=".u($page)."&amp;action=rev&amp;f1=$f2\">".rev_time($f2)."</a>";

	// $CON = str_replace(array("{REVISION1}", "{REVISION2}"), array($r1, $r2), $T_REV_DIFF);
	// $CON .= diff($f1, $f2);
}

func (lw *LionWiki) Search(s *Session) {
	// for($files = array(), $dir = opendir($PG_DIR); $f = readdir($dir);)
	// 	if(substr($f, -4) == '.txt' && ($c = @file_get_contents($PG_DIR . $f)))
	// 		if(!$query || stristr($f . $c, $query) !== false)
	// 			$files[] = clear_path(substr($f, 0, -4));

	// sort($files);

	// foreach($files as $f)
	// 	$list .= "<li><a href=\"$self?page=".u($f).'&amp;redirect=no">'.h($f)."</a></li>";

	// $CON = "<ul>$list</ul>";

	// if($query && !file_exists("$PG_DIR$query.txt")) // offer to create the page
	// 	$CON = "<p><i><a href=\"$self?action=edit&amp;page=".u($query)."\">$T_CREATE_PAGE ".h($query)."</a>.</i></p>".$CON;

	// $TITLE = (!$query ? $T_LIST_OF_ALL_PAGES : "$T_SEARCH_RESULTS $query") . " (".count($files).")";
}

func (lw *LionWiki) Recent(s *Session) {
	// for($files = array(), $dir = opendir($PG_DIR); $f = readdir($dir);)
	// 	if(substr($f, -4) == '.txt')
	// 		$files[substr($f, 0, -4)] = filemtime($PG_DIR . $f);

	// arsort($files);

	// foreach(array_slice($files, 0, 100) as $f => $ts) { // just first 100 files
	// 	if($meta = @fopen($HIST_DIR . basename($f, '.txt') . '/meta.dat', 'r')) {
	// 		$m = meta_getline($meta, 1);
	// 		fclose($meta);
	// 	}

	// 	$recent .= "<tr><td class=\"rc-diff\"><a href=\"$self?page=".u($f)."&amp;action=diff\">$T_DIFF</a></td><td class=\"rc-date\" nowrap>".date($DATE_FORMAT, $ts + $LOCAL_HOUR * 3600)."</td><td class=\"rc-ip\">$m[1]</td><td class=\"rc-page\"><a href=\"$self?page=".u($f)."&amp;redirect=no\">".h($f)."</a> <span class=\"rc-size\">($m[2] B)</span><i class=\"rc-esum\"> ".h($m[3])."</i></td></tr>";
	// }

	// $CON = "<table>$recent</table>";
	// $TITLE = $T_RECENT_CHANGES;
}

func (lw *LionWiki) View(s *Session) {
	title := regexp.MustCompile(`(?<!\^)\{title:([^}\n]*)\}`).FindStringSubmatch(s.Content)
	if len(title) > 0 {
		s.Title = title[1]
		s.Content = strings.Replace(s.Content, title[0], "", 1)
	}

	// if(preg_match("/(?<!\^)\{title:([^}\n]*)\}/U", $CON, $m)) { // Change page title
	// 	$TITLE = $m[1];
	// 	$CON = str_replace($m[0], "", $CON);
	// }

	// // subpages
	// while(preg_match('/(?<!\^){include:([^}]+)}/Um', $CON, $m)) {
	// 	$includePage = clear_path($m[1]);

	// 	if(!strcmp($includePage, $page)) // limited recursion protection
	// 		$CON = str_replace($m[0], "'''Warning: subpage recursion!'''", $CON);
	// 	elseif(file_exists("$PG_DIR$includePage.txt"))
	// 		$CON = str_replace($m[0], file_get_contents("$PG_DIR$includePage.txt"), $CON);
	// 	else
	// 		$CON = str_replace($m[0], "'''Warning: subpage $includePage was not found!'''", $CON);
	// }

	// plugin('subPagesLoaded');

	// // save content not intended for substitutions ({html} tag)
	// if(!$NO_HTML) { // XSS protection
	// 	preg_match_all("/(?<!\^)\{html\}(.+)\{\/html\}/Ums", $CON, $htmlcodes, PREG_PATTERN_ORDER);
	// 	$CON = preg_replace("/(?<!\^)\{html\}.+\{\/html\}/Ums", "{HTML}", $CON);

	// 	foreach($htmlcodes[1] as &$hc)
	// 		$hc = str_replace("&lt;", "<", $hc);
	// }

	// $CON = preg_replace("/(?<!\^)<!--.*-->/U", "", $CON); // internal comments
	// $CON = preg_replace_callback("/\^(.)/", function ($m) { return '&#' . ord($m[1]) . ';'; }, $CON);
	// $CON = str_replace(array("<", "&"), array("&lt;", "&amp;"), $CON);
	// $CON = preg_replace("/&amp;([a-z]+;|\#[0-9]+;)/U", "&$1", $CON); // keep HTML entities
	// $CON = preg_replace("/(\r\n|\r)/", "\n", $CON); // unifying newlines to Unix ones

	// preg_match_all("/{{(.+)}}/Ums", $CON, $codes, PREG_PATTERN_ORDER);
	// $CON = preg_replace("/{{(.+)}}/Ums", "<pre>{CODE}</pre>", $CON);

	// // spans
	// preg_match_all("/\{([\.#][^\s\"\}]*)(\s([^\}\"]*))?\}/m", $CON, $spans, PREG_SET_ORDER);

	// foreach($spans as $m) {
	// 	$class = $id = '';
	// 	$parts = preg_split('/([\.#])/', $m[1], -1, PREG_SPLIT_DELIM_CAPTURE | PREG_SPLIT_NO_EMPTY);

	// 	for($i = 0, $c = count($parts); $c > 1 && $i < $c; $i += 2)
	// 		if($parts[$i] == '.')
	// 			$class .= $parts[$i + 1] . ' ';
	// 		else
	// 			$id = $parts[$i + 1];

	// 	$CON = str_replace($m[0], '<span'.($id ? " id=\"$id\"" : '').($class ? " class=\"$class\"" : '').($m[3] ? " style=\"$m[3]\"" : '').'>', $CON);
	// }

	// $CON = str_replace('{/}', '</span>', $CON);

	// plugin('formatBegin');

	// $CON = strtr($CON, array('&lt;-->' => '&harr;', '-->' => '&rarr;', '&lt;--' => '&larr;', "(c)" => '&copy;', "(r)" => '&reg;'));
	// $CON = preg_replace("/\{small\}(.*)\{\/small\}/U", "<small>$1</small>", $CON); // small
	// $CON = preg_replace("/\{su([bp])\}(.*)\{\/su([bp])\}/U", "<su$1>$2</su$3>", $CON); // sup and sub

	// $CON = preg_replace("/^([^!\*#\n][^\n]+)$/Um", '<p>$1</p>', $CON); // paragraphs

	// // images
	// preg_match_all("#\[((https?://|\./)[^|\]\"]+\.(jpeg|jpg|gif|png))(\|[^\]]+)?\]#", $CON, $imgs, PREG_SET_ORDER);

	// foreach($imgs as $img) {
	// 	$link = $i_attr = $a_attr = $center = $tag = "";

	// 	preg_match_all("/\|([^\]\|=]+)(=([^\]\|\"]+))?(?=[\]\|])/", $img[0], $options, PREG_SET_ORDER);

	// 	foreach($options as $o)
	// 		if($o[1] == 'center') $center = true;
	// 		elseif($o[1] == 'right' || $o[1] == 'left') $i_attr .= " style=\"float:$o[1]\"";
	// 		elseif($o[1] == 'link') $link = (substr($o[3], 0, 4) == "http" || substr($o[3], 0, 2) == "./") ? $o[3] : "$self?page=" . u($o[3]);
	// 		elseif($o[1] == 'alt') $i_attr .= ' alt="'.h($o[3]).'"';
	// 		elseif($o[1] == 'title') $a_attr .= ' title="'.h($o[3]).'"';

	// 	$tag = "<img src=\"$img[1]\"$i_attr/>";

	// 	if($link) $tag = "<a href=\"$link\"$a_attr>$tag</a>";
	// 	if($center) $tag = "<div style=\"text-align:center\">$tag</div>";

	// 	$CON = str_replace($img[0], $tag, $CON);
	// }

	// $CON = preg_replace('#([0-9a-zA-Z\./~\-_]+@[0-9a-z/~\-_]+\.[0-9a-z\./~\-_]+)#i', '<a href="mailto:$0">$0</a>', $CON); // mail recognition

	// // links
	// $CON = preg_replace("#\[([^\]\|]+)\|(\./([^\]]+)|(https?://[0-9a-zA-Z\.\#/~\-_%=\?\&,\+\:@;!\(\)\*\$']*))\]#U", '<a href="$2" class="external">$1</a>', $CON);
	// $CON = preg_replace("#(?<!\")https?://[0-9a-zA-Z\.\#/~\-_%=\?\&,\+\:@;!\(\)\*\$']*#i", '<a href="$0" class="external">$0</a>', $CON);

	// preg_match_all("/\[(?:([^|\]\"]+)\|)?([^\]\"#]+)(?:#([^\]\"]+))?\]/", $CON, $matches, PREG_SET_ORDER); // matching Wiki links

	// foreach($matches as $m) {
	// 	$m[1] = $m[1] ? $m[1] : $m[2]; // is page label same as its name?
	// 	$m[3] = $m[3] ? '#'.u(preg_replace('/[^\da-z]/i', '_', $m[3])) : ''; // anchor

	// 	$attr = file_exists("$PG_DIR$m[2].txt") ? $m[3] : '&amp;action=edit" class="pending';
	// 	$CON = str_replace($m[0], '<a href="'.$self.'?page='.u($m[2]).$attr.'">'.$m[1].'</a>', $CON);
	// }

	// for($i = 10; $i >= 1; $i--) { // Lists, ordered, unordered
	// 	$CON = preg_replace('/^'.str_repeat('\*', $i)."(.*)(\n?)/m", str_repeat('<ul>', $i).'<li>$1</li>'.str_repeat('</ul>', $i).'$2', $CON);
	// 	$CON = preg_replace('/^'.str_repeat('\#', $i)."(.*)(\n?)/m", str_repeat('<ol>', $i).'<li>$1</li>'.str_repeat('</ol>', $i).'$2', $CON);
	// 	$CON = preg_replace("#(</ol>\n?<ol>|</ul>\n?<ul>)#", '', $CON);
	// }

	// // headings
	// preg_match_all('/^(!+)(.*)$/m', $CON, $matches, PREG_SET_ORDER);
	// $stack = array();

	// for($h_id = max($par, 1), $i = 0, $c = count($matches); $i < $c && $m = $matches[$i]; $i++, $h_id++) {
	// 	$excl = strlen($m[1]) + 1;
	// 	$hash = preg_replace('/[^\da-z]/i', '_', $m[2]);

	// 	for($ret = ''; end($stack) >= $excl; $ret .= '</div>', array_pop($stack));

	// 	$stack[] = $excl;

	// 	$ret .= "<div class=\"par-div\" id=\"par-$h_id\"><h$excl id=\"$hash\">$m[2]";

	// 	if(is_writable($PG_DIR . $page . '.txt'))
	// 		$ret .= "<span class=\"par-edit\">(<a href=\"$self?action=edit&amp;page=".u($page)."&amp;par=$h_id\">$T_EDIT</a>)</span>";

	// 	$CON = preg_replace('/' . preg_quote($m[0], '/') . '/', "$ret</h$excl>", $CON, 1);
	// 	$TOC .= str_repeat("<ul>", $excl - 2).'<li><a href="'.$self.'?page='.u($page).'#'.u($hash).'">'.$m[2].'</a></li>'.str_repeat("</ul>", $excl - 2);
	// }

	// $CON .= str_repeat('</div>', count($stack));

	// $TOC = '<ul id="toc">' . preg_replace(array_fill(0, 5, "#</ul>\n*<ul>#"), array_fill(0, 5, ''), $TOC) . '</ul>';
	// $TOC = str_replace(array('</li><ul>', '</ul><li>', '</ul></ul>', '<ul><ul>'), array('<ul>', '</ul></li><li>', '</ul></li></ul>', '<ul><li><ul>'), $TOC);

	// $CON = preg_replace("/'--(.*)--'/Um", '<del>$1</del>', $CON); // strikethrough
	// $CON = preg_replace("/'__(.*)__'/Um", '<u>$1</u>', $CON); // underlining
	// $CON = preg_replace("/'''(.*)'''/Um", '<strong>$1</strong>', $CON); // bold
	// $CON = preg_replace("/''(.*)''/Um", '<em>$1</em>', $CON); // italic
	// $CON = str_replace('{br}', '<br style="clear:both"/>', $CON); // new line
	// $CON = preg_replace('/-----*/', '<hr/>', $CON); // horizontal line
	// $CON = str_replace('--', '&mdash;', $CON); // --

	// $CON = preg_replace(array_fill(0, count($codes[1]) + 1, '/{CODE}/'), $codes[1], $CON, 1); // put HTML and "normal" codes back
	// $CON = preg_replace(array_fill(0, count($htmlcodes[1]) + 1, '/{HTML}/'), $htmlcodes[1], $CON, 1);

	// plugin('formatEnd');
}
