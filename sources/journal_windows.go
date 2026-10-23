//go:build windows

package main

// Journal coloré (RichEdit du système, Msftedit.dll) : information, remarque,
// avertissement, erreur, réussite. Le texte du journal est aussi gardé dans
// logBuf, à l'identique de l'ancien champ texte (lignes séparées par \r\n) :
// c'est lui qui part dans le rapport d'erreur. Sans RichEdit, le journal
// redevient un champ texte simple, sans couleurs.

import (
	"regexp"
	"strings"
	"sync"
	"unsafe"
)

const (
	emExGetSel       = 0x0400 + 52
	emExSetSel       = 0x0400 + 55
	emSetBkgndColor  = 0x0400 + 67
	emSetCharFormat  = 0x0400 + 68
	emGetTextLenEx   = 0x0400 + 95
	emSetTextEx      = 0x0400 + 97
	emSetRect        = 0x00B3
	emScrollCaret    = 0x00B7
	emSetMargins     = 0x00D3
	scfSelection     = 0x1
	scfAll           = 0x4
	cfmBold          = 0x1
	cfmSize          = 0x80000000
	cfmFace          = 0x20000000
	cfmColor         = 0x40000000
	cfmCharset       = 0x08000000
	cfeBold          = 0x1
	richEditClass    = "RICHEDIT50W"
	logInfo          = 0
	logNote          = 1
	logWarn          = 2
	logError         = 3
	logOK            = 4
	logHead          = 5
	logLink          = 6
)

type charFormat struct {
	size, mask, effects uint32
	height, offset      int32
	color               uint32
	charset, pitch      byte
	face                [32]uint16
}

type charRange struct{ min, max int32 }

// richOK : le contrôle RichEdit est disponible (Msftedit.dll chargée).
func richOK() bool { return msftedit.Load() == nil }

// newLog : crée le journal (RichEdit si possible) ; renvoie vrai s'il est en couleurs.
func (a *app) newLog() bool {
	if richOK() {
		a.child(richEditClass, "", esMultiline|esReadOnly|esAutoVScroll|wsVScroll|esDisableNoScrl, 0, 0, 0, 10, 10, idLog)
		if h := a.ctl[idLog]; h != 0 {
			send(h, emSetBkgndColor, 0, colPanneau)
			darkScrollbars(h)
			return true
		}
	}
	a.child("EDIT", "", esMultiline|esReadOnly|esAutoVScroll|wsVScroll, 0, 0, 0, 10, 10, idLog)
	darkScrollbars(a.ctl[idLog])
	return false
}

// logFont : police du journal ; pour RichEdit, seules la police et la taille changent (les couleurs restent).
func (a *app) logFont(face string, pt int) {
	h := a.ctl[idLog]
	if !a.richLog {
		send(h, wmSetFont, a.fontSmall, 1)
		return
	}
	cf := charFormat{mask: cfmFace | cfmSize | cfmCharset, height: int32(pt * 20), charset: 1}
	cf.size = uint32(unsafe.Sizeof(cf))
	copy(cf.face[:31], toUTF16(face))
	send(h, emSetCharFormat, scfAll, uintptr(unsafe.Pointer(&cf)))
	send(h, emSetCharFormat, 0, uintptr(unsafe.Pointer(&cf))) // format par défaut
	a.richZoom(h)
}

// richZoom : le texte d'un RichEdit (en points) suit l'échelle de la fenêtre quand
// celle-ci est réduite pour tenir dans un petit écran.
func (a *app) richZoom(h uintptr) {
	if a.realDpi > 0 && a.realDpi != a.dpi {
		send(h, 0x0400+225, uintptr(a.dpi), uintptr(a.realDpi)) // EM_SETZOOM
	}
}

func toUTF16(s string) []uint16 {
	u := []uint16{}
	for p := w16(s); *p != 0; p = (*uint16)(unsafe.Add(unsafe.Pointer(p), 2)) {
		u = append(u, *p)
	}
	return u
}

// clearLog : journal vidé (écran et texte du rapport).
func (a *app) clearLog() {
	setText(a.ctl[idLog], "")
	a.logBuf.Reset()
	a.logEntries = a.logEntries[:0]
}

// logText : texte complet du journal, tel que l'ancien champ texte le renvoyait.
func (a *app) logText() string { return a.logBuf.String() }

func (a *app) log(s string) {
	h := a.ctl[idLog]
	prefix := "\r\n"
	if a.logBuf.Len() == 0 {
		prefix = ""
	}
	a.logBuf.WriteString(prefix + s)
	a.logEntries = append(a.logEntries, s)
	if a.richLog {
		a.richInsert(prefix, s)
		send(h, emScrollCaret, 0, 0)
		send(h, 0x0115, 7, 0) // WM_VSCROLL SB_BOTTOM : dernière ligne visible
	} else {
		n, _, _ := pGetWindowTextLengthW.Call(h)
		send(h, emSetSel, n, n)
		send(h, emReplaceSel, 0, wp(prefix+s))
	}
	pUpdateWindow.Call(h)
	if !a.inWelcome {
		a.welcomeOnly = false
	}
}

// richInsert : ajoute une ligne en fin de journal, dans la couleur de son type.
func (a *app) richInsert(prefix, s string) {
	h := a.ctl[idLog]
	var n struct{ flags, codepage uint32 }
	n.flags, n.codepage = 0x8|0x2, 1200 // GTL_NUMCHARS | GTL_PRECISE
	end := int32(send(h, emGetTextLenEx, uintptr(unsafe.Pointer(&n)), 0))
	cr := charRange{end, end}
	send(h, emExSetSel, 0, uintptr(unsafe.Pointer(&cr)))
	col, bold := logStyle(logKindOf(s))
	cf := charFormat{mask: cfmColor | cfmBold, color: uint32(col)}
	if bold {
		cf.effects = cfeBold
	}
	cf.size = uint32(unsafe.Sizeof(cf))
	send(h, emSetCharFormat, scfSelection, uintptr(unsafe.Pointer(&cf)))
	send(h, emReplaceSel, 0, wp(prefix+s))
}

// recolorLog : journal redessiné dans les couleurs du thème courant. Le texte (et celui
// du rapport d'erreur, logBuf) ne change pas ; la position de lecture est conservée.
func (a *app) recolorLog() {
	h := a.ctl[idLog]
	if !a.richLog {
		pInvalidateRect.Call(h, 0, 1)
		return
	}
	si := getScroll(h)
	atEnd := int32(si.page) == 0 || si.pos+int32(si.page) >= si.max
	var pos point
	send(h, emGetScrollPos, 0, uintptr(unsafe.Pointer(&pos)))
	var sel charRange // sélection en cours (copie d'un passage), rendue après la reconstruction
	send(h, emExGetSel, 0, uintptr(unsafe.Pointer(&sel)))
	send(h, 0x000B, 0, 0) // WM_SETREDRAW : pas de scintillement pendant la reconstruction
	send(h, emSetBkgndColor, 0, colPanneau)
	setText(h, "")
	for i, s := range a.logEntries {
		prefix := "\r\n"
		if i == 0 || (i > 0 && a.logBufEmptyBefore(i)) {
			prefix = ""
		}
		a.richInsert(prefix, s)
	}
	send(h, emExSetSel, 0, uintptr(unsafe.Pointer(&sel)))
	send(h, 0x000B, 1, 0)
	if atEnd {
		send(h, 0x0115, 7, 0) // SB_BOTTOM
	} else {
		send(h, emSetScrollPos, 0, uintptr(unsafe.Pointer(&pos)))
	}
	pInvalidateRect.Call(h, 0, 1)
}

// logBufEmptyBefore : le journal était-il encore vide avant l'entrée i (lignes vides en tête) ?
func (a *app) logBufEmptyBefore(i int) bool {
	for _, s := range a.logEntries[:i] {
		if s != "" {
			return false
		}
	}
	return true
}

func logStyle(k int) (uintptr, bool) {
	switch k {
	case logNote:
		return colDiscret, false
	case logWarn:
		return colOrange, false
	case logError:
		return colRouge, true
	case logOK:
		return colVert, true
	case logHead:
		return colAccentClair, true
	case logLink:
		return colAccentClair, false
	}
	return colTexte, false
}

// logKindOf : type d'une ligne, d'après les symboles déjà présents dans les textes
// (✖ ⚠ ✔ ℹ ──) ou le modèle complet de la traduction ; sinon, information.
func logKindOf(s string) int {
	t := strings.TrimLeft(s, " \r\n")
	switch {
	case t == "":
		return logInfo
	case strings.HasPrefix(t, "✖"):
		return logError
	case strings.HasPrefix(t, "⚠"):
		return logWarn
	case strings.HasPrefix(t, "✔"):
		return logOK
	case strings.HasPrefix(t, "ℹ"):
		return logNote
	case strings.HasPrefix(t, "──"):
		return logHead
	case s == T("welcome.0"):
		return logHead
	case strings.HasPrefix(s, T("log.report")):
		return logError
	case strings.HasPrefix(s, T("log.outDir")):
		return logOK
	case matchTemplate("log.warnings", s):
		return logWarn
	case matchTemplate("log.update", s):
		return logLink
	}
	return logInfo
}

var (
	tmplMu    sync.Mutex
	tmplCache = map[string]*regexp.Regexp{}
	verbTmpl  = regexp.MustCompile(`%%|%\[?\d*\]?[-+# 0-9.]*[a-zA-Z]`)
)

// matchTemplate : la ligne correspond-elle au modèle traduit de la clé (paramètres compris) ?
func matchTemplate(key, s string) bool {
	tmpl := T(key)
	tmplMu.Lock()
	re, ok := tmplCache[tmpl]
	if !ok {
		var b strings.Builder
		b.WriteString(`(?s)^`)
		last := 0
		for _, m := range verbTmpl.FindAllStringIndex(tmpl, -1) {
			b.WriteString(regexp.QuoteMeta(tmpl[last:m[0]]))
			switch v := tmpl[m[0]:m[1]]; {
			case v == "%%":
				b.WriteString("%")
			case strings.HasSuffix(v, "d"):
				b.WriteString(`-?\d+`)
			default:
				b.WriteString(`.*?`)
			}
			last = m[1]
		}
		b.WriteString(regexp.QuoteMeta(tmpl[last:]) + `$`)
		re = regexp.MustCompile(b.String())
		tmplCache[tmpl] = re
	}
	tmplMu.Unlock()
	return re.MatchString(s)
}
