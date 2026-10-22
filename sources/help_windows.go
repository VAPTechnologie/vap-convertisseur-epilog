//go:build windows

package main

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

var (
	helpHwnd  uintptr
	helpEdit  uintptr
	helpClose uintptr
	helpClass bool
	helpCB    = syscall.NewCallback(helpProc)
	helpSize  struct{ w, h, hdr, foot int } // unités logiques
)

func helpProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	a := A
	switch uint32(msg) {
	case wmPaint:
		paintBuffered(hwnd, func(hdc uintptr) {
			var rc rect
			pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
			fillBox(hdc, 0, 0, rc.r, rc.b, colFond)
			hs := helpSize
			a.paintHeader(hdc, rc.r, hs.hdr, 24, T("help.title"), "", hs.w-20)
			// panneau du texte
			roundBox(hdc, a.px(rect{20, int32(hs.hdr + 3 + 14), int32(hs.w - 20), int32(hs.h - hs.foot - 12)}), int32(a.s(12)), colPanneau, colFilet)
			// pied de page
			fy := int32(a.s(hs.h - hs.foot))
			fillBox(hdc, 0, fy, rc.r, rc.b, colPanneau)
			fillBox(hdc, 0, fy, rc.r, fy+int32(max(1, a.s(1))), colFilet)
			drawText(hdc, appSignature, rect{int32(a.s(20)), fy, int32(a.s(hs.w - 160)), rc.b}, a.fontSmall, colDiscret, dtLeft|dtVCenter|dtSingleLine)
		})
		return 0
	case wmEraseBkgnd:
		return 1
	case wmCtlColorStatic: // notice sans RichEdit (champ texte simple)
		pSetBkColor.Call(wParam, colPanneau)
		pSetTextColor.Call(wParam, colTexte)
		return brPanneau
	case wmNotify:
		if nm := (*nmCustomDrawT)(unsafe.Pointer(lParam)); nm.hdr.code == nmCustomDraw && nm.hdr.hwndFrom == helpClose {
			if nm.drawStage != cddsPrepaint {
				return cdrfDoDefault
			}
			a.drawButton(nm, false, colPanneau)
			return cdrfSkipDef
		}
	case wmCommand:
		if int(wParam&0xFFFF) == idHelpClose || int(wParam&0xFFFF) == 2 {
			pDestroyWindow.Call(hwnd)
		}
		return 0
	case wmClose:
		pDestroyWindow.Call(hwnd)
		return 0
	case wmDestroy:
		enable(a.hwnd, true)
		user32.NewProc("SetForegroundWindow").Call(a.hwnd)
		helpHwnd = 0
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return r
}

func (a *app) showHelp() {
	if helpHwnd != 0 {
		user32.NewProc("SetForegroundWindow").Call(helpHwnd)
		return
	}
	cls := w16("VAPNotice")
	if !helpClass {
		wc := wndClassEx{
			style:         0x3,
			lpfnWndProc:   helpCB,
			hInstance:     a.hInst,
			hbrBackground: brFond,
			lpszClassName: cls,
		}
		wc.hCursor, _, _ = pLoadCursorW.Call(0, 32512)
		wc.hIcon, _, _ = pLoadImageW.Call(a.hInst, 1, 1, uintptr(a.s(32)), uintptr(a.s(32)), 0)
		wc.hIconSm, _, _ = pLoadImageW.Call(a.hInst, 1, 1, uintptr(a.s(16)), uintptr(a.s(16)), 0)
		wc.cbSize = uint32(unsafe.Sizeof(wc))
		pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
		helpClass = true
	}
	// taille : 780 × 720 au plus, toujours dans la zone de travail
	style := uint32(wsOverlapped | wsCaption | wsSysMenu | wsClipChildren)
	wa := workArea()
	fr := rect{0, 0, 100, 100}
	pAdjustWindowRectEx.Call(uintptr(unsafe.Pointer(&fr)), uintptr(style), 0, wsExControlPar)
	m := a.m()
	helpSize.hdr, helpSize.foot = m.hdr-6, m.foot
	helpSize.w = min(780, int(wa.r-wa.l-(fr.r-fr.l-100)-16)*96/a.dpi)
	helpSize.h = min(720, int(wa.b-wa.t-(fr.b-fr.t-100)-8)*96/a.dpi)
	hs := helpSize
	rc := rect{0, 0, int32(a.s(hs.w)), int32(a.s(hs.h))}
	pAdjustWindowRectEx.Call(uintptr(unsafe.Pointer(&rc)), uintptr(style), 0, wsExControlPar)
	ww, wh := rc.r-rc.l, rc.b-rc.t
	x, y := wa.l+(wa.r-wa.l-ww)/2, wa.t+(wa.b-wa.t-wh)/2
	x, y = max(x, wa.l), max(y, wa.t)
	helpHwnd, _, _ = pCreateWindowExW.Call(wsExControlPar, uintptr(unsafe.Pointer(cls)), wp(T("help.caption")),
		uintptr(style), uintptr(x), uintptr(y), uintptr(ww), uintptr(wh), a.hwnd, 0, a.hInst, 0)
	if helpHwnd == 0 {
		return
	}
	darkTitle(helpHwnd)
	mk := func(class, text string, st, ex uint32, x, y, w, h, id int) uintptr {
		hw, _, _ := pCreateWindowExW.Call(uintptr(ex), wp(class), wp(text), uintptr(st|wsChild|wsVisible),
			uintptr(a.s(x)), uintptr(a.s(y)), uintptr(a.s(w)), uintptr(a.s(h)), helpHwnd, uintptr(id), a.hInst, 0)
		send(hw, wmSetFont, a.fontUI, 1)
		return hw
	}
	helpEdit = 0
	tx, ty := 20+14, hs.hdr+3+14+10
	tw, th := hs.w-40-20, hs.h-hs.foot-12-10-ty
	if richOK() {
		helpEdit = mk(richEditClass, "", esMultiline|esReadOnly|esAutoVScroll|wsVScroll, 0, tx, ty, tw, th, idHelpText)
	}
	if helpEdit != 0 {
		send(helpEdit, emSetBkgndColor, 0, colPanneau)
		rtf := noticeRTF(noticeText())
		st := struct{ flags, codepage uint32 }{0, 0} // ST_DEFAULT, texte RTF en ASCII
		send(helpEdit, emSetTextEx, uintptr(unsafe.Pointer(&st)), uintptr(unsafe.Pointer(&append([]byte(rtf), 0)[0])))
	} else {
		helpEdit = mk("EDIT", strings.ReplaceAll(noticeText(), "\n", "\r\n"), esMultiline|esReadOnly|esAutoVScroll|wsVScroll, 0,
			tx, ty, tw, th, idHelpText)
	}
	darkScrollbars(helpEdit)
	m2 := a.m()
	bw := max(120, a.measure(a.fontUI, T("btn.close"))+36)
	helpClose = mk("BUTTON", T("btn.close"), bsDefPushButton|wsTabStop, 0, hs.w-20-bw, hs.h-hs.foot+(hs.foot-m2.btn)/2, bw, m2.btn, idHelpClose)
	enable(a.hwnd, false)
	pShowWindow.Call(helpHwnd, 1)
	pUpdateWindow.Call(helpHwnd)
	pSetFocus.Call(helpClose)
}

// noticeRTF : notice mise en forme sans changer le texte. Titres (ligne non indentée
// après une ligne vide) en gras bleu clair ; premier titre (sauvegarde obligatoire)
// en orange ; corps en blanc avec un interligne aéré.
func noticeRTF(text string) string {
	face, bold, _ := uiFaces()
	if cur != LZH && cur != LJA {
		bold = face
	}
	var b strings.Builder
	fmt.Fprintf(&b, `{\rtf1\ansi\deff0{\fonttbl{\f0\fnil\fcharset1 %s;}{\f1\fnil\fcharset1 %s;}}`, face, bold)
	b.WriteString(`{\colortbl ;`)
	for _, c := range []uintptr{colTexte, colAccentClair, colOrange, colDiscret} {
		fmt.Fprintf(&b, `\red%d\green%d\blue%d;`, c&0xFF, c>>8&0xFF, c>>16&0xFF)
	}
	b.WriteString(`}\uc1`)
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	first := true
	for i, l := range lines {
		heading := l != "" && l[0] != ' ' && (i == 0 || lines[i-1] == "")
		switch {
		case l == "":
			b.WriteString(`\pard\sa0\sl120\slmult0\fs10\par` + "\n")
		case heading:
			col := 2
			if first {
				col = 3
			}
			first = false
			fmt.Fprintf(&b, `\pard\sb160\sa100\f1\b\fs24\cf%d `, col)
			rtfEscape(&b, l)
			b.WriteString(`\b0\par` + "\n")
		default:
			b.WriteString(`\pard\sa0\sl264\slmult1\f0\fs20\cf1 `)
			rtfEscape(&b, l)
			b.WriteString(`\par` + "\n")
		}
	}
	b.WriteString("}")
	return b.String()
}

// rtfEscape : texte en RTF ASCII (caractères accentués, chinois et japonais en \uN?).
func rtfEscape(b *strings.Builder, s string) {
	for _, u := range toUTF16(s) {
		switch {
		case u == '\\' || u == '{' || u == '}':
			b.WriteByte('\\')
			b.WriteByte(byte(u))
		case u == '\t':
			b.WriteString(`\tab `)
		case u < 0x80:
			b.WriteByte(byte(u))
		default:
			fmt.Fprintf(b, `\u%d?`, int16(u))
		}
	}
}
