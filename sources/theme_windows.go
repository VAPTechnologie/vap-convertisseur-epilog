//go:build windows

package main

// Charte sombre VAP, reprise de l'Analyseur Wi-Fi (analyseur/ui/style.py) :
// mêmes couleurs, mêmes rayons, mêmes états (survol, appui, focus, désactivé).
// Les contrôles Windows gardent leur fonctionnement (clavier, coche, sélection) ;
// seul leur dessin est remplacé : boutons et cases par NM_CUSTOMDRAW, listes
// déroulantes par sous-classement, champs par un cadre dessiné par la fenêtre.

import (
	"syscall"
	"unsafe"
)

// couleurs de l'Analyseur Wi-Fi (FOND, PANNEAU, DOUX, FILET, TEXTE, DISCRET, ACCENT…)
var (
	colFond        = rgb(0x0B, 0x12, 0x20)
	colPanneau     = rgb(0x11, 0x1A, 0x2C)
	colDoux        = rgb(0x16, 0x21, 0x3A)
	colFilet       = rgb(0x22, 0x30, 0x4D)
	colTexte       = rgb(0xE6, 0xED, 0xF7)
	colDiscret     = rgb(0xA9, 0xB8, 0xD0) // textes secondaires : 8,6:1 sur un panneau (Analyseur : #8FA1BD, 6,6:1)
	colAccent      = rgb(0x2E, 0x8B, 0xFF)
	colAccentClair = rgb(0x5F, 0xB0, 0xFF)
	colTexteAccent = rgb(0xFF, 0xFF, 0xFF)
	colSelection   = rgb(0x1B, 0x33, 0x58)
	colEnteteDebut = rgb(0x0D, 0x17, 0x30)
	colInactif     = rgb(0x7C, 0x8A, 0xA3) // contrôle verrouillé : encore lisible (5:1)
	colSurvol      = rgb(0x2C, 0x3E, 0x63)
	colBord        = rgb(0x5B, 0x70, 0x96) // contour des champs, cases et boutons : 3,5:1
	colBordSurvol  = rgb(0x74, 0x88, 0xAD) // survol : 4,9:1
	colPrincipal   = rgb(0x1F, 0x6F, 0xE0) // fond de « Convertir » : texte blanc à 4,8:1
	colPrincipalOn = rgb(0x1A, 0x5E, 0xC4) // « Convertir » enfoncé
	colAnneau      = rgb(0x9C, 0xCB, 0xFF) // bord de « Convertir » survolé
	colImpose      = rgb(0x3A, 0x4F, 0x72) // case cochée mais verrouillée
	colVert        = rgb(0x3B, 0xC9, 0x84)
	colOrange      = rgb(0xF3, 0xB0, 0x2B)
	colRouge       = rgb(0xFF, 0x5A, 0x5A)
	colSeparateur  = rgb(70, 90, 120) // trait après le logo
)

var (
	uxtheme  = syscall.NewLazyDLL("uxtheme.dll")
	dwmapi   = syscall.NewLazyDLL("dwmapi.dll")
	msftedit = syscall.NewLazyDLL("Msftedit.dll")

	pSetWindowTheme        = uxtheme.NewProc("SetWindowTheme")
	pDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
	pSetWindowSubclass     = comctl32.NewProc("SetWindowSubclass")
	pDefSubclassProc       = comctl32.NewProc("DefSubclassProc")
	pRemoveWindowSubclass  = comctl32.NewProc("RemoveWindowSubclass")
	pRoundRect             = gdi32.NewProc("RoundRect")
	pPolyline              = gdi32.NewProc("Polyline")
	pGetTextMetricsW       = gdi32.NewProc("GetTextMetricsW")
	pGradientFill          = msimg32.NewProc("GradientFill")
	pTrackMouseEvent       = user32.NewProc("TrackMouseEvent")
	pGetComboBoxInfo       = user32.NewProc("GetComboBoxInfo")
	pGetFocus              = user32.NewProc("GetFocus")
	pGetWindowRect         = user32.NewProc("GetWindowRect")
	pScreenToClient        = user32.NewProc("ScreenToClient")
	pIsWindowEnabled       = user32.NewProc("IsWindowEnabled")
	pSystemParametersInfoW = user32.NewProc("SystemParametersInfoW")
	pMessageBeep           = user32.NewProc("MessageBeep")
	pGetParent             = user32.NewProc("GetParent")
)

const (
	wmNotify       = 0x004E
	wmEraseBkgnd   = 0x0014
	wmSetFocus     = 0x0007
	wmKillFocus    = 0x0008
	wmMouseMove    = 0x0200
	wmLButtonDown  = 0x0201
	wmMouseLeave   = 0x02A3
	wmNcDestroy    = 0x0082
	wmGetFont      = 0x0031
	wmCtlColorLB   = 0x0134
	wmCtlColorDlg  = 0x0136
	nmCustomDraw   = 0xFFFFFFF4 // NM_CUSTOMDRAW (-12)
	cddsPrepaint   = 0x1
	cdrfDoDefault  = 0x0
	cdrfSkipDef    = 0x4
	cdisSelected   = 0x0001
	cdisDisabled   = 0x0004
	cdisFocus      = 0x0010
	cdisHot        = 0x0040
	cdisShowKbCues = 0x0200
	enSetFocus     = 0x0100
	enKillFocus    = 0x0200
	cbnDropDown    = 7
	cbnCloseUp     = 8
	cbGetLBText    = 0x0148
	cbGetLBTextLen = 0x0149
	cbGetDropped   = 0x0157
	cbSetItemH     = 0x0153
	dtCenter       = 0x1
	dtWordBreak    = 0x10
	dtCalcRect     = 0x400
	dtEditControl  = 0x2000
	odsDisabled    = 0x0004
)

type nmhdr struct {
	hwndFrom uintptr
	idFrom   uintptr
	code     uint32
}

type nmCustomDrawT struct {
	hdr        nmhdr
	drawStage  uint32
	hdc        uintptr
	rc         rect
	itemSpec   uintptr
	itemState  uint32
	itemlParam uintptr
}

type trackMouse struct {
	size, flags uint32
	hwnd        uintptr
	hoverTime   uint32
}

type comboInfo struct {
	size              uint32
	rcItem, rcButton  rect
	stateButton       uint32
	combo, item, list uintptr
}

type triVertex struct {
	x, y                    int32
	red, green, blue, alpha uint16
}

// testWorkArea, testDpi : écran simulé pour les captures de contrôle (build « capture » seulement).
var (
	testWorkArea rect
	testDpi      int
)

// pinceaux de la charte, créés une fois
var brFond, brPanneau, brDoux uintptr

func initTheme() {
	brFond, _, _ = pCreateSolidBrush.Call(colFond)
	brPanneau, _, _ = pCreateSolidBrush.Call(colPanneau)
	brDoux, _, _ = pCreateSolidBrush.Call(colDoux)
	comboCB = syscall.NewCallback(comboProc)
	editCB = syscall.NewCallback(editProc)
}

// darkTitle : barre de titre sombre (Windows 10 2004+ / 11) ; sans effet ailleurs.
func darkTitle(hwnd uintptr) {
	if pDwmSetWindowAttribute.Find() != nil {
		return
	}
	on := int32(1)
	if r, _, _ := pDwmSetWindowAttribute.Call(hwnd, 20, uintptr(unsafe.Pointer(&on)), 4); int32(r) < 0 {
		pDwmSetWindowAttribute.Call(hwnd, 19, uintptr(unsafe.Pointer(&on)), 4) // anciennes versions de Windows 10
	}
	c := uint32(colFond)
	pDwmSetWindowAttribute.Call(hwnd, 35, uintptr(unsafe.Pointer(&c)), 4) // couleur de la barre (Windows 11)
	t := uint32(colTexte)
	pDwmSetWindowAttribute.Call(hwnd, 36, uintptr(unsafe.Pointer(&t)), 4)
}

// darkScrollbars : barres de défilement sombres quand Windows les propose ; sinon, barres standard.
func darkScrollbars(hwnd uintptr) {
	if pSetWindowTheme.Find() == nil {
		pSetWindowTheme.Call(hwnd, wp("DarkMode_Explorer"), 0)
	}
}

// workArea : zone de l'écran hors barre des tâches, en pixels.
func workArea() rect {
	if testWorkArea.r > 0 { // captures de contrôle (build « capture ») : écran simulé
		return testWorkArea
	}
	var r rect
	if ok, _, _ := pSystemParametersInfoW.Call(0x30, 0, uintptr(unsafe.Pointer(&r)), 0); ok == 0 { // SPI_GETWORKAREA
		w, _, _ := pGetSystemMetrics.Call(0)
		h, _, _ := pGetSystemMetrics.Call(1)
		r = rect{0, 0, int32(w), int32(h)}
	}
	return r
}

// ---------------------------------------------------------------------------
// Primitives de dessin

func roundBox(hdc uintptr, r rect, rad int32, fill, border uintptr) {
	br, _, _ := pCreateSolidBrush.Call(fill)
	pen, _, _ := pCreatePen.Call(0, 1, border)
	ob, _, _ := pSelectObject.Call(hdc, br)
	op, _, _ := pSelectObject.Call(hdc, pen)
	pRoundRect.Call(hdc, uintptr(r.l), uintptr(r.t), uintptr(r.r), uintptr(r.b), uintptr(2*rad), uintptr(2*rad))
	pSelectObject.Call(hdc, ob)
	pSelectObject.Call(hdc, op)
	pDeleteObject.Call(br)
	pDeleteObject.Call(pen)
}

// roundFrame : contour arrondi seul (focus).
func roundFrame(hdc uintptr, r rect, rad int32, col uintptr) {
	pen, _, _ := pCreatePen.Call(0, 1, col)
	nb, _, _ := pGetStockObject.Call(5) // NULL_BRUSH
	ob, _, _ := pSelectObject.Call(hdc, nb)
	op, _, _ := pSelectObject.Call(hdc, pen)
	pRoundRect.Call(hdc, uintptr(r.l), uintptr(r.t), uintptr(r.r), uintptr(r.b), uintptr(2*rad), uintptr(2*rad))
	pSelectObject.Call(hdc, ob)
	pSelectObject.Call(hdc, op)
	pDeleteObject.Call(pen)
}

// hGradient : dégradé horizontal de c1 (à gauche) à c2 (à droite).
func hGradient(hdc uintptr, r rect, c1, c2 uintptr) {
	v := func(x, y int32, c uintptr) triVertex {
		return triVertex{x, y, uint16(c&0xFF) << 8, uint16(c>>8&0xFF) << 8, uint16(c>>16&0xFF) << 8, 0}
	}
	vs := [2]triVertex{v(r.l, r.t, c1), v(r.r, r.b, c2)}
	gr := [2]uint32{0, 1}
	if ok, _, _ := pGradientFill.Call(hdc, uintptr(unsafe.Pointer(&vs[0])), 2, uintptr(unsafe.Pointer(&gr[0])), 1, 0); ok == 0 {
		fillBox(hdc, r.l, r.t, r.r, r.b, c2)
	}
}

func drawText(hdc uintptr, s string, r rect, font, col uintptr, flags uintptr) {
	of, _, _ := pSelectObject.Call(hdc, font)
	pSetBkMode.Call(hdc, 1)
	pSetTextColor.Call(hdc, col)
	pDrawTextW.Call(hdc, wp(s), ^uintptr(0), uintptr(unsafe.Pointer(&r)), flags|dtNoPrefix)
	pSelectObject.Call(hdc, of)
}

// textHeight : hauteur d'une ligne de la police, en pixels.
func textHeight(hwnd, font uintptr) int32 {
	dc, _, _ := pGetDC.Call(hwnd)
	of, _, _ := pSelectObject.Call(dc, font)
	var tm [64]byte // TEXTMETRICW : tmHeight en tête
	pGetTextMetricsW.Call(dc, uintptr(unsafe.Pointer(&tm[0])))
	pSelectObject.Call(dc, of)
	pReleaseDC.Call(hwnd, dc)
	return *(*int32)(unsafe.Pointer(&tm[0]))
}

// childRect : rectangle d'un contrôle dans la zone cliente de sa fenêtre.
func childRect(parent, h uintptr) rect {
	var r rect
	pGetWindowRect.Call(h, uintptr(unsafe.Pointer(&r)))
	pScreenToClient.Call(parent, uintptr(unsafe.Pointer(&r)))
	pScreenToClient.Call(parent, uintptr(unsafe.Pointer(&r.r)))
	return r
}

func isEnabled(h uintptr) bool { r, _, _ := pIsWindowEnabled.Call(h); return r != 0 }

// ---------------------------------------------------------------------------
// Boutons et cases à cocher (NM_CUSTOMDRAW)

const (
	btnNormal  = iota // bouton secondaire
	btnPrimary        // action principale bleue (« Convertir », comme « Scanner »)
	btnLink           // lien discret (don, suggestion), comme QPushButton#lien
)

// drawButton : bouton arrondi. États : normal, survol, enfoncé, focus (clavier), désactivé.
func (a *app) drawButton(nm *nmCustomDrawT, kind int, back uintptr) {
	hdc, r, st := nm.hdc, nm.rc, nm.itemState
	fillBox(hdc, r.l, r.t, r.r, r.b, back)
	off, hot, down := st&cdisDisabled != 0, st&cdisHot != 0, st&cdisSelected != 0
	font, _, _ := pSendMessageW.Call(nm.hdr.hwndFrom, wmGetFont, 0, 0)
	text := getText(nm.hdr.hwndFrom)
	rad := int32(a.s(7))
	in := int32(max(2, a.s(2)))
	focus := st&cdisFocus != 0 && !off && st&cdisShowKbCues != 0
	if kind == btnLink {
		fg := colAccentClair
		if hot || down {
			fg = colTexte
		}
		drawText(hdc, text, r, font, fg, dtCenter|dtVCenter|dtSingleLine|dtEndEllip)
		if hot || down { // souligné au survol
			w := int32(a.measurePx(font, text))
			cx, y := (r.l+r.r)/2, (r.t+r.b)/2+textHeight(a.hwnd, font)/2
			fillBox(hdc, cx-w/2, y, cx+w/2, y+int32(max(1, a.s(1))), fg)
		}
		if focus {
			roundFrame(hdc, rect{r.l + 1, r.t + in, r.r - 1, r.b - in}, rad, colAccentClair)
		}
		return
	}
	bg, bd, fg := colDoux, colBord, colTexte
	switch {
	case off:
		bg, bd, fg = colFond, colFilet, colInactif
	case kind == btnPrimary && down:
		bg, bd, fg = colPrincipalOn, colAnneau, colTexteAccent
	case kind == btnPrimary && hot:
		bg, bd, fg = colPrincipal, colAnneau, colTexteAccent
	case kind == btnPrimary:
		bg, bd, fg = colPrincipal, colPrincipal, colTexteAccent
	case down:
		bg, bd = colSelection, colAccent
	case hot:
		bg, bd = colSurvol, colBordSurvol
	}
	roundBox(hdc, r, rad, bg, bd)
	if kind == btnPrimary && hot && !off { // survol de l'action principale : bord clair de 2 px
		roundFrame(hdc, rect{r.l + 1, r.t + 1, r.r - 1, r.b - 1}, rad, colAnneau)
	}
	if focus {
		ring := colAccentClair
		if kind == btnPrimary {
			ring = colTexteAccent
		}
		roundFrame(hdc, rect{r.l + in, r.t + in, r.r - in, r.b - in}, rad-in/2, ring)
	}
	drawText(hdc, text, r, font, fg, dtCenter|dtVCenter|dtSingleLine|dtEndEllip)
}

// paintField : cadre d'un champ de saisie ou d'une liste (contour 3:1 au moins, bleu et
// épais avec le focus, fond plus sombre quand le contrôle est verrouillé).
func (a *app) paintField(hdc uintptr, f rect, focus, on bool) {
	rad := int32(a.s(8))
	switch {
	case !on:
		roundBox(hdc, f, rad, colFond, colFilet)
	case focus:
		roundBox(hdc, f, rad, colPanneau, colAccent)
		roundFrame(hdc, rect{f.l + 1, f.t + 1, f.r - 1, f.b - 1}, rad-1, colAccent)
	default:
		roundBox(hdc, f, rad, colPanneau, colBord)
	}
}

// drawCheck : case à cocher de 15 px (comme QCheckBox::indicator), texte sur une ou deux lignes.
// Une case cochée mais verrouillée (assistance d'air imposée) reste lisible ; avec required,
// le badge « Obligatoire » explique pourquoi elle ne peut pas être décochée.
func (a *app) drawCheck(nm *nmCustomDrawT, back uintptr, required bool) {
	hdc, r, st := nm.hdc, nm.rc, nm.itemState
	fillBox(hdc, r.l, r.t, r.r, r.b, back)
	on := send(nm.hdr.hwndFrom, bmGetCheck, 0, 0) == 1
	off := st&cdisDisabled != 0
	d := int32(a.s(15))
	x, y := r.l+1, r.t+(r.b-r.t-d)/2
	box := rect{x, y, x + d, y + d}
	bg, bd := colPanneau, colBordSurvol
	switch {
	case on && off:
		bg, bd = colImpose, colBord
	case on:
		bg, bd = colAccent, colAccent
	case off:
		bg, bd = colFond, colBord
	case st&cdisHot != 0:
		bd = colAccentClair
	}
	roundBox(hdc, box, int32(a.s(4)), bg, bd)
	if on {
		mk := colTexteAccent
		if off {
			mk = colDiscret
		}
		u := float64(d) / 15
		pts := [3]point{
			{x + int32(3.5*u), y + int32(7.5*u)},
			{x + int32(6.3*u), y + int32(10.3*u)},
			{x + int32(11.8*u), y + int32(4.6*u)},
		}
		pen, _, _ := pCreatePen.Call(0, uintptr(max(2, a.s(2))), mk)
		op, _, _ := pSelectObject.Call(hdc, pen)
		pPolyline.Call(hdc, uintptr(unsafe.Pointer(&pts[0])), 3)
		pSelectObject.Call(hdc, op)
		pDeleteObject.Call(pen)
	}
	fg := colTexte
	if off {
		fg = colDiscret
	}
	font, _, _ := pSendMessageW.Call(nm.hdr.hwndFrom, wmGetFont, 0, 0)
	text := getText(nm.hdr.hwndFrom)
	tr := rect{x + d + int32(a.s(8)), r.t, r.r, r.b}
	// badge « Obligatoire » : juste après le texte s'il tient sur la ligne, sinon au bord droit
	var badge rect
	bt := T("lbl.required")
	if required {
		bw := int32(a.measurePx(a.fontSmall, bt)) + int32(a.s(16))
		bh := textHeight(a.hwnd, a.fontSmall) + int32(a.s(4))
		bx := tr.l + int32(a.measurePx(font, text)) + int32(a.s(10))
		if bx+bw > r.r {
			bx = r.r - bw
			tr.r = bx - int32(a.s(8))
		}
		badge = rect{bx, (r.t + r.b - bh) / 2, bx + bw, (r.t+r.b-bh)/2 + bh}
	}
	// hauteur réelle du texte (une ou deux lignes), centrée verticalement
	calc := tr
	of, _, _ := pSelectObject.Call(hdc, font)
	pDrawTextW.Call(hdc, wp(text), ^uintptr(0), uintptr(unsafe.Pointer(&calc)), dtWordBreak|dtCalcRect|dtNoPrefix)
	pSelectObject.Call(hdc, of)
	th := calc.b - calc.t
	tr.t = r.t + max(0, (r.b-r.t-th)/2)
	tr.b = tr.t + th
	drawText(hdc, text, tr, font, fg, dtWordBreak)
	if required {
		roundBox(hdc, badge, (badge.b-badge.t)/2, colDoux, colBord)
		drawText(hdc, bt, badge, a.fontSmall, colTexte, dtCenter|dtVCenter|dtSingleLine)
	}
	if st&cdisFocus != 0 && st&cdisShowKbCues != 0 {
		in := int32(max(2, a.s(2)))
		roundFrame(hdc, rect{box.l - in, box.t - in, box.r + in, box.b + in}, int32(a.s(5)), colAccentClair)
	}
}

// ---------------------------------------------------------------------------
// Listes déroulantes : partie fermée dessinée par le programme (cadre arrondi,
// texte, chevron) ; ouverture, clavier et liste restent ceux de Windows.

var (
	comboCB    uintptr // syscall.NewCallback(comboProc), créé par initTheme
	comboHover = map[uintptr]bool{}
	comboBack  = map[uintptr]uintptr{} // couleur derrière les coins arrondis
)

func (a *app) themeCombo(h uintptr, back uintptr) {
	comboBack[h] = back
	pSetWindowTheme.Call(h, wp(""), wp("")) // pas d'animation de thème par-dessus notre dessin
	pSetWindowSubclass.Call(h, comboCB, 1, 0)
	ci := comboInfo{size: uint32(unsafe.Sizeof(comboInfo{}))}
	if r, _, _ := pGetComboBoxInfo.Call(h, uintptr(unsafe.Pointer(&ci))); r != 0 && ci.list != 0 {
		darkScrollbars(ci.list)
	}
}

func comboProc(hwnd, msg, wParam, lParam, id, ref uintptr) uintptr {
	switch uint32(msg) {
	case wmPaint:
		var ps paintStruct
		hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		A.paintCombo(hwnd, hdc)
		pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0
	case wmEraseBkgnd:
		return 1
	case wmMouseMove:
		if !comboHover[hwnd] {
			comboHover[hwnd] = true
			tm := trackMouse{size: uint32(unsafe.Sizeof(trackMouse{})), flags: 0x2, hwnd: hwnd} // TME_LEAVE
			pTrackMouseEvent.Call(uintptr(unsafe.Pointer(&tm)))
			pInvalidateRect.Call(hwnd, 0, 0)
		}
	case wmMouseLeave:
		comboHover[hwnd] = false
		pInvalidateRect.Call(hwnd, 0, 0)
	case wmSetFocus, wmKillFocus:
		r, _, _ := pDefSubclassProc.Call(hwnd, msg, wParam, lParam)
		pInvalidateRect.Call(hwnd, 0, 0)
		return r
	case wmNcDestroy:
		pRemoveWindowSubclass.Call(hwnd, comboCB, id)
		delete(comboHover, hwnd)
		delete(comboBack, hwnd)
	}
	r, _, _ := pDefSubclassProc.Call(hwnd, msg, wParam, lParam)
	return r
}

func (a *app) paintCombo(h, hdc uintptr) {
	var r rect
	pGetClientRect.Call(h, uintptr(unsafe.Pointer(&r)))
	fillBox(hdc, r.l, r.t, r.r, r.b, comboBack[h])
	on := isEnabled(h)
	focus, _, _ := pGetFocus.Call()
	dropped := send(h, cbGetDropped, 0, 0) != 0
	switch {
	case on && (focus == h || dropped):
		a.paintField(hdc, r, true, true)
	case on && comboHover[h]:
		roundBox(hdc, r, int32(a.s(8)), colPanneau, colBordSurvol)
	default:
		a.paintField(hdc, r, false, on)
	}
	fg := colTexte
	if !on {
		fg = colInactif
	}
	pad := int32(a.s(10))
	aw := int32(a.s(10))
	tr := rect{r.l + pad, r.t, r.r - pad - aw - int32(a.s(6)), r.b}
	i := int32(send(h, cbGetCurSel, 0, 0))
	if h == a.ctl[idLang] && i >= 0 && i < int32(nLang) {
		fh := int32(a.s(13))
		fw := fh * 3 / 2
		drawFlag(hdc, Lang(i), tr.l, r.t+(r.b-r.t-fh)/2, fw, fh)
		tr.l += fw + int32(a.s(8))
	}
	if i >= 0 {
		n := send(h, cbGetLBTextLen, uintptr(i), 0)
		if int32(n) > 0 {
			buf := make([]uint16, n+1)
			send(h, cbGetLBText, uintptr(i), uintptr(unsafe.Pointer(&buf[0])))
			font, _, _ := pSendMessageW.Call(h, wmGetFont, 0, 0)
			drawText(hdc, syscall.UTF16ToString(buf), tr, font, fg, dtLeft|dtVCenter|dtSingleLine|dtEndEllip)
		}
	}
	// chevron
	cx := r.r - pad - aw/2
	cy := (r.t + r.b) / 2
	k := aw / 2
	pts := [3]point{{cx - k, cy - k/2}, {cx, cy + k/2}, {cx + k, cy - k/2}}
	ac := colDiscret
	if !on {
		ac = colInactif
	}
	pen, _, _ := pCreatePen.Call(0, uintptr(max(1, a.s(2)-1)), ac)
	op, _, _ := pSelectObject.Call(hdc, pen)
	pPolyline.Call(hdc, uintptr(unsafe.Pointer(&pts[0])), 3)
	pSelectObject.Call(hdc, op)
	pDeleteObject.Call(pen)
}

// drawComboItem : une ligne de liste ouverte (langues : avec le drapeau).
func (a *app) drawComboItem(di *drawItem) {
	if int32(di.itemID) < 0 {
		return
	}
	hdc, r := di.hDC, di.rcItem
	bg, fg := colPanneau, colTexte
	if di.itemState&odsSelected != 0 {
		bg = colSelection
	}
	fillBox(hdc, r.l, r.t, r.r, r.b, bg)
	tr := rect{r.l + int32(a.s(8)), r.t, r.r - 2, r.b}
	var s string
	if di.ctlID == idLang {
		if di.itemID >= uint32(nLang) {
			return
		}
		fh := int32(a.s(13))
		fw := fh * 3 / 2
		drawFlag(hdc, Lang(di.itemID), tr.l, r.t+(r.b-r.t-fh)/2, fw, fh)
		tr.l += fw + int32(a.s(8))
		s = langNames[di.itemID]
	} else {
		n := send(di.hwndItem, cbGetLBTextLen, uintptr(di.itemID), 0)
		if int32(n) <= 0 {
			return
		}
		buf := make([]uint16, n+1)
		send(di.hwndItem, cbGetLBText, uintptr(di.itemID), uintptr(unsafe.Pointer(&buf[0])))
		s = syscall.UTF16ToString(buf)
	}
	font, _, _ := pSendMessageW.Call(di.hwndItem, wmGetFont, 0, 0)
	drawText(hdc, s, tr, font, fg, dtLeft|dtVCenter|dtSingleLine|dtEndEllip)
}

// ---------------------------------------------------------------------------
// Champs verrouillés : Windows impose le gris système au texte d'un champ désactivé
// (3,6:1 sur le fond sombre). Le champ reste désactivé ; seul son dessin est remplacé,
// valeur en entier ou abrégée au milieu pour un chemin trop long.

var editCB uintptr // syscall.NewCallback(editProc), créé par initTheme

func (a *app) themeEdit(h uintptr) { pSetWindowSubclass.Call(h, editCB, 3, 0) }

func editProc(hwnd, msg, wParam, lParam, id, ref uintptr) uintptr {
	switch uint32(msg) {
	case wmPaint:
		if !isEnabled(hwnd) {
			var ps paintStruct
			hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
			var r rect
			pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
			fillBox(hdc, r.l, r.t, r.r, r.b, colFond)
			font, _, _ := pSendMessageW.Call(hwnd, wmGetFont, 0, 0)
			r.l++
			drawText(hdc, getText(hwnd), r, font, colInactif, dtLeft|dtVCenter|dtSingleLine|0x4000) // DT_PATH_ELLIPSIS
			pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
			return 0
		}
	case wmEraseBkgnd:
		if !isEnabled(hwnd) {
			return 1
		}
	case wmNcDestroy:
		pRemoveWindowSubclass.Call(hwnd, editCB, id)
	}
	r, _, _ := pDefSubclassProc.Call(hwnd, msg, wParam, lParam)
	return r
}
