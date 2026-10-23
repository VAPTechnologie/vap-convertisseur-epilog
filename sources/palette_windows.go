//go:build windows

package main

// Deux palettes : sombre (charte V2 reprise de l'Analyseur Wi-Fi) et claire.
// Les couleurs courantes sont les variables col… utilisées partout dans le dessin :
// usePalette les remplace d'un coup, sans toucher aux contrôles ni à leur contenu.
// Le bandeau du haut reste bleu nuit dans les deux thèmes (logo VAP blanc) : il est
// dessiné avec la palette sombre (withPalette).

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

type palette struct {
	fond, panneau, doux, filet, texte, discret, accent, accentClair, texteAccent  uintptr
	selection, enteteDebut, inactif, survol, bord, bordSurvol, principal          uintptr
	principalOn, anneau, impose, vert, orange, rouge, separateur                  uintptr
	brFond, brPanneau, brDoux                                                     uintptr // pinceaux, créés une fois
}

var (
	darkPal = palette{
		fond: rgb(0x0B, 0x12, 0x20), panneau: rgb(0x11, 0x1A, 0x2C), doux: rgb(0x16, 0x21, 0x3A), filet: rgb(0x22, 0x30, 0x4D),
		texte: rgb(0xE6, 0xED, 0xF7), discret: rgb(0xA9, 0xB8, 0xD0), // 8,6:1 sur un panneau (Analyseur : #8FA1BD, 6,6:1)
		accent: rgb(0x2E, 0x8B, 0xFF), accentClair: rgb(0x5F, 0xB0, 0xFF), texteAccent: rgb(0xFF, 0xFF, 0xFF),
		selection: rgb(0x1B, 0x33, 0x58), enteteDebut: rgb(0x0D, 0x17, 0x30),
		inactif:   rgb(0x7C, 0x8A, 0xA3), // contrôle verrouillé : encore lisible (5:1)
		survol:    rgb(0x2C, 0x3E, 0x63),
		bord:      rgb(0x5B, 0x70, 0x96), // contour des champs, cases et boutons : 3,5:1
		bordSurvol: rgb(0x74, 0x88, 0xAD),
		principal: rgb(0x1F, 0x6F, 0xE0), principalOn: rgb(0x1A, 0x5E, 0xC4), anneau: rgb(0x9C, 0xCB, 0xFF),
		impose: rgb(0x3A, 0x4F, 0x72),
		vert:   rgb(0x3B, 0xC9, 0x84), orange: rgb(0xF3, 0xB0, 0x2B), rouge: rgb(0xFF, 0x5A, 0x5A),
		separateur: rgb(70, 90, 120),
	}
	// palette claire : fond gris-bleu très clair, panneaux blancs, texte bleu nuit, même bleu VAP
	lightPal = palette{
		fond: rgb(0xEE, 0xF2, 0xF8), panneau: rgb(0xFF, 0xFF, 0xFF), doux: rgb(0xE8, 0xEE, 0xF6), filet: rgb(0xD2, 0xDB, 0xE8),
		texte: rgb(0x0E, 0x1A, 0x2E), discret: rgb(0x46, 0x56, 0x6F),
		accent: rgb(0x1F, 0x6F, 0xE0), accentClair: rgb(0x1A, 0x5F, 0xC8), texteAccent: rgb(0xFF, 0xFF, 0xFF),
		selection: rgb(0xDC, 0xE8, 0xFB), enteteDebut: rgb(0x0D, 0x17, 0x30),
		inactif:   rgb(0x5E, 0x6B, 0x82),
		survol:    rgb(0xDC, 0xE5, 0xF2),
		bord:      rgb(0x74, 0x83, 0x9C),
		bordSurvol: rgb(0x4F, 0x60, 0x7E),
		principal: rgb(0x1F, 0x6F, 0xE0), principalOn: rgb(0x1A, 0x5E, 0xC4), anneau: rgb(0x0B, 0x3F, 0x91),
		impose: rgb(0xC7, 0xD2, 0xE3),
		vert:   rgb(0x13, 0x78, 0x4A), orange: rgb(0x9A, 0x58, 0x00), rouge: rgb(0xC4, 0x2B, 0x2B),
		separateur: rgb(70, 90, 120),
	}
	curPal = &darkPal
	isDark = true
)

// couleurs courantes (palette active)
var (
	colFond, colPanneau, colDoux, colFilet, colTexte, colDiscret, colAccent, colAccentClair, colTexteAccent uintptr
	colSelection, colEnteteDebut, colInactif, colSurvol, colBord, colBordSurvol, colPrincipal               uintptr
	colPrincipalOn, colAnneau, colImpose, colVert, colOrange, colRouge, colSeparateur                       uintptr
	brFond, brPanneau, brDoux                                                                               uintptr
)

func init() { usePalette(&darkPal) }

// usePalette : la palette p devient la palette courante.
func usePalette(p *palette) {
	curPal = p
	colFond, colPanneau, colDoux, colFilet, colTexte = p.fond, p.panneau, p.doux, p.filet, p.texte
	colDiscret, colAccent, colAccentClair, colTexteAccent = p.discret, p.accent, p.accentClair, p.texteAccent
	colSelection, colEnteteDebut, colInactif, colSurvol, colBord = p.selection, p.enteteDebut, p.inactif, p.survol, p.bord
	colBordSurvol, colPrincipal, colPrincipalOn, colAnneau, colImpose = p.bordSurvol, p.principal, p.principalOn, p.anneau, p.impose
	colVert, colOrange, colRouge, colSeparateur = p.vert, p.orange, p.rouge, p.separateur
	brFond, brPanneau, brDoux = p.brFond, p.brPanneau, p.brDoux
}

// withPalette : f dessinée avec la palette p (le bandeau reste sombre en mode clair).
func withPalette(p *palette, f func()) {
	saved := curPal
	usePalette(p)
	defer usePalette(saved)
	f()
}

// initBrushes : pinceaux des deux palettes.
func initBrushes() {
	for _, p := range []*palette{&darkPal, &lightPal} {
		p.brFond, _, _ = pCreateSolidBrush.Call(p.fond)
		p.brPanneau, _, _ = pCreateSolidBrush.Call(p.panneau)
		p.brDoux, _, _ = pCreateSolidBrush.Call(p.doux)
	}
	usePalette(curPal)
}

// ---------------------------------------------------------------------------
// Choix mémorisé dans %APPDATA%\VAP Convertisseur\theme.txt (« clair » ou « sombre ») ;
// sans choix enregistré : sombre.

func themeFile() string {
	d, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(d, "VAP Convertisseur", "theme.txt")
}

func loadTheme() bool {
	if p := themeFile(); p != "" {
		if b, err := os.ReadFile(p); err == nil && strings.TrimSpace(string(b)) == "clair" {
			return false
		}
	}
	return true
}

func saveTheme(dark bool) {
	p := themeFile()
	if p == "" {
		return
	}
	v := "sombre"
	if !dark {
		v = "clair"
	}
	// écriture dans un fichier provisoire puis renommage : jamais de fichier à moitié écrit ;
	// en cas d'échec, le choix précédent reste et le programme démarre en sombre au pire
	if os.MkdirAll(filepath.Dir(p), 0o755) != nil {
		return
	}
	tmp := p + ".tmp"
	if os.WriteFile(tmp, []byte(v), 0o644) != nil || os.Rename(tmp, p) != nil {
		os.Remove(tmp)
	}
}

func setPaletteDark(dark bool) {
	isDark = dark
	if dark {
		usePalette(&darkPal)
	} else {
		usePalette(&lightPal)
	}
}

// ---------------------------------------------------------------------------
// Bouton « Mode clair » / « Mode sombre » : le texte annonce le thème qui sera activé.

func (a *app) themeLabel() string {
	if isDark {
		return T("btn.themeLight")
	}
	return T("btn.themeDark")
}

// toggleTheme : bascule immédiate, mémorisée. Champs, options, journal, progression
// et conversion en cours ne sont pas touchés : seules les couleurs changent.
func (a *app) toggleTheme() {
	setPaletteDark(!isDark)
	saveTheme(isDark)
	a.applyTheme()
}

func (a *app) applyTheme() {
	setText(a.ctl[idTheme], a.themeLabel())
	a.layout() // largeur du bouton selon la langue et le libellé
	a.recolorLog()
	scheme := "DarkMode_Explorer"
	if !isDark {
		scheme = "Explorer"
	}
	for _, id := range []int{idLang, idMachine, idSort, idLaser} {
		ci := comboInfo{size: uint32(unsafe.Sizeof(comboInfo{}))}
		if r, _, _ := pGetComboBoxInfo.Call(a.ctl[id], uintptr(unsafe.Pointer(&ci))); r != 0 && ci.list != 0 {
			pSetWindowTheme.Call(ci.list, wp(scheme), 0)
		}
	}
	const rdw = 0x1 | 0x4 | 0x80 | 0x400 // RDW_INVALIDATE | RDW_ERASE | RDW_ALLCHILDREN | RDW_FRAME
	darkTitle(a.hwnd)
	pRedrawWindow.Call(a.hwnd, 0, 0, rdw)
	if helpHwnd != 0 {
		darkTitle(helpHwnd)
		a.renderNotice()
		pRedrawWindow.Call(helpHwnd, 0, 0, rdw)
	}
}

var pRedrawWindow = user32.NewProc("RedrawWindow")

// drawThemeIcon : soleil (vers le mode clair) ou lune (vers le mode sombre), avant le texte.
func (a *app) drawThemeIcon(nm *nmCustomDrawT) {
	hdc, r := nm.hdc, nm.rc
	font, _, _ := pSendMessageW.Call(nm.hdr.hwndFrom, wmGetFont, 0, 0)
	text := getText(nm.hdr.hwndFrom)
	tw := int32(a.measurePx(font, text))
	is := int32(a.s(14))
	gap := int32(a.s(7))
	x := (r.l+r.r)/2 - (is+gap+tw)/2
	cy := (r.t + r.b) / 2
	fg := colTexte
	if nm.itemState&cdisDisabled != 0 {
		fg = colInactif
	}
	// le texte est recentré à droite de l'icône
	inner := rect{r.l + 2, r.t + 2, r.r - 2, r.b - 2}
	bg := colDoux
	switch {
	case nm.itemState&cdisSelected != 0:
		bg = colSelection
	case nm.itemState&cdisHot != 0:
		bg = colSurvol
	}
	fillBox(hdc, inner.l+int32(a.s(6)), inner.t+int32(a.s(3)), inner.r-int32(a.s(6)), inner.b-int32(a.s(3)), bg)
	cx := x + is/2
	if isDark { // soleil
		rad := is / 4
		disc(hdc, cx, cy, rad, fg)
		for i := 0; i < 8; i++ {
			ang := float64(i) * math.Pi / 4
			c, s := math.Cos(ang), math.Sin(ang)
			r1, r2 := float64(rad)+float64(a.s(2)), float64(is)/2
			line(hdc, int32(max(1, a.s(1))), fg, cx+int32(c*r1), cy+int32(s*r1), cx+int32(c*r2), cy+int32(s*r2))
		}
	} else { // lune : disque coupé par un disque de la couleur du bouton
		disc(hdc, cx, cy, is/2, fg)
		disc(hdc, cx+is/3, cy-is/5, is/2-int32(a.s(1)), bg)
	}
	drawText(hdc, text, rect{x + is + gap, r.t, r.r, r.b}, font, fg, dtLeft|dtVCenter|dtSingleLine)
}
