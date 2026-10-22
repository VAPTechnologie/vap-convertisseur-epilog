//go:build windows

package main

// Boîtes de message sombres : même appel et mêmes réponses que MessageBoxW
// (IDOK, IDYES, IDNO), dans une vraie boîte de dialogue modale Windows
// (DialogBoxIndirectParamW : clavier, Entrée, Échap et fenêtre propriétaire
// gérés par Windows). Icône dessinée, texte sélectionnable et copiable,
// défilement si le texte est long. En cas d'échec : MessageBoxW standard.

import (
	"strings"
	"syscall"
	"unsafe"
)

var (
	pDialogBoxIndirectParamW = user32.NewProc("DialogBoxIndirectParamW")
	pEndDialog               = user32.NewProc("EndDialog")
	pSetWindowPos            = user32.NewProc("SetWindowPos")
	pGetDlgItem              = user32.NewProc("GetDlgItem")
	dlgCB                    = syscall.NewCallback(msgDlgProc)
)

const (
	wmInitDialog = 0x0110
	idOK         = 1
	idCancel     = 2
	idNo         = 7
	idMsgText    = 200
)

// état de la boîte affichée (une seule à la fois : elle est modale)
var mb struct {
	text, title string
	flags       uintptr
	iconR       rect // pixels
	footY       int32
}

func msgBox(owner uintptr, text, title string, flags uintptr) int {
	if owner == 0 {
		return nativeBox(owner, text, title, flags)
	}
	mb.text, mb.title, mb.flags = text, title, flags
	// modèle minimal : WS_POPUP | WS_CAPTION | WS_SYSMENU | DS_MODALFRAME, aucun contrôle (créés à l'ouverture)
	tpl := make([]uint16, 16)
	style := uint32(0x80000000 | 0x00C00000 | 0x00080000 | 0x80)
	tpl[0], tpl[1] = uint16(style), uint16(style>>16)
	r, _, _ := pDialogBoxIndirectParamW.Call(A.hInst, uintptr(unsafe.Pointer(&tpl[0])), owner, dlgCB, 0)
	if int(r) <= 0 {
		return nativeBox(owner, text, title, flags)
	}
	return int(r)
}

func nativeBox(owner uintptr, text, title string, flags uintptr) int {
	r, _, _ := pMessageBoxW.Call(owner, wp(text), wp(title), flags)
	return int(r)
}

func msgDlgProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	a := A
	switch uint32(msg) {
	case wmInitDialog:
		a.initMsgDlg(hwnd)
		return 0 // focus déjà placé
	case wmCtlColorDlg:
		return brFond
	case wmCtlColorStatic:
		pSetBkColor.Call(wParam, colFond)
		pSetTextColor.Call(wParam, colTexte)
		return brFond
	case wmPaint:
		var ps paintStruct
		hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		a.paintMsgDlg(hwnd, hdc)
		pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 1
	case wmNotify:
		if nm := (*nmCustomDrawT)(unsafe.Pointer(lParam)); nm.hdr.code == nmCustomDraw {
			if nm.drawStage != cddsPrepaint {
				return 0
			}
			a.drawButton(nm, nm.hdr.idFrom == uintptr(mbDefault()), colPanneau)
			setDlgResult(hwnd, cdrfSkipDef)
			return 1
		}
	case wmCommand:
		switch id := int(wParam & 0xFFFF); id {
		case idOK, idYes, idNo:
			// seulement les boutons présents (pas d'IDOK implicite dans une question Oui / Non)
			if h, _, _ := pGetDlgItem.Call(hwnd, uintptr(id)); h != 0 {
				pEndDialog.Call(hwnd, uintptr(id))
			}
			return 1
		case idCancel: // Échap ou croix : la réponse prudente (Non), ou OK
			if mb.flags&0xF == mbYesNo {
				pEndDialog.Call(hwnd, idNo)
			} else {
				pEndDialog.Call(hwnd, idOK)
			}
			return 1
		}
	}
	return 0
}

var pSetWindowLongPtrW = user32.NewProc("SetWindowLongPtrW")

// setDlgResult : valeur renvoyée par une procédure de dialogue pour WM_NOTIFY (DWLP_MSGRESULT).
func setDlgResult(hwnd, v uintptr) { pSetWindowLongPtrW.Call(hwnd, 0, v) }

// mbDefault : bouton par défaut (premier bouton, comme MessageBoxW).
func mbDefault() int {
	if mb.flags&0xF == mbYesNo {
		return idYes
	}
	return idOK
}

func (a *app) initMsgDlg(hwnd uintptr) {
	setText(hwnd, mb.title)
	darkTitle(hwnd)
	text := strings.ReplaceAll(strings.ReplaceAll(mb.text, "\r\n", "\n"), "\n", "\r\n")

	pad, icon, gap := int32(a.s(22)), int32(a.s(36)), int32(a.s(16))
	textW := int32(a.s(440))
	// hauteur du texte, coupé aux mots ; au-delà de 55 % de l'écran, il défile
	dc, _, _ := pGetDC.Call(hwnd)
	of, _, _ := pSelectObject.Call(dc, a.fontUI)
	calc := rect{0, 0, textW, 10000}
	pDrawTextW.Call(dc, wp(text), ^uintptr(0), uintptr(unsafe.Pointer(&calc)), dtWordBreak|dtEditControl|dtNoPrefix|0x40|dtCalcRect) // DT_EXPANDTABS
	pSelectObject.Call(dc, of)
	pReleaseDC.Call(hwnd, dc)
	wa := workArea()
	textH := calc.b + int32(a.s(4))
	maxH := (wa.b - wa.t) * 55 / 100
	scroll := uint32(0)
	if textH > maxH {
		textH, scroll = maxH, wsVScroll
	}
	bodyH := max(textH, icon)
	footH := int32(a.s(58))
	cw := pad + icon + gap + textW + pad
	if scroll != 0 {
		cw += int32(a.s(18))
	}
	ch := pad + bodyH + pad + footH
	mb.iconR = rect{pad, pad, pad + icon, pad + icon}
	mb.footY = ch - footH

	ty := pad + max(0, (icon-textH)/2)
	te, _, _ := pCreateWindowExW.Call(0, wp("EDIT"), wp(text), uintptr(wsChild|wsVisible|esMultiline|esReadOnly|scroll),
		uintptr(pad+icon+gap), uintptr(ty), uintptr(cw-pad-(pad+icon+gap)), uintptr(textH), hwnd, idMsgText, a.hInst, 0)
	send(te, wmSetFont, a.fontUI, 1)
	send(te, emSetMargins, 3, 0)
	if scroll != 0 {
		darkScrollbars(te)
	}

	type btn struct {
		id  int
		key string
	}
	btns := []btn{{idOK, "btn.ok"}}
	if mb.flags&0xF == mbYesNo {
		btns = []btn{{idYes, "btn.yes"}, {idNo, "btn.no"}}
	}
	bh := int32(a.s(34))
	x := cw - pad
	var def uintptr
	for i := len(btns) - 1; i >= 0; i-- {
		b := btns[i]
		bw := int32(max(a.s(104), a.measurePx(a.fontUI, T(b.key))+int(a.s(36))))
		x -= bw
		st := uint32(wsChild | wsVisible | wsTabStop | bsPushButton)
		if b.id == mbDefault() {
			st |= bsDefPushButton
		}
		h, _, _ := pCreateWindowExW.Call(0, wp("BUTTON"), wp(T(b.key)), uintptr(st),
			uintptr(x), uintptr(mb.footY+(footH-bh)/2), uintptr(bw), uintptr(bh), hwnd, uintptr(b.id), a.hInst, 0)
		send(h, wmSetFont, a.fontUI, 1)
		if b.id == mbDefault() {
			def = h
		}
		x -= int32(a.s(8))
	}

	// taille et position : centrée sur la fenêtre propriétaire, dans l'écran
	style := uint32(0x80000000 | 0x00C00000 | 0x00080000)
	wr := rect{0, 0, cw, ch}
	pAdjustWindowRectEx.Call(uintptr(unsafe.Pointer(&wr)), uintptr(style), 0, 0x1) // WS_EX_DLGMODALFRAME
	ww, wh := wr.r-wr.l, wr.b-wr.t
	var or rect
	owner, _, _ := pGetParent.Call(hwnd)
	if owner == 0 {
		owner = a.hwnd
	}
	pGetWindowRect.Call(owner, uintptr(unsafe.Pointer(&or)))
	px := or.l + (or.r-or.l-ww)/2
	py := or.t + (or.b-or.t-wh)/2
	px = max(wa.l, min(px, wa.r-ww))
	py = max(wa.t, min(py, wa.b-wh))
	pSetWindowPos.Call(hwnd, 0, uintptr(px), uintptr(py), uintptr(ww), uintptr(wh), 0x4|0x10) // NOZORDER | NOACTIVATE

	if mb.flags&0xF0 != 0 {
		pMessageBeep.Call(mb.flags & 0xF0)
	}
	pSetFocus.Call(def)
}

// measurePx : largeur d'un texte en pixels.
func (a *app) measurePx(font uintptr, s string) int {
	return a.measure(font, s) * a.dpi / 96
}

func (a *app) paintMsgDlg(hwnd, hdc uintptr) {
	var rc rect
	pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	fillBox(hdc, 0, 0, rc.r, mb.footY, colFond)
	fillBox(hdc, 0, mb.footY, rc.r, rc.b, colPanneau)
	fillBox(hdc, 0, mb.footY, rc.r, mb.footY+1, colFilet)
	// icône : rond coloré et symbole
	col, glyph, ink := colAccent, "i", colTexteAccent
	switch mb.flags & 0xF0 {
	case mbIconError:
		col, glyph = colRouge, "✕"
	case mbIconWarning:
		col, glyph, ink = colOrange, "!", colFond
	case mbIconQuestion:
		glyph = "?"
	}
	r := mb.iconR
	d := r.r - r.l
	disc(hdc, r.l+d/2, r.t+d/2, d/2, col)
	drawText(hdc, glyph, r, a.fontIcon, ink, dtCenter|dtVCenter|dtSingleLine)
}
