//go:build windows

package main

import (
	"strings"
	"syscall"
	"unsafe"
)

var (
	helpHwnd  uintptr
	helpEdit  uintptr
	helpClose uintptr
	helpClass bool
)

const (
	helpW  = 700
	helpH  = 640
	helpHd = 70
)

func helpProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	a := A
	switch uint32(msg) {
	case wmPaint:
		var ps paintStruct
		hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		var rc rect
		pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
		hdr := rect{0, 0, rc.r, int32(a.s(helpHd))}
		pFillRect.Call(hdc, uintptr(unsafe.Pointer(&hdr)), a.brHeader)
		acc := rect{0, int32(a.s(helpHd)), rc.r, int32(a.s(helpHd + 4))}
		pFillRect.Call(hdc, uintptr(unsafe.Pointer(&acc)), a.brAccent)
		tx := a.logoTitleX(hdc, 20, helpHd/2, 24)
		if tx == 20 && a.icoBig != 0 {
			pDrawIconEx.Call(hdc, uintptr(a.s(20)), uintptr(a.s(10)), a.icoBig, uintptr(a.s(50)), uintptr(a.s(50)), 0, 0, 3)
			tx = 86
		}
		pSetBkMode.Call(hdc, 1)
		old, _, _ := pSelectObject.Call(hdc, a.fontTitle)
		pSetTextColor.Call(hdc, colTitle)
		t := rect{int32(a.s(tx)), 0, rc.r - int32(a.s(16)), int32(a.s(helpHd))}
		pDrawTextW.Call(hdc, wp(T("help.title")), ^uintptr(0), uintptr(unsafe.Pointer(&t)), dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
		pSelectObject.Call(hdc, a.fontSmall)
		pSetTextColor.Call(hdc, colMuted)
		f := rect{int32(a.s(20)), int32(a.s(helpH - 50)), int32(a.s(helpW - 160)), int32(a.s(helpH - 16))}
		pDrawTextW.Call(hdc, wp(appSignature), ^uintptr(0), uintptr(unsafe.Pointer(&f)), dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
		pSelectObject.Call(hdc, old)
		pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0
	case wmCtlColorStatic:
		pSetBkColor.Call(wParam, rgb(255, 255, 255))
		pSetTextColor.Call(wParam, colText)
		return a.brWhite
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
			lpfnWndProc:   syscall.NewCallback(helpProc),
			hInstance:     a.hInst,
			hbrBackground: a.brBody,
			lpszClassName: cls,
		}
		wc.hCursor, _, _ = pLoadCursorW.Call(0, 32512)
		wc.hIcon, _, _ = pLoadImageW.Call(a.hInst, 1, 1, uintptr(a.s(32)), uintptr(a.s(32)), 0)
		wc.hIconSm, _, _ = pLoadImageW.Call(a.hInst, 1, 1, uintptr(a.s(16)), uintptr(a.s(16)), 0)
		wc.cbSize = uint32(unsafe.Sizeof(wc))
		pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
		helpClass = true
	}
	style := uint32(wsOverlapped | wsCaption | wsSysMenu | wsClipChildren)
	rc := rect{0, 0, int32(a.s(helpW)), int32(a.s(helpH))}
	pAdjustWindowRectEx.Call(uintptr(unsafe.Pointer(&rc)), uintptr(style), 0, wsExControlPar)
	ww, wh := rc.r-rc.l, rc.b-rc.t
	sw, _, _ := pGetSystemMetrics.Call(0)
	sh, _, _ := pGetSystemMetrics.Call(1)
	x, y := (int32(sw)-ww)/2, (int32(sh)-wh)/2
	if y < 0 {
		y = 0
	}
	helpHwnd, _, _ = pCreateWindowExW.Call(wsExControlPar, uintptr(unsafe.Pointer(cls)), wp(T("help.caption")),
		uintptr(style), uintptr(x), uintptr(y), uintptr(ww), uintptr(wh), a.hwnd, 0, a.hInst, 0)
	if helpHwnd == 0 {
		return
	}
	mk := func(class, text string, st, ex uint32, x, y, w, h, id int) uintptr {
		hw, _, _ := pCreateWindowExW.Call(uintptr(ex), wp(class), wp(text), uintptr(st|wsChild|wsVisible),
			uintptr(a.s(x)), uintptr(a.s(y)), uintptr(a.s(w)), uintptr(a.s(h)), helpHwnd, uintptr(id), a.hInst, 0)
		send(hw, wmSetFont, a.fontUI, 1)
		return hw
	}
	helpEdit = mk("EDIT", strings.ReplaceAll(noticeText(), "\n", "\r\n"), esMultiline|esReadOnly|esAutoVScroll|wsVScroll, wsExClientEdge,
		20, helpHd+18, helpW-40, helpH-helpHd-18-64, idHelpText)
	helpClose = mk("BUTTON", T("btn.close"), bsDefPushButton|wsTabStop, 0, helpW-20-120, helpH-50, 120, 34, idHelpClose)
	enable(a.hwnd, false)
	pShowWindow.Call(helpHwnd, 1)
	pUpdateWindow.Call(helpHwnd)
	pSetFocus.Call(helpClose)
}
