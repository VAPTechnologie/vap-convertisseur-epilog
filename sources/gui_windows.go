//go:build windows

package main

// Interface graphique Win32 native (aucune dépendance externe).

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

// ---------------------------------------------------------------------------
// API Win32

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	comctl32 = syscall.NewLazyDLL("comctl32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")

	pRegisterClassExW     = user32.NewProc("RegisterClassExW")
	pCreateWindowExW      = user32.NewProc("CreateWindowExW")
	pDefWindowProcW       = user32.NewProc("DefWindowProcW")
	pGetMessageW          = user32.NewProc("GetMessageW")
	pPeekMessageW         = user32.NewProc("PeekMessageW")
	pTranslateMessage     = user32.NewProc("TranslateMessage")
	pDispatchMessageW     = user32.NewProc("DispatchMessageW")
	pIsDialogMessageW     = user32.NewProc("IsDialogMessageW")
	pPostQuitMessage      = user32.NewProc("PostQuitMessage")
	pSendMessageW         = user32.NewProc("SendMessageW")
	pShowWindow           = user32.NewProc("ShowWindow")
	pUpdateWindow         = user32.NewProc("UpdateWindow")
	pDestroyWindow        = user32.NewProc("DestroyWindow")
	pLoadCursorW          = user32.NewProc("LoadCursorW")
	pSetCursor            = user32.NewProc("SetCursor")
	pLoadImageW           = user32.NewProc("LoadImageW")
	pMessageBoxW          = user32.NewProc("MessageBoxW")
	pGetWindowTextW       = user32.NewProc("GetWindowTextW")
	pGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	pSetWindowTextW       = user32.NewProc("SetWindowTextW")
	pEnableWindow         = user32.NewProc("EnableWindow")
	pBeginPaint           = user32.NewProc("BeginPaint")
	pEndPaint             = user32.NewProc("EndPaint")
	pFillRect             = user32.NewProc("FillRect")
	pGetClientRect        = user32.NewProc("GetClientRect")
	pDrawTextW            = user32.NewProc("DrawTextW")
	pDrawIconEx           = user32.NewProc("DrawIconEx")
	pGetSystemMetrics     = user32.NewProc("GetSystemMetrics")
	pAdjustWindowRectEx   = user32.NewProc("AdjustWindowRectEx")
	pGetDC                = user32.NewProc("GetDC")
	pReleaseDC            = user32.NewProc("ReleaseDC")
	pGetDpiForSystem      = user32.NewProc("GetDpiForSystem")
	pInvalidateRect       = user32.NewProc("InvalidateRect")
	pSetFocus             = user32.NewProc("SetFocus")
	pMoveWindow           = user32.NewProc("MoveWindow")

	pCreateFontW      = gdi32.NewProc("CreateFontW")
	pCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	pCreatePen        = gdi32.NewProc("CreatePen")
	pSelectObject     = gdi32.NewProc("SelectObject")
	pSetBkMode        = gdi32.NewProc("SetBkMode")
	pSetBkColor       = gdi32.NewProc("SetBkColor")
	pSetTextColor     = gdi32.NewProc("SetTextColor")
	pEllipse          = gdi32.NewProc("Ellipse")
	pMoveToEx         = gdi32.NewProc("MoveToEx")
	pLineTo           = gdi32.NewProc("LineTo")
	pGetDeviceCaps    = gdi32.NewProc("GetDeviceCaps")
	pGetStockObject   = gdi32.NewProc("GetStockObject")

	pGetModuleHandleW         = kernel32.NewProc("GetModuleHandleW")
	pGetUserDefaultUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")

	pInitCommonControlsEx = comctl32.NewProc("InitCommonControlsEx")

	pShellExecuteW               = shell32.NewProc("ShellExecuteW")
	pSHCreateItemFromParsingName = shell32.NewProc("SHCreateItemFromParsingName")
	pCoInitializeEx              = ole32.NewProc("CoInitializeEx")
	pCoCreateInstance            = ole32.NewProc("CoCreateInstance")
	pCoTaskMemFree               = ole32.NewProc("CoTaskMemFree")
)

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type point struct{ x, y int32 }

type msgT struct {
	hwnd     uintptr
	message  uint32
	wParam   uintptr
	lParam   uintptr
	time     uint32
	pt       point
	lPrivate uint32
}

type rect struct{ l, t, r, b int32 }

type paintStruct struct {
	hdc         uintptr
	fErase      int32
	rcPaint     rect
	fRestore    int32
	fIncUpdate  int32
	rgbReserved [32]byte
}

type initCC struct{ dwSize, dwICC uint32 }

type guid struct {
	d1     uint32
	d2, d3 uint16
	d4     [8]byte
}

type filterSpec struct{ name, spec *uint16 }

const (
	wsOverlapped   = 0x00000000
	wsCaption      = 0x00C00000
	wsSysMenu      = 0x00080000
	wsMinimizeBox  = 0x00020000
	wsChild        = 0x40000000
	wsVisible      = 0x10000000
	wsTabStop      = 0x00010000
	wsVScroll      = 0x00200000
	wsClipChildren = 0x02000000
	wsExClientEdge = 0x00000200
	wsExControlPar = 0x00010000

	esAutoHScroll = 0x0080
	esMultiline   = 0x0004
	esReadOnly    = 0x0800
	esAutoVScroll = 0x0040

	bsPushButton    = 0x0
	bsDefPushButton = 0x1
	bsAutoCheckbox  = 0x3
	bsMultiline     = 0x2000

	cbsDropDownList   = 0x0003
	cbsOwnerDrawFixed = 0x0010
	cbsHasStrings     = 0x0200

	ssLeft      = 0x0
	ssRight     = 0x2
	ssEndEllip  = 0x4000
	ssCenterImg = 0x0200

	wmDestroy        = 0x0002
	wmPaint          = 0x000F
	wmClose          = 0x0010
	wmSetFont        = 0x0030
	wmCommand        = 0x0111
	wmCtlColorEdit   = 0x0133
	wmCtlColorStatic = 0x0138
	wmSetCursor      = 0x0020

	bmGetCheck = 0x00F0
	bmSetCheck = 0x00F1

	cbAddString    = 0x0143
	cbGetCurSel    = 0x0147
	cbResetContent = 0x014B
	cbSetCurSel    = 0x014E
	cbnSelChange   = 1
	emSetSel       = 0x00B1
	emReplaceSel   = 0x00C2
	pbmSetRange32  = 0x0406
	pbmSetPos      = 0x0402
	pbmSetBarColor = 0x0409

	mbOK           = 0x0
	mbYesNo        = 0x4
	mbIconError    = 0x10
	mbIconQuestion = 0x20
	mbIconWarning  = 0x30
	mbIconInfo     = 0x40
	idYes          = 6

	dtLeft       = 0x0
	dtVCenter    = 0x4
	dtSingleLine = 0x20
	dtEndEllip   = 0x8000
	dtNoPrefix   = 0x800
)

// IDs des contrôles
const (
	idDatEdit = 100 + iota
	idDatBrowse
	idRecursive
	idDatInfo
	idBkEdit
	idBkBrowse
	idMachine
	idFolderName
	idAir
	idSort
	idTop
	idLaser
	idProgress
	idLog
	idOpen
	idConvert
	idQuit
	idHelp
	idHelpText
	idHelpClose
	idLang
	idLblMachine
	idLblFolder
	idLblSort
	idLblLaser
	idDonate
	idSuggest
	idUpdate
)

// Page de don (PayPal.Me de VAP Technologie), ouverte dans le navigateur.
const donateURL = "https://www.paypal.com/paypalme/VAPTechnologie"

// Suggestions et bugs : e-mail préparé (tout le monde) ; le message rappelle la page GitHub.
const (
	suggestMail = "vapfrenchtec@gmail.com"
	issuesURL   = "https://github.com/VAPTechnologie/vap-convertisseur-epilog/issues"
)

func w16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }
func wp(s string) uintptr  { return uintptr(unsafe.Pointer(w16(s))) }
func rgb(r, g, b uint32) uintptr {
	return uintptr(r | g<<8 | b<<16)
}

func send(h uintptr, m uint32, w, l uintptr) uintptr {
	r, _, _ := pSendMessageW.Call(h, uintptr(m), w, l)
	return r
}

func getText(h uintptr) string {
	n, _, _ := pGetWindowTextLengthW.Call(h)
	buf := make([]uint16, n+1)
	pGetWindowTextW.Call(h, uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}

func setText(h uintptr, s string) { pSetWindowTextW.Call(h, wp(s)) }

func enable(h uintptr, on bool) {
	v := uintptr(0)
	if on {
		v = 1
	}
	pEnableWindow.Call(h, v)
}


// ---------------------------------------------------------------------------
// Boîtes de dialogue modernes (IFileOpenDialog)

var (
	clsidFileOpenDialog = guid{0xDC1C5A9C, 0xE88A, 0x4DDE, [8]byte{0xA5, 0xA1, 0x60, 0xF8, 0x2A, 0x20, 0xAE, 0xF7}}
	iidFileOpenDialog   = guid{0xD57C7288, 0xD4AD, 0x4768, [8]byte{0xBE, 0x02, 0x9D, 0x96, 0x95, 0x32, 0xD9, 0x60}}
	iidShellItem        = guid{0x43826D1E, 0xE718, 0x42EE, [8]byte{0xBC, 0x55, 0xA1, 0xE2, 0x61, 0xC3, 0x7B, 0xFE}}
)

func comCall(obj uintptr, idx int, args ...uintptr) uintptr {
	vtbl := *(*uintptr)(unsafe.Pointer(obj))
	fn := *(*uintptr)(unsafe.Pointer(vtbl + uintptr(idx)*unsafe.Sizeof(uintptr(0))))
	r, _, _ := syscall.SyscallN(fn, append([]uintptr{obj}, args...)...)
	return r
}

func pwstrToString(p uintptr) string {
	if p == 0 {
		return ""
	}
	var s []uint16
	for i := uintptr(0); ; i += 2 {
		c := *(*uint16)(unsafe.Pointer(p + i))
		if c == 0 {
			break
		}
		s = append(s, c)
	}
	return syscall.UTF16ToString(s)
}

func pickPath(owner uintptr, folder bool, title, start string, filters [][2]string) (string, bool) {
	var dlg uintptr
	hr, _, _ := pCoCreateInstance.Call(uintptr(unsafe.Pointer(&clsidFileOpenDialog)), 0, 1,
		uintptr(unsafe.Pointer(&iidFileOpenDialog)), uintptr(unsafe.Pointer(&dlg)))
	if int32(hr) < 0 || dlg == 0 {
		A.log(Tf("info.modernUnavail", uint32(hr)))
		return pickClassic(owner, folder, title, start, filters)
	}
	defer comCall(dlg, 2)
	var opts uint32
	comCall(dlg, 10, uintptr(unsafe.Pointer(&opts)))
	opts |= 0x40 | 0x800 | 0x1000 // FORCEFILESYSTEM | PATHMUSTEXIST | FILEMUSTEXIST
	if folder {
		opts |= 0x20 // PICKFOLDERS
	}
	comCall(dlg, 9, uintptr(opts))
	comCall(dlg, 17, wp(title))
	var specs []filterSpec
	if !folder && len(filters) > 0 {
		for _, f := range filters {
			specs = append(specs, filterSpec{w16(f[0]), w16(f[1])})
		}
		comCall(dlg, 4, uintptr(len(specs)), uintptr(unsafe.Pointer(&specs[0])))
	}
	if start != "" {
		if st, err := os.Stat(start); err == nil {
			dir := start
			if !st.IsDir() {
				dir = filepath.Dir(start)
			}
			var item uintptr
			hr, _, _ := pSHCreateItemFromParsingName.Call(wp(dir), 0, uintptr(unsafe.Pointer(&iidShellItem)), uintptr(unsafe.Pointer(&item)))
			if int32(hr) >= 0 && item != 0 {
				comCall(dlg, 12, item)
				comCall(item, 2)
			}
		}
	}
	hr = comCall(dlg, 3, owner)
	runtime.KeepAlive(specs)
	if int32(hr) < 0 {
		if uint32(hr) != 0x800704C7 { // annulé
			A.log(Tf("info.modernErr", uint32(hr)))
			return pickClassic(owner, folder, title, start, filters)
		}
		return "", false
	}
	var item uintptr
	if int32(comCall(dlg, 20, uintptr(unsafe.Pointer(&item)))) < 0 || item == 0 {
		return "", false
	}
	defer comCall(item, 2)
	var psz uintptr
	if int32(comCall(item, 5, 0x80058000, uintptr(unsafe.Pointer(&psz)))) < 0 {
		return "", false
	}
	s := pwstrToString(psz)
	pCoTaskMemFree.Call(psz)
	return s, s != ""
}

// ---------------------------------------------------------------------------
// Application

type app struct {
	hInst, hwnd                          uintptr
	dpi                                  int
	fontUI, fontBold, fontTitle, fontSub uintptr
	fontSmall, fontSection, fontIcon     uintptr
	icoBig                               uintptr
	ctl                                  map[int]uintptr
	machines                             []Machine
	outDir                               string
	report                               string // dernier rapport d'erreur écrit
	busy                                 bool
	cards                                []card
	titleRight, footerRight              int  // limites des textes dessinés (unités logiques)
	inWelcome, welcomeOnly               bool // journal = seulement le message d'accueil
	newRelease                           release // version plus récente trouvée sur GitHub
	compact                              bool    // petit écran : en-tête et marges réduits
	cliH                                 int     // hauteur de la zone cliente (unités logiques)
	frames                               map[int]rect // cadres des champs de saisie (pixels)
	progRc, logRc                        rect         // barre de progression et journal (unités logiques)
	footY                                int          // haut du pied de page (unités logiques)
	progress                             int          // 0 à 100
	richLog                              bool         // journal en couleurs (RichEdit)
	realDpi                              int          // DPI réel du système (texte des RichEdit, en points)
	logBuf                               strings.Builder
}

// card : panneau arrondi d'une étape (unités logiques).
type card struct {
	y, h  int
	num   string
	title string
}

var A = &app{ctl: map[int]uintptr{}, frames: map[int]rect{}}

const cliW = 860

// metrics : dimensions de la fenêtre principale (unités logiques), normales ou compactes
// (écran 1366 × 768, mise à l'échelle 125 % ou 150 %).
type metrics struct {
	hdr, top, pad, title, gap, row, chk, cardGap, prog, foot, btn, logo, logMin, logDesign int
}

var (
	normalM  = metrics{hdr: 76, top: 14, pad: 12, title: 26, gap: 8, row: 30, chk: 24, cardGap: 10, prog: 8, foot: 58, btn: 34, logo: 28, logMin: 84, logDesign: 150}
	compactM = metrics{hdr: 56, top: 6, pad: 6, title: 22, gap: 4, row: 26, chk: 22, cardGap: 5, prog: 6, foot: 46, btn: 30, logo: 22, logMin: 80, logDesign: 130}
)

func (a *app) m() metrics {
	if a.compact {
		return compactM
	}
	return normalM
}

// hauteurs des quatre panneaux
func (m metrics) cards() [4]int {
	two := m.pad*2 + m.title + m.gap + m.row
	return [4]int{two + m.gap + m.chk, two, two, two + m.gap + m.row}
}

// space : espace autour de la progression et du journal.
func (m metrics) space() int { return min(10, m.pad+4) }

// clientHeight : hauteur totale pour un journal de hauteur log.
func (m metrics) clientHeight(log int) int {
	c := m.cards()
	return m.hdr + 3 + m.top + c[0] + c[1] + c[2] + c[3] + 3*m.cardGap + m.space() + m.prog + m.space() + log + m.space() + m.foot
}

func (a *app) s(v int) int { return v * a.dpi / 96 }

func (a *app) font(pt int, weight int, face string) uintptr {
	h := -(pt * a.dpi / 72)
	f, _, _ := pCreateFontW.Call(uintptr(int32(h)), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, wp(face))
	return f
}

func (a *app) child(class string, text string, style uint32, exStyle uint32, x, y, w, h, id int) uintptr {
	hw, _, _ := pCreateWindowExW.Call(uintptr(exStyle), wp(class), wp(text),
		uintptr(style|wsChild|wsVisible), uintptr(a.s(x)), uintptr(a.s(y)), uintptr(a.s(w)), uintptr(a.s(h)),
		a.hwnd, uintptr(id), a.hInst, 0)
	send(hw, wmSetFont, a.fontUI, 1)
	a.ctl[id] = hw
	return hw
}

func (a *app) label(id int, style uint32) uintptr {
	hw, _, _ := pCreateWindowExW.Call(0, wp("STATIC"), wp(""), uintptr(style|wsChild|wsVisible|ssCenterImg),
		0, 0, 10, 10, a.hwnd, uintptr(id), a.hInst, 0)
	send(hw, wmSetFont, a.fontUI, 1)
	a.ctl[id] = hw
	return hw
}

// buildUI : crée les contrôles ; textes et positions sont posés par applyLanguage / layout.
func (a *app) buildUI() {
	const combo = cbsDropDownList | cbsOwnerDrawFixed | cbsHasStrings | wsVScroll | wsTabStop
	a.child("BUTTON", "", bsPushButton|wsTabStop, 0, 0, 0, 10, 10, idHelp)
	// liste des langues avec le drapeau du pays (flags_windows.go)
	a.child("COMBOBOX", "", combo, 0, 0, 0, 10, 300, idLang)
	for i := Lang(0); i < nLang; i++ {
		send(a.ctl[idLang], cbAddString, 0, wp(langNames[i]))
	}
	a.child("EDIT", "", esAutoHScroll|wsTabStop, 0, 0, 0, 10, 10, idDatEdit)
	a.child("BUTTON", "", bsPushButton|wsTabStop, 0, 0, 0, 10, 10, idDatBrowse)
	a.child("BUTTON", "", bsAutoCheckbox|wsTabStop, 0, 0, 0, 10, 10, idRecursive)
	a.child("STATIC", "", ssRight|ssEndEllip|ssCenterImg, 0, 0, 0, 10, 10, idDatInfo)
	a.child("EDIT", "", esAutoHScroll|wsTabStop, 0, 0, 0, 10, 10, idBkEdit)
	a.child("BUTTON", "", bsPushButton|wsTabStop, 0, 0, 0, 10, 10, idBkBrowse)
	a.label(idLblMachine, ssLeft)
	a.child("COMBOBOX", "", combo, 0, 0, 0, 10, 300, idMachine)
	a.label(idLblFolder, ssLeft)
	a.child("EDIT", "", esAutoHScroll|wsTabStop, 0, 0, 0, 10, 10, idFolderName)
	a.child("BUTTON", "", bsAutoCheckbox|bsMultiline|wsTabStop, 0, 0, 0, 10, 10, idAir)
	a.label(idLblSort, ssLeft)
	a.child("COMBOBOX", "", combo, 0, 0, 0, 10, 200, idSort)
	a.child("BUTTON", "", bsAutoCheckbox|bsMultiline|wsTabStop, 0, 0, 0, 10, 10, idTop)
	a.label(idLblLaser, ssLeft)
	a.child("COMBOBOX", "", combo, 0, 0, 0, 10, 200, idLaser)
	a.richLog = a.newLog()
	a.child("BUTTON", "", bsPushButton|wsTabStop, 0, 0, 0, 10, 10, idOpen)
	a.child("BUTTON", "", bsDefPushButton|wsTabStop, 0, 0, 0, 10, 10, idConvert)
	a.child("BUTTON", "", bsPushButton|wsTabStop, 0, 0, 0, 10, 10, idQuit)
	a.child("BUTTON", "", bsPushButton|wsTabStop, 0, 0, 0, 10, 10, idDonate)
	a.child("BUTTON", "", bsPushButton|wsTabStop, 0, 0, 0, 10, 10, idSuggest)
	a.child("STATIC", "", ssLeft|ssNotify|ssEndEllip|ssCenterImg, 0, 0, 0, 10, 10, idUpdate)
	enable(a.ctl[idOpen], false)
	a.themeCombo(a.ctl[idLang], colFond)
	for _, id := range []int{idMachine, idSort, idLaser} {
		a.themeCombo(a.ctl[id], colPanneau)
	}

	send(a.ctl[idAir], bmSetCheck, 1, 0)
	enable(a.ctl[idAir], false) // imposée par le moteur, affichée pour information
	send(a.ctl[idTop], bmSetCheck, 1, 0)

	a.applyLanguage(true)
}

// applyLanguage : (re)pose tous les textes dans la langue courante.
func (a *app) applyLanguage(first bool) {
	a.makeFonts()
	send(a.ctl[idLang], cbSetCurSel, uintptr(cur), 0)
	setText(a.hwnd, T("app.title"))
	for id, k := range map[int]string{
		idHelp: "btn.help", idDatBrowse: "btn.browse", idBkBrowse: "btn.browse", idRecursive: "chk.recursive",
		idLblMachine: "lbl.machine", idLblFolder: "lbl.folder", idAir: "chk.air", idLblSort: "lbl.sort",
		idTop: "chk.top", idLblLaser: "lbl.laser", idOpen: "btn.open", idConvert: "btn.convert", idQuit: "btn.quit", idDonate: "btn.donate", idSuggest: "btn.suggest",
	} {
		setText(a.ctl[id], T(k))
	}
	fill := func(id int, keys []string, def int) {
		sel := int(int32(send(a.ctl[id], cbGetCurSel, 0, 0)))
		if first || sel < 0 {
			sel = def
		}
		send(a.ctl[id], cbResetContent, 0, 0)
		for _, k := range keys {
			send(a.ctl[id], cbAddString, 0, wp(T(k)))
		}
		send(a.ctl[id], cbSetCurSel, uintptr(sel), 0)
	}
	fill(idSort, []string{"sort.none", "sort.inout", "sort.opt", "sort.dat"}, 1)
	fill(idLaser, []string{"laser.co2", "laser.fiber"}, 0)
	a.countDats(false)
	a.applyUpdateText()
	a.layout()
	if first || a.welcomeOnly {
		a.showWelcome()
	}
	pInvalidateRect.Call(a.hwnd, 0, 1)
}

func (a *app) showWelcome() {
	a.clearLog()
	a.inWelcome = true
	for i := 0; i <= 5; i++ {
		a.log(T(fmt.Sprintf("welcome.%d", i)))
	}
	if a.newRelease.tag != "" {
		a.log(a.updateLogLine())
	}
	a.inWelcome = false
	a.welcomeOnly = true
	send(a.ctl[idLog], 0x0115, 6, 0) // WM_VSCROLL SB_TOP
}

// uiFaces : polices adaptées à la langue (Segoe UI n'a pas les caractères chinois / japonais).
func uiFaces() (face, bold string, bw int) {
	switch cur {
	case LZH:
		return "Microsoft YaHei UI", "Microsoft YaHei UI", 700
	case LJA:
		return "Yu Gothic UI", "Yu Gothic UI Semibold", 600
	}
	return "Segoe UI", "Segoe UI Semibold", 600
}

// makeFonts : polices de l'interface ; titres en Bahnschrift comme l'Analyseur Wi-Fi.
func (a *app) makeFonts() {
	old := []uintptr{a.fontUI, a.fontSmall, a.fontBold, a.fontTitle, a.fontSub, a.fontSection, a.fontIcon}
	face, bold, bw := uiFaces()
	display, dw := "Bahnschrift", 600
	if cur == LZH || cur == LJA {
		display, dw = bold, bw
	}
	titlePt, secPt := 17, 12
	if a.compact {
		titlePt, secPt = 15, 11
	}
	a.fontUI = a.font(10, 400, face)
	a.fontSmall = a.font(9, 400, face)
	a.fontBold = a.font(10, bw, bold)
	a.fontTitle = a.font(titlePt, dw, display)
	a.fontSection = a.font(secPt, dw, display)
	a.fontSub = a.font(10, 400, face)
	a.fontIcon = a.font(15, 700, "Segoe UI")
	for id, h := range a.ctl {
		if id != idLog {
			send(h, wmSetFont, a.fontUI, 1)
		}
	}
	a.logFont(face, 9)
	send(a.ctl[idConvert], wmSetFont, a.fontBold, 1)
	for _, f := range old {
		if f != 0 {
			pDeleteObject.Call(f)
		}
	}
}

// measure : largeur d'un texte en unités logiques (96 dpi).
func (a *app) measure(font uintptr, s string) int {
	dc, _, _ := pGetDC.Call(a.hwnd)
	old, _, _ := pSelectObject.Call(dc, font)
	r := rect{0, 0, 4000, 200}
	pDrawTextW.Call(dc, wp(s), ^uintptr(0), uintptr(unsafe.Pointer(&r)), dtLeft|dtSingleLine|dtNoPrefix|0x400) // DT_CALCRECT
	pSelectObject.Call(dc, old)
	pReleaseDC.Call(a.hwnd, dc)
	return int(r.r)*96/a.dpi + 1
}

// px : rectangle logique -> pixels.
func (a *app) px(r rect) rect {
	return rect{int32(a.s(int(r.l))), int32(a.s(int(r.t))), int32(a.s(int(r.r))), int32(a.s(int(r.b)))}
}

// layout : place les contrôles en fonction de la longueur des textes de la langue.
func (a *app) layout() {
	m := a.m()
	x0, x1 := 20, cliW-20 // bords des panneaux
	ix0, ix1 := x0+16, x1-16
	mv := func(id, x, y, w, h int) {
		pMoveWindow.Call(a.ctl[id], uintptr(a.s(x)), uintptr(a.s(y)), uintptr(a.s(w)), uintptr(a.s(h)), 1)
	}
	eh := int(textHeight(a.hwnd, a.fontUI)) + a.s(2)
	// champ de saisie : cadre arrondi dessiné par la fenêtre, texte centré dedans
	field := func(id, x, y, w, h int) {
		f := a.px(rect{int32(x), int32(y), int32(x + w), int32(y + h)})
		a.frames[id] = f
		pMoveWindow.Call(a.ctl[id], uintptr(f.l+int32(a.s(10))), uintptr(f.t+(f.b-f.t-int32(eh))/2),
			uintptr(f.r-f.l-int32(a.s(20))), uintptr(eh), 1)
	}
	// liste déroulante : hauteur de la partie fermée = hauteur du champ
	combo := func(id, x, y, w, h, list int) {
		send(a.ctl[id], cbSetItemH, ^uintptr(0), uintptr(a.s(h)-6))
		send(a.ctl[id], cbSetItemH, 0, uintptr(a.s(24)))
		mv(id, x, y, w, list)
	}
	tw := func(k string) int { return a.measure(a.fontUI, T(k)) }

	// en-tête : notice + langue, à droite
	lw := 0
	for _, n := range langNames {
		lw = max(lw, a.measure(a.fontUI, n))
	}
	lw += 10 + 20 + 8 + 10 + 10 + 6 // marges, drapeau, chevron
	hw := max(150, tw("btn.help")+36)
	hy := (m.hdr - m.btn) / 2
	combo(idLang, x1-lw, hy, lw, m.btn, 300)
	mv(idHelp, x1-lw-8-hw, hy, hw, m.btn)
	a.titleRight = x1 - lw - 8 - hw - 16

	a.cards = a.cards[:0]
	ch := m.cards()
	y := m.hdr + 3 + m.top
	bw := max(118, tw("btn.browse")+32)

	// 1 : dossier des .DAT
	a.cards = append(a.cards, card{y, ch[0], "1", T("sec1")})
	ry := y + m.pad + m.title + m.gap
	field(idDatEdit, ix0, ry, ix1-ix0-bw-10, m.row)
	mv(idDatBrowse, ix1-bw, ry, bw, m.row)
	ry += m.row + m.gap
	cw := tw("chk.recursive") + 30
	mv(idRecursive, ix0, ry, cw, m.chk)
	mv(idDatInfo, ix0+cw+10, ry, ix1-(ix0+cw+10), m.chk)
	y += ch[0] + m.cardGap

	// 2 : sauvegarde .epidb
	a.cards = append(a.cards, card{y, ch[1], "2", T("sec2")})
	ry = y + m.pad + m.title + m.gap
	field(idBkEdit, ix0, ry, ix1-ix0-bw-10, m.row)
	mv(idBkBrowse, ix1-bw, ry, bw, m.row)
	y += ch[1] + m.cardGap

	// 3 : machine et nom du dossier
	a.cards = append(a.cards, card{y, ch[2], "3", T("sec3")})
	ry = y + m.pad + m.title + m.gap
	w1 := tw("lbl.machine") + 10
	w2 := tw("lbl.folder") + 10
	comboW := 270
	if free := (ix1 - ix0) - w1 - 18 - w2 - 160; free < comboW {
		comboW = max(free, 160)
	}
	mv(idLblMachine, ix0, ry, w1, m.row)
	combo(idMachine, ix0+w1, ry, comboW, m.row, 300)
	x := ix0 + w1 + comboW + 18
	mv(idLblFolder, x, ry, w2, m.row)
	field(idFolderName, x+w2, ry, ix1-x-w2, m.row)
	y += ch[2] + m.cardGap

	// 4 : options
	a.cards = append(a.cards, card{y, ch[3], "4", T("sec4")})
	ry = y + m.pad + m.title + m.gap
	cbw := 150
	for _, k := range []string{"sort.none", "sort.inout", "sort.opt", "sort.dat", "laser.co2", "laser.fiber"} {
		cbw = max(cbw, tw(k)+46)
	}
	lw2 := max(tw("lbl.sort"), tw("lbl.laser")) + 12
	gx := ix1 - cbw - lw2
	chkW := gx - 16 - ix0
	mv(idAir, ix0, ry, chkW, m.row)
	mv(idLblSort, gx, ry, lw2, m.row)
	combo(idSort, gx+lw2, ry, cbw, m.row, 200)
	ry += m.row + m.gap
	mv(idTop, ix0, ry, chkW, m.row)
	mv(idLblLaser, gx, ry, lw2, m.row)
	combo(idLaser, gx+lw2, ry, cbw, m.row, 200)
	y += ch[3] + m.space()

	// progression, journal, lien « nouvelle version »
	a.progRc = rect{int32(x0), int32(y), int32(x1), int32(y + m.prog)}
	y += m.prog + m.space()
	a.footY = a.cliH - m.foot
	logB := a.footY - m.space()
	if a.newRelease.tag != "" {
		logB -= 22
		mv(idUpdate, x0+4, logB+3, min(x1-x0-4, a.measure(a.fontUI, Tf("upd.link", a.newRelease.tag))+4), 20)
	}
	a.logRc = rect{int32(x0), int32(y), int32(x1), int32(logB)}
	vm := min(6, m.pad-2) // marge haute et basse du texte
	mv(idLog, x0+12, y+vm, x1-x0-16, logB-y-2*vm)
	if a.richLog {
		// marge intérieure du texte
		var r rect
		pGetClientRect.Call(a.ctl[idLog], uintptr(unsafe.Pointer(&r)))
		r.l += int32(a.s(2))
		r.t += int32(a.s(2))
		send(a.ctl[idLog], emSetRect, 0, uintptr(unsafe.Pointer(&r)))
	}

	// pied de page : signature, don, suggestion | ouvrir, convertir, quitter
	by := a.footY + (m.foot-m.btn)/2
	qw := max(110, tw("btn.quit")+36)
	cvw := max(150, a.measure(a.fontBold, T("btn.convert"))+36)
	ow := max(150, tw("btn.open")+36)
	mv(idQuit, x1-qw, by, qw, m.btn)
	mv(idConvert, x1-qw-8-cvw, by, cvw, m.btn)
	mv(idOpen, x1-qw-8-cvw-8-ow, by, ow, m.btn)
	sigW := a.measure(a.fontSmall, appSignature)
	dw := max(110, tw("btn.donate")+32)
	mv(idDonate, x0+sigW+14, by, dw, m.btn)
	sw := max(110, tw("btn.suggest")+32)
	mv(idSuggest, x0+sigW+14+dw+8, by, sw, m.btn)
	a.footerRight = x0 + sigW + 4
}

func (a *app) checked(id int) bool { return send(a.ctl[id], bmGetCheck, 0, 0) == 1 }

// setProgress : barre de progression (valeur réelle du moteur, 0 à 100).
func (a *app) setProgress(p int) {
	a.progress = p
	r := a.px(a.progRc)
	pInvalidateRect.Call(a.hwnd, uintptr(unsafe.Pointer(&r)), 0)
	pUpdateWindow.Call(a.hwnd)
}

var (
	pCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	pBitBlt                 = gdi32.NewProc("BitBlt")
)

// paintBuffered : dessin dans une image en mémoire, puis copie (pas de scintillement).
func paintBuffered(hwnd uintptr, draw func(hdc uintptr)) {
	var ps paintStruct
	hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	var rc rect
	pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	mem, _, _ := pCreateCompatibleDC.Call(hdc)
	bmp, _, _ := pCreateCompatibleBitmap.Call(hdc, uintptr(rc.r), uintptr(rc.b))
	if mem == 0 || bmp == 0 {
		draw(hdc)
	} else {
		ob, _, _ := pSelectObject.Call(mem, bmp)
		draw(mem)
		p := ps.rcPaint
		pBitBlt.Call(hdc, uintptr(p.l), uintptr(p.t), uintptr(p.r-p.l), uintptr(p.b-p.t), mem, uintptr(p.l), uintptr(p.t), 0x00CC0020) // SRCCOPY
		pSelectObject.Call(mem, ob)
	}
	if bmp != 0 {
		pDeleteObject.Call(bmp)
	}
	if mem != 0 {
		pDeleteDC.Call(mem)
	}
	pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
}

// paintHeader : bandeau commun (fenêtre principale et notice) : dégradé bleu nuit comme
// l'Analyseur Wi-Fi, logo VAP, titre, sous-titre éventuel, filet bleu de 3 px.
func (a *app) paintHeader(hdc uintptr, width int32, hdr, logo int, title, sub string, titleRight int) {
	H := int32(a.s(hdr))
	hGradient(hdc, rect{0, 0, width * 6 / 10, H}, colEnteteDebut, colFond)
	fillBox(hdc, width*6/10, 0, width, H, colFond)
	fillBox(hdc, 0, H, width, H+int32(a.s(3)), colAccent)
	tx := a.logoTitleX(hdc, 22, hdr/2, logo)
	if tx == 22 && a.icoBig != 0 { // repli sur l'icône si le logo ne peut pas être dessiné
		d := hdr - 20
		pDrawIconEx.Call(hdc, uintptr(a.s(22)), uintptr(a.s(10)), a.icoBig, uintptr(a.s(d)), uintptr(a.s(d)), 0, 0, 3)
		tx = 22 + d + 14
	}
	th := textHeight(a.hwnd, a.fontTitle)
	if sub == "" {
		drawText(hdc, title, rect{int32(a.s(tx)), 0, int32(a.s(titleRight)), H}, a.fontTitle, colTexte, dtLeft|dtVCenter|dtSingleLine|dtEndEllip)
		return
	}
	sh := textHeight(a.hwnd, a.fontSub)
	top := (H - th - sh) / 2
	drawText(hdc, title, rect{int32(a.s(tx)), top, int32(a.s(titleRight)), top + th}, a.fontTitle, colTexte, dtLeft|dtVCenter|dtSingleLine|dtEndEllip)
	drawText(hdc, sub, rect{int32(a.s(tx + 1)), top + th, int32(a.s(titleRight)), top + th + sh}, a.fontSub, colDiscret, dtLeft|dtVCenter|dtSingleLine|dtEndEllip)
}

func (a *app) paint(hdc uintptr) {
	m := a.m()
	var rc rect
	pGetClientRect.Call(a.hwnd, uintptr(unsafe.Pointer(&rc)))
	fillBox(hdc, 0, 0, rc.r, rc.b, colFond)
	a.paintHeader(hdc, rc.r, m.hdr, m.logo, T("hdr.title"), T("hdr.sub"), a.titleRight)

	// étapes : panneau arrondi, numéro dans un rond bleu, titre blanc, filet
	x0, x1 := 20, cliW-20
	for _, c := range a.cards {
		roundBox(hdc, a.px(rect{int32(x0), int32(c.y), int32(x1), int32(c.y + c.h)}), int32(a.s(12)), colPanneau, colFilet)
		ty := c.y + m.pad
		d := 22
		cy := ty + m.title/2
		disc(hdc, int32(a.s(x0+16+d/2)), int32(a.s(cy)), int32(a.s(d/2)), colAccent)
		drawText(hdc, c.num, a.px(rect{int32(x0 + 16), int32(cy - d/2), int32(x0 + 16 + d), int32(cy + d/2)}), a.fontBold, colTexteAccent, dtCenter|dtVCenter|dtSingleLine)
		tl := x0 + 16 + d + 10
		drawText(hdc, c.title, a.px(rect{int32(tl), int32(ty), int32(x1 - 16), int32(ty + m.title)}), a.fontSection, colTexte, dtLeft|dtVCenter|dtSingleLine|dtEndEllip)
		lx := tl + a.measure(a.fontSection, c.title) + 10
		if lx < x1-16 {
			fillBox(hdc, int32(a.s(lx)), int32(a.s(cy)), int32(a.s(x1-16)), int32(a.s(cy))+int32(max(1, a.s(1))), colFilet)
		}
	}

	// champs de saisie : bord bleu quand ils ont le focus
	focus, _, _ := pGetFocus.Call()
	for id, f := range a.frames {
		bd := colFilet
		if focus == a.ctl[id] {
			bd = colAccent
		}
		roundBox(hdc, f, int32(a.s(8)), colPanneau, bd)
	}

	// progression : seulement la valeur transmise par la conversion
	p := a.px(a.progRc)
	rad := (p.b - p.t) / 2
	roundBox(hdc, p, rad, colDoux, colFilet)
	if a.progress > 0 {
		w := (p.r - p.l) * int32(min(a.progress, 100)) / 100
		if w < 2*rad {
			w = 2 * rad
		}
		roundBox(hdc, rect{p.l, p.t, p.l + w, p.b}, rad, colAccent, colAccent)
	}

	// journal
	roundBox(hdc, a.px(a.logRc), int32(a.s(10)), colPanneau, colFilet)

	// pied de page
	fy := int32(a.s(a.footY))
	fillBox(hdc, 0, fy, rc.r, rc.b, colPanneau)
	fillBox(hdc, 0, fy, rc.r, fy+int32(max(1, a.s(1))), colFilet)
	drawText(hdc, appSignature, rect{int32(a.s(x0)), fy, int32(a.s(a.footerRight)), rc.b}, a.fontSmall, colDiscret, dtLeft|dtVCenter|dtSingleLine|dtEndEllip)
}

// onNotify : dessin des boutons et des cases (NM_CUSTOMDRAW).
func (a *app) onNotify(lParam uintptr) (uintptr, bool) {
	nm := (*nmCustomDrawT)(unsafe.Pointer(lParam))
	if nm.hdr.code != nmCustomDraw {
		return 0, false
	}
	id := int(nm.hdr.idFrom)
	switch id {
	case idHelp, idDatBrowse, idBkBrowse, idOpen, idConvert, idQuit, idDonate, idSuggest, idRecursive, idAir, idTop:
	default:
		return 0, false
	}
	if nm.drawStage != cddsPrepaint {
		return cdrfDoDefault, true
	}
	back := colPanneau
	if id == idHelp {
		back = colFond
	}
	switch id {
	case idRecursive, idAir, idTop:
		a.drawCheck(nm, back)
	default:
		a.drawButton(nm, id == idConvert, back)
	}
	return cdrfSkipDef, true
}

// onCtlColor : couleurs des champs, étiquettes et listes (texte et fond).
func (a *app) onCtlColor(msg uint32, hdc, h uintptr) uintptr {
	switch {
	case h == a.ctl[idUpdate]:
		pSetBkColor.Call(hdc, colFond)
		pSetTextColor.Call(hdc, colAccentClair)
		return brFond
	case h == a.ctl[idDatInfo]:
		pSetTextColor.Call(hdc, colDiscret)
	case msg == wmCtlColorStatic && !isEnabled(h):
		pSetTextColor.Call(hdc, colInactif)
	case h == a.ctl[idLblMachine] || h == a.ctl[idLblFolder] || h == a.ctl[idLblSort] || h == a.ctl[idLblLaser]:
		pSetTextColor.Call(hdc, colDiscret)
	default:
		pSetTextColor.Call(hdc, colTexte)
	}
	pSetBkColor.Call(hdc, colPanneau)
	return brPanneau
}

// ---------------------------------------------------------------------------
// Logique de l'interface

func (a *app) countDats(autoLaser bool) {
	dir := strings.TrimSpace(getText(a.ctl[idDatEdit]))
	if dir == "" {
		setText(a.ctl[idDatInfo], "")
		return
	}
	ps, err := ListDats(dir, a.checked(idRecursive))
	if err != nil {
		setText(a.ctl[idDatInfo], T("info.unreadable"))
		return
	}
	setText(a.ctl[idDatInfo], Tf("info.count", len(ps)))
	// dossier « … fibre … » : Fibre proposée, si la machine choisie en a une
	if autoLaser && (strings.Contains(foldKey(dir), "fibre") || strings.Contains(foldKey(dir), "fiber")) {
		if m, ok := a.curMachine(); !ok || m.HasFiber() {
			send(a.ctl[idLaser], cbSetCurSel, 1, 0)
		}
	}
}

func (a *app) curMachine() (Machine, bool) {
	i := int(int32(send(a.ctl[idMachine], cbGetCurSel, 0, 0)))
	if i < 0 || i >= len(a.machines) {
		return Machine{}, false
	}
	return a.machines[i], true
}

// fixDatDir : si le dossier choisi est un dossier de résultat, propose le dossier source.
func (a *app) fixDatDir() {
	dir := strings.TrimSpace(getText(a.ctl[idDatEdit]))
	if dir == "" {
		return
	}
	src := SourceDirFor(dir, a.checked(idRecursive))
	if src == "" {
		return
	}
	ps, _ := ListDats(src, a.checked(idRecursive))
	if msgBox(a.hwnd, Tf("ask.useSource", dir, len(ps), src), T("msg.title"), mbYesNo|mbIconQuestion) == idYes {
		setText(a.ctl[idDatEdit], src)
		a.countDats(true)
	}
}

func tokens(s string) map[string]bool {
	t := map[string]bool{}
	for _, w := range strings.FieldsFunc(foldKey(s), func(r rune) bool { return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '+') }) {
		if w != "de" && w != "watts" && w != "w" {
			t[w] = true
		}
	}
	return t
}

func (a *app) bestMachine() int {
	if len(a.machines) == 0 {
		return -1
	}
	dir := filepath.Base(strings.TrimSpace(getText(a.ctl[idDatEdit])))
	best, score := 0, -1
	dt := tokens(dir)
	for i, m := range a.machines {
		sc := 0
		if foldKey(m.Name) == foldKey(dir) {
			sc = 1000
		}
		for w := range tokens(m.Name) {
			if dt[w] {
				sc += 10
			}
		}
		if sc > score {
			best, score = i, sc
		}
	}
	return best
}

func (a *app) selectMachine(i int) {
	if i < 0 || i >= len(a.machines) {
		return
	}
	send(a.ctl[idMachine], cbSetCurSel, uintptr(i), 0)
	m := a.machines[i]
	setText(a.ctl[idFolderName], m.Name)
	// source imposée quand la machine n'en a qu'une, au choix sinon
	switch {
	case !m.HasFiber():
		send(a.ctl[idLaser], cbSetCurSel, 0, 0)
	case !m.HasCO2():
		send(a.ctl[idLaser], cbSetCurSel, 1, 0)
	}
	enable(a.ctl[idLaser], m.HasCO2() && m.HasFiber())
}

func (a *app) loadBackup(path string) {
	send(a.ctl[idMachine], cbResetContent, 0, 0)
	a.machines = nil
	enable(a.ctl[idLaser], true)
	db, err := openRO(path)
	if err != nil {
		msgBox(a.hwnd, T("err.openBackup")+err.Error(), T("msg.title"), mbIconError)
		return
	}
	ms, err := LoadMachines(db)
	if err != nil {
		db.Close()
		msgBox(a.hwnd, T("err.notBackup")+err.Error(), T("msg.title"), mbIconError)
		return
	}
	err = CheckSchema(db)
	db.Close()
	if err != nil {
		msgBox(a.hwnd, err.Error(), T("msg.title"), mbIconError)
		return
	}
	if len(ms) == 0 {
		msgBox(a.hwnd, T("warn.noMachine"), T("msg.title"), mbIconWarning)
		return
	}
	a.machines = ms
	for _, m := range ms {
		send(a.ctl[idMachine], cbAddString, 0, wp(m.Name))
	}
	a.selectMachine(a.bestMachine())
	a.log(Tf("log.backupLoaded", len(ms)))
}

func (a *app) convert() {
	if a.busy {
		return
	}
	a.fixDatDir()
	dir := strings.TrimSpace(getText(a.ctl[idDatEdit]))
	bk := strings.TrimSpace(getText(a.ctl[idBkEdit]))
	folder := strings.TrimSpace(getText(a.ctl[idFolderName]))
	sel := int(int32(send(a.ctl[idMachine], cbGetCurSel, 0, 0)))
	switch {
	case dir == "":
		msgBox(a.hwnd, T("warn.chooseDat"), T("msg.title"), mbIconWarning)
		return
	case bk == "":
		msgBox(a.hwnd, T("warn.chooseBackup"), T("msg.title"), mbIconWarning)
		return
	case len(a.machines) == 0:
		a.loadBackup(bk)
		if len(a.machines) == 0 {
			return
		}
		sel = int(int32(send(a.ctl[idMachine], cbGetCurSel, 0, 0)))
	}
	if sel < 0 || sel >= len(a.machines) {
		msgBox(a.hwnd, T("warn.chooseMachine"), T("msg.title"), mbIconWarning)
		return
	}
	if folder == "" {
		folder = a.machines[sel].Name
		setText(a.ctl[idFolderName], folder)
	}
	o := Options{
		DatDir:       dir,
		Recursive:    a.checked(idRecursive),
		Machine:      a.machines[sel],
		FolderName:   folder,
		VectorSort:   int(int32(send(a.ctl[idSort], cbGetCurSel, 0, 0))),
		OnTop:        a.checked(idTop),
		Fiber:        send(a.ctl[idLaser], cbGetCurSel, 0, 0) == 1,
	}
	if o.VectorSort < 0 {
		o.VectorSort = 1
	}
	// dossier déjà présent ?
	if db, err := openRO(bk); err == nil {
		fid, _ := FindJobFolder(db, folder, o.Machine.ID)
		db.Close()
		if fid != 0 {
			if msgBox(a.hwnd, Tf("ask.replace", folder),
				T("msg.title"), mbYesNo|mbIconWarning) != idYes {
				return
			}
			o.ReplaceFolder = true
		}
	}

	a.busy = true
	// options, langue et machine figées pendant la conversion
	locked := []int{idConvert, idDatBrowse, idBkBrowse, idOpen, idDatEdit, idBkEdit, idRecursive, idMachine, idFolderName, idSort, idTop, idLaser, idLang}
	for _, id := range locked {
		enable(a.ctl[id], false)
	}
	wait, _, _ := pLoadCursorW.Call(0, 32514)
	pSetCursor.Call(wait)
	a.setProgress(0)
	a.log("")
	a.log(Tf("log.convStart", folder, o.Machine.Name))
	out, res, err := Run(openRW, bk, o, a.log, func(d, t int) {
		a.setProgress(d * 100 / t)
		pumpMessages()
	})
	a.busy = false
	for _, id := range locked {
		if id != idOpen {
			enable(a.ctl[id], true)
		}
	}
	enable(a.ctl[idLaser], o.Machine.HasCO2() && o.Machine.HasFiber())
	arrow, _, _ := pLoadCursorW.Call(0, 32512)
	pSetCursor.Call(arrow)
	if err != nil {
		a.setProgress(0)
		a.log(T("log.fail") + err.Error())
		rp, rerr := WriteReport(openRO, ReportInput{Version: appVersion(), System: reportSystem(),
			Backup: bk, Options: o, Log: a.logText(), Err: err})
		if rerr != nil {
			msgBox(a.hwnd, T("err.convFailed")+err.Error(), T("msg.title"), mbIconError)
			return
		}
		a.report = rp
		a.log(T("log.report") + rp)
		if msgBox(a.hwnd, T("err.convFailed")+err.Error()+Tf("ask.sendReport", rp)+T("rep.privacy"), T("msg.title"), mbYesNo|mbIconError) == idYes {
			pShellExecuteW.Call(a.hwnd, wp("open"), wp(rp), 0, 0, 1) // relecture avant envoi
			a.sendSuggestion()
		}
		return
	}
	a.setProgress(100)
	a.outDir = out
	enable(a.ctl[idOpen], true)
	a.log(Tf("log.done", len(res.Converted), len(res.Skipped)))
	a.log(T("log.outDir") + out)
	done := Tf("msg.done", len(res.Converted), folder)
	if n := res.Noted(); n > 0 || len(res.Skipped) > 0 {
		// bilan inséré avant la question finale (« Ouvrir le dossier maintenant ? »)
		sum := Tf("msg.summary", len(res.Converted)-n, n, len(res.Skipped))
		if i := strings.LastIndex(done, "\n\n"); i >= 0 {
			done = done[:i] + sum + done[i:]
		}
	}
	if msgBox(a.hwnd, done, T("msg.title"), mbYesNo|mbIconInfo) == idYes {
		a.openOut()
	}
}

func (a *app) openOut() {
	if a.outDir != "" {
		pShellExecuteW.Call(a.hwnd, wp("open"), wp(a.outDir), 0, 0, 1)
	}
}

func pumpMessages() {
	var m msgT
	for {
		r, _, _ := pPeekMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0, 1)
		if r == 0 {
			return
		}
		if m.message == 0x0012 { // WM_QUIT : remis dans la file pour la boucle principale
			pPostQuitMessage.Call(m.wParam)
			return
		}
		// notice ouverte pendant la conversion : Tabulation, Entrée et Échap comme d'habitude
		if helpHwnd != 0 {
			if d, _, _ := pIsDialogMessageW.Call(helpHwnd, uintptr(unsafe.Pointer(&m))); d != 0 {
				continue
			}
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func (a *app) onCommand(id int, code int) {
	switch id {
	case idDatEdit, idBkEdit, idFolderName:
		if code == enSetFocus || code == enKillFocus { // bord bleu du champ actif
			if f, ok := a.frames[id]; ok {
				f = rect{f.l - 1, f.t - 1, f.r + 1, f.b + 1}
				pInvalidateRect.Call(a.hwnd, uintptr(unsafe.Pointer(&f)), 0)
			}
		}
	case idDatBrowse:
		start := strings.TrimSpace(getText(a.ctl[idDatEdit]))
		if p, ok := pickPath(a.hwnd, true, T("dlg.datTitle"), start, nil); ok {
			setText(a.ctl[idDatEdit], p)
			a.countDats(true)
			a.fixDatDir()
			if len(a.machines) > 0 {
				a.selectMachine(a.bestMachine())
			}
		}
	case idBkBrowse:
		start := strings.TrimSpace(getText(a.ctl[idBkEdit]))
		if start == "" {
			if d := strings.TrimSpace(getText(a.ctl[idDatEdit])); d != "" {
				start = filepath.Dir(filepath.Clean(d))
			}
		}
		if p, ok := pickPath(a.hwnd, false, T("dlg.bkTitle"), start,
			[][2]string{{T("dlg.bkFilter"), "*.epidb"}, {T("dlg.allFilter"), "*.*"}}); ok {
			setText(a.ctl[idBkEdit], p)
			a.loadBackup(p)
		}
	case idRecursive:
		a.countDats(true)
	case idLang:
		if code == cbnDropDown || code == cbnCloseUp {
			pInvalidateRect.Call(a.ctl[id], 0, 0)
		}
		if code == cbnSelChange {
			if i := int(int32(send(a.ctl[idLang], cbGetCurSel, 0, 0))); i >= 0 && i < int(nLang) && Lang(i) != cur {
				cur = Lang(i)
				saveLang()
				a.applyLanguage(false)
			}
		}
	case idMachine:
		if code == cbnDropDown || code == cbnCloseUp {
			pInvalidateRect.Call(a.ctl[id], 0, 0)
		}
		if code == cbnSelChange {
			a.selectMachine(int(int32(send(a.ctl[idMachine], cbGetCurSel, 0, 0))))
		}
	case idConvert, 1: // 1 = IDOK (touche Entrée)
		a.convert()
	case idSort, idLaser:
		if code == cbnDropDown || code == cbnCloseUp {
			pInvalidateRect.Call(a.ctl[id], 0, 0)
		}
	case idOpen:
		a.openOut()
	case idHelp:
		a.showHelp()
	case idQuit: // comme la croix : sans effet pendant une conversion
		pPostMessageW.Call(a.hwnd, wmClose, 0, 0)
	case idDonate:
		pShellExecuteW.Call(a.hwnd, wp("open"), wp(donateURL), 0, 0, 1)
	case idSuggest:
		a.sendSuggestion()
	case idUpdate:
		if code == stnClicked && a.newRelease.page != "" {
			pShellExecuteW.Call(a.hwnd, wp("open"), wp(a.newRelease.page), 0, 0, 1)
		}
	}
}

// sendSuggestion : adresse copiée et explications ; s'il y a eu une erreur,
// l'explorateur s'ouvre aussi sur le rapport à joindre.
func (a *app) sendSuggestion() {
	report := ""
	if a.report != "" {
		if _, err := os.Stat(a.report); err == nil {
			report = a.report
			pShellExecuteW.Call(a.hwnd, wp("open"), wp("explorer.exe"), wp(`/select,"`+report+`"`), 0, 1)
		}
	}
	copyToClipboard(a.hwnd, suggestMail)
	msgBox(a.hwnd, suggestText(report), T("msg.title"), mbOK|mbIconInfo)
}

// afterShow : appelé une fois la fenêtre affichée (captures d'écran de contrôle, build « capture »).
var afterShow = func() {}

func wndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch uint32(msg) {
	case wmPaint:
		paintBuffered(hwnd, A.paint)
		return 0
	case wmEraseBkgnd:
		return 1
	case wmCtlColorStatic, wmCtlColorEdit, wmCtlColorLB:
		return A.onCtlColor(uint32(msg), wParam, lParam)
	case wmNotify:
		if r, ok := A.onNotify(lParam); ok {
			return r
		}
	case wmCommand:
		A.onCommand(int(wParam&0xFFFF), int((wParam>>16)&0xFFFF))
		return 0
	case wmLButtonDown:
		// clic dans le cadre d'un champ, à côté du texte : le champ prend le focus
		x, y := int32(int16(lParam&0xFFFF)), int32(int16(lParam>>16&0xFFFF))
		for id, f := range A.frames {
			if x >= f.l && x < f.r && y >= f.t && y < f.b {
				pSetFocus.Call(A.ctl[id])
				return 0
			}
		}
	case wmMeasureItem:
		if mi := (*measureItem)(unsafe.Pointer(lParam)); mi.ctlType == 3 { // ODT_COMBOBOX
			mi.itemHeight = uint32(A.s(24))
			return 1
		}
	case wmDrawItem:
		if di := (*drawItem)(unsafe.Pointer(lParam)); di.ctlType == 3 {
			A.drawComboItem(di)
			return 1
		}
	case wmSetCursor:
		if A.busy {
			wait, _, _ := pLoadCursorW.Call(0, 32514)
			pSetCursor.Call(wait)
			return 1
		}
		if wParam == A.ctl[idUpdate] {
			hand, _, _ := pLoadCursorW.Call(0, idcHand)
			pSetCursor.Call(hand)
			return 1
		}
	case wmUpdate:
		A.onUpdate()
		return 0
	case wmClose:
		if A.busy {
			return 0
		}
		pDestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		pPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return r
}

// fitScreen : mode normal, compact, ou échelle réduite, pour que la fenêtre tienne
// dans la zone de travail (1366 × 768 à 125 % ou 150 %…). Renvoie le cadre de fenêtre.
func (a *app) fitScreen(style uint32) (wa rect, frameW, frameH int32) {
	wa = workArea()
	fr := rect{0, 0, 100, 100}
	pAdjustWindowRectEx.Call(uintptr(unsafe.Pointer(&fr)), uintptr(style), 0, wsExControlPar)
	frameW, frameH = fr.r-fr.l-100, fr.b-fr.t-100
	availH := int(wa.b-wa.t-frameH) - 8
	availW := int(wa.r-wa.l-frameW) - 8
	sys := a.dpi
	switch {
	case a.s(normalM.clientHeight(normalM.logMin)) <= availH:
		a.compact = false
	case a.s(compactM.clientHeight(compactM.logMin)) <= availH:
		a.compact = true
	default:
		a.compact = true
		a.dpi = max(80, availH*96/compactM.clientHeight(compactM.logMin))
	}
	if a.s(cliW) > availW {
		a.dpi = max(80, min(a.dpi, availW*96/cliW))
	}
	if a.dpi > sys {
		a.dpi = sys
	}
	m := a.m()
	a.cliH = min(m.clientHeight(m.logDesign), availH*96/a.dpi)
	a.cliH = max(a.cliH, m.clientHeight(m.logMin))
	return
}

func main() {
	runtime.LockOSThread()
	pCoInitializeEx.Call(0, 2)
	icc := initCC{8, 0x0000FFFF}
	pInitCommonControlsEx.Call(uintptr(unsafe.Pointer(&icc)))
	initTheme()

	a := A
	a.hInst, _, _ = pGetModuleHandleW.Call(0)
	a.dpi = 96
	if pGetDpiForSystem.Find() == nil {
		if d, _, _ := pGetDpiForSystem.Call(); d > 0 {
			a.dpi = int(d)
		}
	} else {
		dc, _, _ := pGetDC.Call(0)
		d, _, _ := pGetDeviceCaps.Call(dc, 90)
		pReleaseDC.Call(0, dc)
		if d > 0 {
			a.dpi = int(d)
		}
	}
	a.realDpi = a.dpi
	if testDpi > 0 {
		a.dpi = testDpi
	}
	sysDpi := a.dpi
	style := uint32(wsOverlapped | wsCaption | wsSysMenu | wsMinimizeBox | wsClipChildren)
	wa, frameW, frameH := a.fitScreen(style)
	cur = initialLang()
	a.makeFonts()
	// icônes à la taille du système (barre des tâches), logo de secours à l'échelle de la fenêtre
	icoL, _, _ := pLoadImageW.Call(a.hInst, 1, 1, uintptr(32*sysDpi/96), uintptr(32*sysDpi/96), 0)
	icoS, _, _ := pLoadImageW.Call(a.hInst, 1, 1, uintptr(16*sysDpi/96), uintptr(16*sysDpi/96), 0)
	a.icoBig, _, _ = pLoadImageW.Call(a.hInst, 1, 1, uintptr(a.s(56)), uintptr(a.s(56)), 0)
	cur, _, _ := pLoadCursorW.Call(0, 32512)

	cls := w16("VAPConvertisseurEpilog")
	wc := wndClassEx{
		style:         0x3, // CS_HREDRAW|CS_VREDRAW
		lpfnWndProc:   syscall.NewCallback(wndProc),
		hInstance:     a.hInst,
		hIcon:         icoL,
		hCursor:       cur,
		hbrBackground: brFond,
		lpszClassName: cls,
		hIconSm:       icoS,
	}
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	ww, wh := int32(a.s(cliW))+frameW, int32(a.s(a.cliH))+frameH
	x := wa.l + (wa.r-wa.l-ww)/2
	y := wa.t + (wa.b-wa.t-wh)/2
	x, y = max(x, wa.l), max(y, wa.t)
	a.hwnd, _, _ = pCreateWindowExW.Call(wsExControlPar, uintptr(unsafe.Pointer(cls)), wp(T("app.title")),
		uintptr(style), uintptr(x), uintptr(y), uintptr(ww), uintptr(wh), 0, 0, a.hInst, 0)
	if a.hwnd == 0 {
		msgBox(0, T("err.window"), T("msg.title"), mbIconError)
		return
	}
	darkTitle(a.hwnd)
	a.buildUI()

	// dossier passé en argument (glisser-déposer sur l'icône)
	if len(os.Args) > 1 {
		if st, err := os.Stat(os.Args[1]); err == nil {
			if st.IsDir() {
				setText(a.ctl[idDatEdit], os.Args[1])
				a.countDats(true)
				a.fixDatDir()
			} else if strings.EqualFold(filepath.Ext(os.Args[1]), ".epidb") {
				setText(a.ctl[idBkEdit], os.Args[1])
				a.loadBackup(os.Args[1])
			}
		}
	}

	pShowWindow.Call(a.hwnd, 1)
	pUpdateWindow.Call(a.hwnd)
	a.checkUpdate()
	user32.NewProc("SetForegroundWindow").Call(a.hwnd)
	pSetFocus.Call(a.ctl[idDatBrowse])
	afterShow()

	var m msgT
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		target := a.hwnd
		if helpHwnd != 0 {
			target = helpHwnd
		}
		if d, _, _ := pIsDialogMessageW.Call(target, uintptr(unsafe.Pointer(&m))); d != 0 {
			continue
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

// ---------------------------------------------------------------------------
// Boîtes classiques (secours)

var (
	comdlg32              = syscall.NewLazyDLL("comdlg32.dll")
	pGetOpenFileNameW     = comdlg32.NewProc("GetOpenFileNameW")
	pSHBrowseForFolderW   = shell32.NewProc("SHBrowseForFolderW")
	pSHGetPathFromIDListW = shell32.NewProc("SHGetPathFromIDListW")
)

type openFileName struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        uintptr
	dwReserved        uint32
	flagsEx           uint32
}

type browseInfo struct {
	hwndOwner      uintptr
	pidlRoot       uintptr
	pszDisplayName *uint16
	lpszTitle      *uint16
	ulFlags        uint32
	lpfn           uintptr
	lParam         uintptr
	iImage         int32
}

func pickClassic(owner uintptr, folder bool, title, start string, filters [][2]string) (string, bool) {
	if folder {
		disp := make([]uint16, 260)
		bi := browseInfo{hwndOwner: owner, pszDisplayName: &disp[0], lpszTitle: w16(title), ulFlags: 0x1 | 0x40} // RETURNONLYFSDIRS | NEWDIALOGSTYLE
		pidl, _, _ := pSHBrowseForFolderW.Call(uintptr(unsafe.Pointer(&bi)))
		if pidl == 0 {
			return "", false
		}
		buf := make([]uint16, 1024)
		pSHGetPathFromIDListW.Call(pidl, uintptr(unsafe.Pointer(&buf[0])))
		pCoTaskMemFree.Call(pidl)
		s := syscall.UTF16ToString(buf)
		return s, s != ""
	}
	var f []uint16
	for _, x := range filters {
		f = append(f, syscall.StringToUTF16(x[0])...)
		f = append(f, syscall.StringToUTF16(x[1])...)
	}
	f = append(f, 0)
	buf := make([]uint16, 1024)
	ofn := openFileName{hwndOwner: owner, lpstrFilter: &f[0], nFilterIndex: 1, lpstrFile: &buf[0], nMaxFile: uint32(len(buf)),
		lpstrTitle: w16(title), flags: 0x00080000 | 0x00001000 | 0x00000800 | 0x00000008} // EXPLORER|FILEMUSTEXIST|PATHMUSTEXIST|NOCHANGEDIR
	if start != "" {
		d := start
		if st, err := os.Stat(start); err == nil && !st.IsDir() {
			d = filepath.Dir(start)
		}
		ofn.lpstrInitialDir = w16(d)
	}
	ofn.lStructSize = uint32(unsafe.Sizeof(ofn))
	r, _, _ := pGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	runtime.KeepAlive(f)
	if r == 0 {
		return "", false
	}
	s := syscall.UTF16ToString(buf)
	return s, s != ""
}

// ---------------------------------------------------------------------------
// Langue : choix mémorisé dans %APPDATA%\VAP Convertisseur\langue.txt,
// sinon langue de Windows (anglais si elle n'est pas proposée).

const appSignature = "VAP  V1.8"

func langFile() string {
	d, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(d, "VAP Convertisseur", "langue.txt")
}

func saveLang() {
	if p := langFile(); p != "" {
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(langCodes[cur]), 0o644)
	}
}

func initialLang() Lang {
	if p := langFile(); p != "" {
		if b, err := os.ReadFile(p); err == nil {
			if l, ok := langFromCode(strings.TrimSpace(string(b))); ok {
				return l
			}
		}
	}
	if pGetUserDefaultUILanguage.Find() == nil {
		id, _, _ := pGetUserDefaultUILanguage.Call()
		switch id & 0x3FF {
		case 0x0C:
			return LFR
		case 0x0A:
			return LES
		case 0x07:
			return LDE
		case 0x10:
			return LIT
		case 0x16:
			return LPT
		case 0x13:
			return LNL
		case 0x04:
			return LZH
		case 0x11:
			return LJA
		}
	}
	return LEN
}
