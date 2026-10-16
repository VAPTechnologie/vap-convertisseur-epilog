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

	mbOK          = 0x0
	mbYesNo       = 0x4
	mbIconError   = 0x10
	mbIconWarning = 0x30
	mbIconInfo    = 0x40
	idYes         = 6

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

func msgBox(owner uintptr, text, title string, flags uintptr) int {
	r, _, _ := pMessageBoxW.Call(owner, wp(text), wp(title), flags)
	return int(r)
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
	fontSmall                            uintptr
	brBody, brWhite, brHeader, brAccent  uintptr
	icoBig                               uintptr
	ctl                                  map[int]uintptr
	machines                             []Machine
	outDir                               string
	report                               string // dernier rapport d'erreur écrit
	busy                               bool
	sections                             []section
	titleRight, footerRight              int  // limites des textes dessinés (unités logiques)
	inWelcome, welcomeOnly               bool // journal = seulement le message d'accueil
}

type section struct {
	y     int
	num   string
	title string
}

var A = &app{ctl: map[int]uintptr{}}

const (
	cliW    = 860
	cliH    = 660
	headerH = 92
)

var (
	colBody    = rgb(244, 246, 250)
	colHeader  = rgb(14, 22, 36)
	colAccent  = rgb(30, 140, 255)
	colTitle   = rgb(255, 255, 255)
	colSub     = rgb(150, 190, 240)
	colSection = rgb(14, 40, 80)
	colText    = rgb(45, 52, 64)
	colMuted   = rgb(120, 128, 140)
	colLine    = rgb(214, 220, 230)
)

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
	a.child("BUTTON", "", bsPushButton|wsTabStop, 0, 0, 0, 10, 10, idHelp)
	// liste des langues dessinée par le programme, avec le drapeau du pays (flags_windows.go)
	a.child("COMBOBOX", "", cbsDropDownList|cbsOwnerDrawFixed|cbsHasStrings|wsVScroll|wsTabStop, 0, 0, 0, 10, 300, idLang)
	for i := Lang(0); i < nLang; i++ {
		send(a.ctl[idLang], cbAddString, 0, wp(langNames[i]))
	}
	a.child("EDIT", "", esAutoHScroll|wsTabStop, wsExClientEdge, 0, 0, 10, 10, idDatEdit)
	a.child("BUTTON", "", bsPushButton|wsTabStop, 0, 0, 0, 10, 10, idDatBrowse)
	a.child("BUTTON", "", bsAutoCheckbox|wsTabStop, 0, 0, 0, 10, 10, idRecursive)
	a.child("STATIC", "", ssRight|ssEndEllip|ssCenterImg, 0, 0, 0, 10, 10, idDatInfo)
	a.child("EDIT", "", esAutoHScroll|wsTabStop, wsExClientEdge, 0, 0, 10, 10, idBkEdit)
	a.child("BUTTON", "", bsPushButton|wsTabStop, 0, 0, 0, 10, 10, idBkBrowse)
	a.label(idLblMachine, ssLeft)
	a.child("COMBOBOX", "", cbsDropDownList|wsVScroll|wsTabStop, 0, 0, 0, 10, 300, idMachine)
	a.label(idLblFolder, ssLeft)
	a.child("EDIT", "", esAutoHScroll|wsTabStop, wsExClientEdge, 0, 0, 10, 10, idFolderName)
	a.child("BUTTON", "", bsAutoCheckbox|bsMultiline|wsTabStop, 0, 0, 0, 10, 10, idAir)
	a.label(idLblSort, ssLeft)
	a.child("COMBOBOX", "", cbsDropDownList|wsVScroll|wsTabStop, 0, 0, 0, 10, 200, idSort)
	a.child("BUTTON", "", bsAutoCheckbox|bsMultiline|wsTabStop, 0, 0, 0, 10, 10, idTop)
	a.label(idLblLaser, ssLeft)
	a.child("COMBOBOX", "", cbsDropDownList|wsVScroll|wsTabStop, 0, 0, 0, 10, 200, idLaser)
	a.child("msctls_progress32", "", 0, 0, 0, 0, 10, 10, idProgress)
	a.child("EDIT", "", esMultiline|esReadOnly|esAutoVScroll|wsVScroll, wsExClientEdge, 0, 0, 10, 10, idLog)
	a.child("BUTTON", "", bsPushButton|wsTabStop, 0, 0, 0, 10, 10, idOpen)
	a.child("BUTTON", "", bsDefPushButton|wsTabStop, 0, 0, 0, 10, 10, idConvert)
	a.child("BUTTON", "", bsPushButton|wsTabStop, 0, 0, 0, 10, 10, idQuit)
	a.child("BUTTON", "", bsPushButton|wsTabStop, 0, 0, 0, 10, 10, idDonate)
	a.child("BUTTON", "", bsPushButton|wsTabStop, 0, 0, 0, 10, 10, idSuggest)
	enable(a.ctl[idOpen], false)

	send(a.ctl[idAir], bmSetCheck, 1, 0)
	send(a.ctl[idTop], bmSetCheck, 1, 0)
	send(a.ctl[idProgress], pbmSetRange32, 0, 100)
	send(a.ctl[idProgress], pbmSetBarColor, 0, colAccent)

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
	a.layout()
	if first || a.welcomeOnly {
		a.showWelcome()
	}
	pInvalidateRect.Call(a.hwnd, 0, 1)
}

func (a *app) showWelcome() {
	setText(a.ctl[idLog], "")
	a.inWelcome = true
	for i := 0; i <= 5; i++ {
		a.log(T(fmt.Sprintf("welcome.%d", i)))
	}
	a.inWelcome = false
	a.welcomeOnly = true
	send(a.ctl[idLog], 0x0115, 6, 0) // WM_VSCROLL SB_TOP
}

// makeFonts : polices adaptées à la langue (Segoe UI n'a pas les caractères chinois / japonais).
func (a *app) makeFonts() {
	old := []uintptr{a.fontUI, a.fontSmall, a.fontBold, a.fontTitle, a.fontSub}
	face, bold, bw := "Segoe UI", "Segoe UI Semibold", 600
	switch cur {
	case LZH:
		face, bold, bw = "Microsoft YaHei UI", "Microsoft YaHei UI", 700
	case LJA:
		face, bold, bw = "Yu Gothic UI", "Yu Gothic UI Semibold", 600
	}
	a.fontUI = a.font(10, 400, face)
	a.fontSmall = a.font(9, 400, face)
	a.fontBold = a.font(10, bw, bold)
	a.fontTitle = a.font(18, bw, bold)
	a.fontSub = a.font(10, 400, face)
	for _, h := range a.ctl {
		send(h, wmSetFont, a.fontUI, 1)
	}
	send(a.ctl[idLog], wmSetFont, a.fontSmall, 1)
	send(a.ctl[idConvert], wmSetFont, a.fontBold, 1)
	del := gdi32.NewProc("DeleteObject")
	for _, f := range old {
		if f != 0 {
			del.Call(f)
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

// layout : place les contrôles en fonction de la longueur des textes de la langue.
func (a *app) layout() {
	m := 26
	right := cliW - m
	mv := func(id, x, y, w, h int) {
		pMoveWindow.Call(a.ctl[id], uintptr(a.s(x)), uintptr(a.s(y)), uintptr(a.s(w)), uintptr(a.s(h)), 1)
	}
	tw := func(k string) int { return a.measure(a.fontUI, T(k)) }

	// en-tête : notice + langue
	hw := max(150, tw("btn.help")+36)
	mv(idHelp, right-hw, 12, hw, 32)
	mv(idLang, right-hw, 52, hw, 300)
	a.titleRight = right - hw - 16

	a.sections = a.sections[:0]
	y := headerH + 22

	a.sections = append(a.sections, section{y, "1", T("sec1")})
	y += 30
	bw := max(118, tw("btn.browse")+32)
	mv(idDatEdit, m, y, right-m-bw-10, 27)
	mv(idDatBrowse, right-bw, y-1, bw, 29)
	y += 34
	cw := tw("chk.recursive") + 30
	mv(idRecursive, m+2, y, cw, 22)
	mv(idDatInfo, m+2+cw+10, y, right-(m+2+cw+10), 22)
	y += 40

	a.sections = append(a.sections, section{y, "2", T("sec2")})
	y += 30
	mv(idBkEdit, m, y, right-m-bw-10, 27)
	mv(idBkBrowse, right-bw, y-1, bw, 29)
	y += 46

	a.sections = append(a.sections, section{y, "3", T("sec3")})
	y += 30
	w1 := tw("lbl.machine") + 10
	w2 := tw("lbl.folder") + 10
	comboW := 270
	if free := (right - m) - w1 - 18 - w2 - 160; free < comboW {
		comboW = max(free, 160)
	}
	mv(idLblMachine, m, y, w1, 26)
	mv(idMachine, m+w1, y, comboW, 300)
	x := m + w1 + comboW + 18
	mv(idLblFolder, x, y, w2, 26)
	mv(idFolderName, x+w2, y, right-x-w2, 27)
	y += 46

	a.sections = append(a.sections, section{y, "4", T("sec4")})
	y += 30
	cbw := 150
	for _, k := range []string{"sort.none", "sort.inout", "sort.opt", "sort.dat", "laser.co2", "laser.fiber"} {
		cbw = max(cbw, tw(k)+40)
	}
	lw := max(tw("lbl.sort"), tw("lbl.laser")) + 12
	gx := right - cbw - lw
	chkW := gx - 12 - (m + 2)
	mv(idAir, m+2, y-3, chkW, 30)
	mv(idLblSort, gx, y, lw, 26)
	mv(idSort, gx+lw, y, cbw, 200)
	y += 34
	mv(idTop, m+2, y-3, chkW, 30)
	mv(idLblLaser, gx, y, lw, 26)
	mv(idLaser, gx+lw, y, cbw, 200)
	y += 42

	mv(idProgress, m, y, right-m, 10)
	y += 18
	mv(idLog, m, y, right-m, cliH-y-66)

	by := cliH - 50
	qw := max(110, tw("btn.quit")+36)
	cvw := max(150, a.measure(a.fontBold, T("btn.convert"))+36)
	ow := max(150, tw("btn.open")+36)
	mv(idQuit, right-qw, by, qw, 34)
	mv(idConvert, right-qw-12-cvw, by, cvw, 34)
	mv(idOpen, right-qw-12-cvw-12-ow, by, ow, 34)
	// bouton de don juste après « VAP V1.4 », en bas à gauche
	sigW := a.measure(a.fontSmall, appSignature)
	dw := max(110, tw("btn.donate")+32)
	mv(idDonate, m+sigW+14, by, dw, 34)
	sw := max(110, tw("btn.suggest")+32)
	mv(idSuggest, m+sigW+14+dw+8, by, sw, 34)
	a.footerRight = m + sigW + 4
}

func (a *app) log(s string) {
	h := a.ctl[idLog]
	n, _, _ := pGetWindowTextLengthW.Call(h)
	send(h, emSetSel, n, n)
	prefix := "\r\n"
	if n == 0 {
		prefix = ""
	}
	send(h, emReplaceSel, 0, wp(prefix+s))
	pUpdateWindow.Call(h)
	if !a.inWelcome {
		a.welcomeOnly = false
	}
}

func (a *app) checked(id int) bool { return send(a.ctl[id], bmGetCheck, 0, 0) == 1 }

func (a *app) paint(hdc uintptr) {
	var rc rect
	pGetClientRect.Call(a.hwnd, uintptr(unsafe.Pointer(&rc)))
	hdr := rect{0, 0, rc.r, int32(a.s(headerH))}
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&hdr)), a.brHeader)
	acc := rect{0, int32(a.s(headerH)), rc.r, int32(a.s(headerH + 4))}
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&acc)), a.brAccent)
	// logo VAP + trait, puis le titre (repli sur l'icône si le logo ne peut pas être dessiné)
	tx := a.logoTitleX(hdc, 24, 46, 27)
	if tx == 24 && a.icoBig != 0 {
		pDrawIconEx.Call(hdc, uintptr(a.s(24)), uintptr(a.s(16)), a.icoBig, uintptr(a.s(60)), uintptr(a.s(60)), 0, 0, 3)
		tx = 100
	}
	pSetBkMode.Call(hdc, 1)
	old, _, _ := pSelectObject.Call(hdc, a.fontTitle)
	pSetTextColor.Call(hdc, colTitle)
	t := rect{int32(a.s(tx)), int32(a.s(16)), int32(a.s(a.titleRight)), int32(a.s(52))}
	pDrawTextW.Call(hdc, wp(T("hdr.title")), ^uintptr(0), uintptr(unsafe.Pointer(&t)), dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllip)
	pSelectObject.Call(hdc, a.fontSub)
	pSetTextColor.Call(hdc, colSub)
	t = rect{int32(a.s(tx + 1)), int32(a.s(52)), int32(a.s(a.titleRight)), int32(a.s(76))}
	pDrawTextW.Call(hdc, wp(T("hdr.sub")), ^uintptr(0), uintptr(unsafe.Pointer(&t)), dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllip)

	// sections
	nullPen, _, _ := pGetStockObject.Call(8)
	linePen, _, _ := pCreatePen.Call(0, uintptr(a.s(1)), colLine)
	for _, s := range a.sections {
		y := a.s(s.y)
		pSelectObject.Call(hdc, nullPen)
		ob, _, _ := pSelectObject.Call(hdc, a.brAccent)
		d := a.s(22)
		pEllipse.Call(hdc, uintptr(a.s(26)), uintptr(y), uintptr(a.s(26)+d), uintptr(y+d))
		pSelectObject.Call(hdc, ob)
		pSelectObject.Call(hdc, a.fontBold)
		pSetTextColor.Call(hdc, colTitle)
		c := rect{int32(a.s(26)), int32(y), int32(a.s(26) + d), int32(y + d)}
		pDrawTextW.Call(hdc, wp(s.num), ^uintptr(0), uintptr(unsafe.Pointer(&c)), 0x1|dtVCenter|dtSingleLine)
		pSetTextColor.Call(hdc, colSection)
		tr := rect{int32(a.s(58)), int32(y), int32(a.s(cliW - 26)), int32(y + d)}
		pDrawTextW.Call(hdc, wp(s.title), ^uintptr(0), uintptr(unsafe.Pointer(&tr)), dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
		// filet après le titre
		pSelectObject.Call(hdc, linePen)
		tw := a.s(66) + a.textWidth(hdc, s.title)
		pMoveToEx.Call(hdc, uintptr(tw), uintptr(y+d/2), 0)
		pLineTo.Call(hdc, uintptr(a.s(cliW-26)), uintptr(y+d/2))
	}
	pSelectObject.Call(hdc, nullPen)
	syscall.SyscallN(gdi32.NewProc("DeleteObject").Addr(), linePen)

	// pied de page
	pSelectObject.Call(hdc, a.fontSmall)
	pSetTextColor.Call(hdc, colMuted)
	f := rect{int32(a.s(26)), int32(a.s(cliH - 50)), int32(a.s(a.footerRight)), int32(a.s(cliH - 16))}
	pDrawTextW.Call(hdc, wp(appSignature), ^uintptr(0), uintptr(unsafe.Pointer(&f)), dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllip)
	pSelectObject.Call(hdc, old)
}

func (a *app) textWidth(hdc uintptr, s string) int {
	r := rect{0, 0, 2000, 100}
	pDrawTextW.Call(hdc, wp(s), ^uintptr(0), uintptr(unsafe.Pointer(&r)), dtLeft|dtSingleLine|0x400) // DT_CALCRECT
	return int(r.r)
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
	if autoLaser && strings.Contains(foldKey(dir), "fibre") || strings.Contains(foldKey(dir), "fiber") {
		send(a.ctl[idLaser], cbSetCurSel, 1, 0)
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
	if m.HasFiber() && !m.HasCO2() {
		send(a.ctl[idLaser], cbSetCurSel, 1, 0)
	}
}

func (a *app) loadBackup(path string) {
	send(a.ctl[idMachine], cbResetContent, 0, 0)
	a.machines = nil
	db, err := openRO(path)
	if err != nil {
		msgBox(a.hwnd, T("err.openBackup")+err.Error(), T("msg.title"), mbIconError)
		return
	}
	ms, err := LoadMachines(db)
	db.Close()
	if err != nil {
		msgBox(a.hwnd, T("err.notBackup")+err.Error(), T("msg.title"), mbIconError)
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
		AirAssistCut: a.checked(idAir),
		VectorSort:   int(int32(send(a.ctl[idSort], cbGetCurSel, 0, 0))),
		OnTop:        a.checked(idTop),
		Fiber:        send(a.ctl[idLaser], cbGetCurSel, 0, 0) == 1,
	}
	if o.VectorSort < 0 {
		o.VectorSort = 1
	}
	// dossier déjà présent ?
	if db, err := openRO(bk); err == nil {
		fid, _ := FindJobFolder(db, folder)
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
	for _, id := range []int{idConvert, idDatBrowse, idBkBrowse, idOpen} {
		enable(a.ctl[id], false)
	}
	wait, _, _ := pLoadCursorW.Call(0, 32514)
	pSetCursor.Call(wait)
	send(a.ctl[idProgress], pbmSetPos, 0, 0)
	a.log("")
	a.log(Tf("log.convStart", folder, o.Machine.Name))
	out, res, err := Run(openRW, bk, o, a.log, func(d, t int) {
		send(a.ctl[idProgress], pbmSetPos, uintptr(d*100/t), 0)
		pumpMessages()
	})
	a.busy = false
	for _, id := range []int{idConvert, idDatBrowse, idBkBrowse} {
		enable(a.ctl[id], true)
	}
	arrow, _, _ := pLoadCursorW.Call(0, 32512)
	pSetCursor.Call(arrow)
	if err != nil {
		send(a.ctl[idProgress], pbmSetPos, 0, 0)
		a.log(T("log.fail") + err.Error())
		rp, rerr := WriteReport(openRO, ReportInput{Version: appVersion(), System: reportSystem(),
			Backup: bk, Options: o, Log: getText(a.ctl[idLog]), Err: err})
		if rerr != nil {
			msgBox(a.hwnd, T("err.convFailed")+err.Error(), T("msg.title"), mbIconError)
			return
		}
		a.report = rp
		a.log(T("log.report") + rp)
		if msgBox(a.hwnd, T("err.convFailed")+err.Error()+Tf("ask.sendReport", rp), T("msg.title"), mbYesNo|mbIconError) == idYes {
			a.sendSuggestion()
		}
		return
	}
	send(a.ctl[idProgress], pbmSetPos, 100, 0)
	a.outDir = out
	enable(a.ctl[idOpen], true)
	a.log(Tf("log.done", len(res.Converted), len(res.Skipped)))
	a.log(T("log.outDir") + out)
	if msgBox(a.hwnd, Tf("msg.done", len(res.Converted), folder), T("msg.title"), mbYesNo|mbIconInfo) == idYes {
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
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func (a *app) onCommand(id int, code int) {
	switch id {
	case idDatBrowse:
		start := strings.TrimSpace(getText(a.ctl[idDatEdit]))
		if p, ok := pickPath(a.hwnd, true, T("dlg.datTitle"), start, nil); ok {
			setText(a.ctl[idDatEdit], p)
			a.countDats(true)
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
		if code == cbnSelChange {
			if i := int(int32(send(a.ctl[idLang], cbGetCurSel, 0, 0))); i >= 0 && i < int(nLang) && Lang(i) != cur {
				cur = Lang(i)
				saveLang()
				a.applyLanguage(false)
			}
		}
	case idMachine:
		if code == cbnSelChange {
			a.selectMachine(int(int32(send(a.ctl[idMachine], cbGetCurSel, 0, 0))))
		}
	case idConvert, 1: // 1 = IDOK (touche Entrée)
		a.convert()
	case idOpen:
		a.openOut()
	case idHelp:
		a.showHelp()
	case idQuit:
		pDestroyWindow.Call(a.hwnd)
	case idDonate:
		pShellExecuteW.Call(a.hwnd, wp("open"), wp(donateURL), 0, 0, 1)
	case idSuggest:
		a.sendSuggestion()
	}
}

// sendSuggestion : e-mail préparé ; s'il y a eu une erreur, l'explorateur
// s'ouvre aussi sur le rapport pour le joindre (mailto ne permet pas de pièce jointe).
func (a *app) sendSuggestion() {
	report := ""
	if a.report != "" {
		if _, err := os.Stat(a.report); err == nil {
			report = a.report
			pShellExecuteW.Call(a.hwnd, wp("open"), wp("explorer.exe"), wp(`/select,"`+report+`"`), 0, 1)
		}
	}
	pShellExecuteW.Call(a.hwnd, wp("open"), wp(suggestMailto(report)), 0, 0, 1)
}

func wndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch uint32(msg) {
	case wmPaint:
		var ps paintStruct
		hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		A.paint(hdc)
		pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0
	case wmCtlColorStatic:
		hdc := wParam
		if lParam == A.ctl[idLog] {
			pSetBkColor.Call(hdc, rgb(255, 255, 255))
			pSetTextColor.Call(hdc, colText)
			return A.brWhite
		}
		pSetBkColor.Call(hdc, colBody)
		if lParam == A.ctl[idDatInfo] {
			pSetTextColor.Call(hdc, colMuted)
		} else {
			pSetTextColor.Call(hdc, colText)
		}
		return A.brBody
	case wmCommand:
		A.onCommand(int(wParam&0xFFFF), int((wParam>>16)&0xFFFF))
		return 0
	case wmMeasureItem:
		if mi := (*measureItem)(unsafe.Pointer(lParam)); mi.ctlID == idLang {
			mi.itemHeight = uint32(A.s(20))
			return 1
		}
	case wmDrawItem:
		if di := (*drawItem)(unsafe.Pointer(lParam)); di.ctlID == idLang {
			A.drawLangItem(di)
			return 1
		}
	case wmSetCursor:
		if A.busy {
			wait, _, _ := pLoadCursorW.Call(0, 32514)
			pSetCursor.Call(wait)
			return 1
		}
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

func main() {
	runtime.LockOSThread()
	pCoInitializeEx.Call(0, 2)
	icc := initCC{8, 0x0000FFFF}
	pInitCommonControlsEx.Call(uintptr(unsafe.Pointer(&icc)))

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
	cur = initialLang()
	a.makeFonts()
	a.brBody, _, _ = pCreateSolidBrush.Call(colBody)
	a.brWhite, _, _ = pCreateSolidBrush.Call(rgb(255, 255, 255))
	a.brHeader, _, _ = pCreateSolidBrush.Call(colHeader)
	a.brAccent, _, _ = pCreateSolidBrush.Call(colAccent)
	icoL, _, _ := pLoadImageW.Call(a.hInst, 1, 1, uintptr(a.s(32)), uintptr(a.s(32)), 0)
	icoS, _, _ := pLoadImageW.Call(a.hInst, 1, 1, uintptr(a.s(16)), uintptr(a.s(16)), 0)
	a.icoBig, _, _ = pLoadImageW.Call(a.hInst, 1, 1, uintptr(a.s(60)), uintptr(a.s(60)), 0)
	cur, _, _ := pLoadCursorW.Call(0, 32512)

	cls := w16("VAPConvertisseurEpilog")
	wc := wndClassEx{
		style:         0x3, // CS_HREDRAW|CS_VREDRAW
		lpfnWndProc:   syscall.NewCallback(wndProc),
		hInstance:     a.hInst,
		hIcon:         icoL,
		hCursor:       cur,
		hbrBackground: a.brBody,
		lpszClassName: cls,
		hIconSm:       icoS,
	}
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	style := uint32(wsOverlapped | wsCaption | wsSysMenu | wsMinimizeBox | wsClipChildren)
	rc := rect{0, 0, int32(a.s(cliW)), int32(a.s(cliH))}
	pAdjustWindowRectEx.Call(uintptr(unsafe.Pointer(&rc)), uintptr(style), 0, wsExControlPar)
	ww, wh := rc.r-rc.l, rc.b-rc.t
	sw, _, _ := pGetSystemMetrics.Call(0)
	sh, _, _ := pGetSystemMetrics.Call(1)
	x, y := (int32(sw)-ww)/2, (int32(sh)-wh)/2
	if y < 0 {
		y = 0
	}
	a.hwnd, _, _ = pCreateWindowExW.Call(wsExControlPar, uintptr(unsafe.Pointer(cls)), wp(T("app.title")),
		uintptr(style), uintptr(x), uintptr(y), uintptr(ww), uintptr(wh), 0, 0, a.hInst, 0)
	if a.hwnd == 0 {
		msgBox(0, T("err.window"), T("msg.title"), mbIconError)
		return
	}
	a.buildUI()

	// dossier passé en argument (glisser-déposer sur l'icône)
	if len(os.Args) > 1 {
		if st, err := os.Stat(os.Args[1]); err == nil {
			if st.IsDir() {
				setText(a.ctl[idDatEdit], os.Args[1])
				a.countDats(true)
			} else if strings.EqualFold(filepath.Ext(os.Args[1]), ".epidb") {
				setText(a.ctl[idBkEdit], os.Args[1])
				a.loadBackup(os.Args[1])
			}
		}
	}

	pShowWindow.Call(a.hwnd, 1)
	pUpdateWindow.Call(a.hwnd)
	user32.NewProc("SetForegroundWindow").Call(a.hwnd)
	pSetFocus.Call(a.ctl[idDatBrowse])

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

const appSignature = "VAP  V1.4"

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
