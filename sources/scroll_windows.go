//go:build windows

package main

// Barre de défilement VAP pour le journal et la notice (RichEdit) : plus visible que
// la barre sombre de Windows (piste DOUX, curseur au contour des champs, bleu clair
// au survol). La barre de Windows reste en place, masquée par une région de
// découpe : c'est elle qui garde la position, la molette, le clavier et la sélection.
// Sans RichEdit (champ texte simple), la barre de Windows reste visible.

import (
	"syscall"
	"unsafe"
)

var (
	pGetScrollInfo  = user32.NewProc("GetScrollInfo")
	pSetWindowRgn   = user32.NewProc("SetWindowRgn")
	pCreateRectRgn  = gdi32.NewProc("CreateRectRgn")
	pSetCapture     = user32.NewProc("SetCapture")
	pReleaseCapture = user32.NewProc("ReleaseCapture")
	pGetClassNameW  = user32.NewProc("GetClassNameW")

	scrollCls    bool
	scrollWndCB  uintptr // procédure de la barre
	scrollSubCB  uintptr // sous-classement du texte suivi
	scrollTarget = map[uintptr]uintptr{} // barre -> texte
	scrollBar    = map[uintptr]uintptr{} // texte -> barre
	scrollLast   = map[uintptr]scrollInfo{}
	scrollState  = map[uintptr]*barState{}
)

const (
	wmVScroll       = 0x0115
	wmMouseWheel    = 0x020A
	wmLButtonUp     = 0x0202
	esDisableNoScrl = 0x2000 // RichEdit : barre toujours présente (largeur du texte constante)
	emGetScrollPos  = 0x0400 + 221
	emSetScrollPos  = 0x0400 + 222
)

type scrollInfo struct {
	size, mask             uint32
	min, max               int32
	page                   uint32
	pos, trackPos          int32
}

type barState struct {
	hover, drag bool
	grab        int32 // décalage du clic dans le curseur
}

func isRich(h uintptr) bool {
	buf := make([]uint16, 64)
	pGetClassNameW.Call(h, uintptr(unsafe.Pointer(&buf[0])), 64)
	return syscall.UTF16ToString(buf) == richEditClass
}

func getScroll(h uintptr) scrollInfo {
	si := scrollInfo{mask: 0x17} // SIF_ALL
	si.size = uint32(unsafe.Sizeof(si))
	pGetScrollInfo.Call(h, 1, uintptr(unsafe.Pointer(&si))) // SB_VERT
	return si
}

// newScrollBar : barre VAP pour le texte target (créée une fois, placée par placeScroll).
func (a *app) newScrollBar(parent, target uintptr) uintptr {
	if !isRich(target) {
		return 0
	}
	if !scrollCls {
		scrollWndCB = syscall.NewCallback(scrollProc)
		scrollSubCB = syscall.NewCallback(scrollSubProc)
		wc := wndClassEx{lpfnWndProc: scrollWndCB, hInstance: a.hInst, lpszClassName: w16("VAPDefilement")}
		wc.hCursor, _, _ = pLoadCursorW.Call(0, 32512)
		wc.cbSize = uint32(unsafe.Sizeof(wc))
		pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
		scrollCls = true
	}
	const wsClipSiblings = 0x04000000
	bar, _, _ := pCreateWindowExW.Call(0, wp("VAPDefilement"), 0, uintptr(wsChild|wsVisible|wsClipSiblings),
		0, 0, 1, 1, parent, 0, a.hInst, 0)
	if bar == 0 {
		return 0
	}
	scrollTarget[bar], scrollBar[target], scrollState[bar] = target, bar, &barState{}
	pSetWindowSubclass.Call(target, scrollSubCB, 2, 0)
	return bar
}

// placeScroll : texte et barre dans le rectangle r (pixels de la fenêtre parente).
func (a *app) placeScroll(target, bar uintptr, r rect) {
	if bar == 0 {
		pMoveWindow.Call(target, uintptr(r.l), uintptr(r.t), uintptr(r.r-r.l), uintptr(r.b-r.t), 1)
		return
	}
	bw, gap := int32(a.s(10)), int32(a.s(6))
	sb, _, _ := pGetSystemMetrics.Call(2) // SM_CXVSCROLL
	tw := r.r - r.l - bw - gap
	// la barre de Windows dépasse sous la marge et la barre VAP ; la région la cache
	pMoveWindow.Call(target, uintptr(r.l), uintptr(r.t), uintptr(tw+int32(sb)), uintptr(r.b-r.t), 1)
	rgn, _, _ := pCreateRectRgn.Call(0, 0, uintptr(tw), uintptr(r.b-r.t))
	pSetWindowRgn.Call(target, rgn, 1) // la région appartient désormais à la fenêtre
	pMoveWindow.Call(bar, uintptr(r.r-bw), uintptr(r.t), uintptr(bw), uintptr(r.b-r.t), 1)
	pSetWindowPos.Call(bar, 0, 0, 0, 0, 0, 0x1|0x2|0x10) // au-dessus du texte (HWND_TOP, sans bouger)
	// marge intérieure du texte
	var cr rect
	pGetClientRect.Call(target, uintptr(unsafe.Pointer(&cr)))
	cr.l += int32(a.s(2))
	cr.t += int32(a.s(2))
	send(target, emSetRect, 0, uintptr(unsafe.Pointer(&cr)))
	pInvalidateRect.Call(bar, 0, 1)
}

// thumb : position et hauteur du curseur (pixels) ; ok = faux si tout le texte est visible.
func thumb(bar uintptr, si scrollInfo) (y, h int32, ok bool) {
	var r rect
	pGetClientRect.Call(bar, uintptr(unsafe.Pointer(&r)))
	rng := si.max - si.min + 1
	if si.page == 0 || int32(si.page) >= rng {
		return 0, r.b, false
	}
	h = max(r.b*int32(si.page)/rng, r.r*3)
	y = (si.pos - si.min) * (r.b - h) / max(1, rng-int32(si.page))
	return max(0, min(y, r.b-h)), h, true
}

func setScrollPos(target uintptr, pos int32) {
	var p point
	send(target, emGetScrollPos, 0, uintptr(unsafe.Pointer(&p)))
	p.y = pos
	send(target, emSetScrollPos, 0, uintptr(unsafe.Pointer(&p)))
}

func scrollProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	target := scrollTarget[hwnd]
	st := scrollState[hwnd]
	switch uint32(msg) {
	case wmPaint:
		paintBuffered(hwnd, func(hdc uintptr) {
			var r rect
			pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
			fillBox(hdc, r.l, r.t, r.r, r.b, colPanneau)
			roundBox(hdc, r, r.r/2, colDoux, colDoux)
			if y, h, ok := thumb(hwnd, getScroll(target)); ok {
				c := colBord
				switch {
				case st != nil && st.drag:
					c = colAccentClair
				case st != nil && st.hover:
					c = colBordSurvol
				}
				roundBox(hdc, rect{0, y, r.r, y + h}, r.r/2, c, c)
			}
		})
		return 0
	case wmEraseBkgnd:
		return 1
	case wmMouseMove:
		y := int32(int16(lParam >> 16 & 0xFFFF))
		if st != nil && st.drag {
			si := getScroll(target)
			_, h, ok := thumb(hwnd, si)
			var r rect
			pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
			if ok && r.b > h {
				rng := si.max - si.min + 1 - int32(si.page)
				setScrollPos(target, si.min+(y-st.grab)*rng/(r.b-h))
			}
		} else if st != nil && !st.hover {
			st.hover = true
			tm := trackMouse{size: uint32(unsafe.Sizeof(trackMouse{})), flags: 0x2, hwnd: hwnd}
			pTrackMouseEvent.Call(uintptr(unsafe.Pointer(&tm)))
			pInvalidateRect.Call(hwnd, 0, 0)
		}
		return 0
	case wmMouseLeave:
		if st != nil {
			st.hover = false
			pInvalidateRect.Call(hwnd, 0, 0)
		}
		return 0
	case wmLButtonDown:
		y := int32(int16(lParam >> 16 & 0xFFFF))
		ty, h, ok := thumb(hwnd, getScroll(target))
		switch {
		case !ok || st == nil:
		case y >= ty && y < ty+h: // saisie du curseur
			st.drag, st.grab = true, y-ty
			pSetCapture.Call(hwnd)
		case y < ty:
			send(target, wmVScroll, 2, 0) // SB_PAGEUP
		default:
			send(target, wmVScroll, 3, 0) // SB_PAGEDOWN
		}
		pInvalidateRect.Call(hwnd, 0, 0)
		return 0
	case wmLButtonUp:
		if st != nil && st.drag {
			st.drag = false
			pReleaseCapture.Call()
			pInvalidateRect.Call(hwnd, 0, 0)
		}
		return 0
	case wmMouseWheel:
		return send(target, uint32(msg), wParam, lParam)
	case wmNcDestroy:
		delete(scrollBar, target)
		delete(scrollTarget, hwnd)
		delete(scrollState, hwnd)
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return r
}

// scrollSubProc : après chaque message du texte, la barre VAP suit la position.
func scrollSubProc(hwnd, msg, wParam, lParam, id, ref uintptr) uintptr {
	r, _, _ := pDefSubclassProc.Call(hwnd, msg, wParam, lParam)
	if uint32(msg) == wmNcDestroy {
		pRemoveWindowSubclass.Call(hwnd, scrollSubCB, id)
		delete(scrollLast, hwnd)
		return r
	}
	if bar := scrollBar[hwnd]; bar != 0 {
		si := getScroll(hwnd)
		if l, ok := scrollLast[hwnd]; !ok || l.pos != si.pos || l.max != si.max || l.page != si.page {
			scrollLast[hwnd] = si
			pInvalidateRect.Call(bar, 0, 0)
		}
	}
	return r
}
