//go:build windows

package main

// Drapeaux devant chaque langue dans la liste déroulante (dessin GDI, sans image).
// Ordre : fr, en (Royaume-Uni), es, de, it, pt, nl, zh, ja — comme langCodes.

import (
	"math"
	"unsafe"
)

var (
	pDeleteObject      = gdi32.NewProc("DeleteObject")
	pPolygon           = gdi32.NewProc("Polygon")
	pSaveDC            = gdi32.NewProc("SaveDC")
	pRestoreDC         = gdi32.NewProc("RestoreDC")
	pIntersectClipRect = gdi32.NewProc("IntersectClipRect")
	pGetSysColor       = user32.NewProc("GetSysColor")
	pDrawFocusRect     = user32.NewProc("DrawFocusRect")
)

const (
	wmDrawItem    = 0x002B
	wmMeasureItem = 0x002C
	odsSelected   = 0x0001
	odsFocus      = 0x0010
	odsComboEdit  = 0x1000
)

type measureItem struct {
	ctlType, ctlID, itemID, itemWidth, itemHeight uint32
	itemData                                      uintptr
}

type drawItem struct {
	ctlType, ctlID, itemID, itemAction, itemState uint32
	hwndItem, hDC                                 uintptr
	rcItem                                        rect
	itemData                                      uintptr
}

func fillBox(hdc uintptr, l, t, r, b int32, col uintptr) {
	br, _, _ := pCreateSolidBrush.Call(col)
	rc := rect{l, t, r, b}
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), br)
	pDeleteObject.Call(br)
}

// bands : bandes égales, verticales ou horizontales.
func bands(hdc uintptr, x, y, w, h int32, vertical bool, cols ...uintptr) {
	n := int32(len(cols))
	for i, c := range cols {
		k := int32(i)
		if vertical {
			fillBox(hdc, x+w*k/n, y, x+w*(k+1)/n, y+h, c)
		} else {
			fillBox(hdc, x, y+h*k/n, x+w, y+h*(k+1)/n, c)
		}
	}
}

func line(hdc uintptr, width int32, col uintptr, x1, y1, x2, y2 int32) {
	pen, _, _ := pCreatePen.Call(0, uintptr(width), col)
	old, _, _ := pSelectObject.Call(hdc, pen)
	pMoveToEx.Call(hdc, uintptr(x1), uintptr(y1), 0)
	pLineTo.Call(hdc, uintptr(x2), uintptr(y2))
	pSelectObject.Call(hdc, old)
	pDeleteObject.Call(pen)
}

func disc(hdc uintptr, cx, cy, r int32, col uintptr) {
	br, _, _ := pCreateSolidBrush.Call(col)
	pen, _, _ := pCreatePen.Call(0, 1, col)
	ob, _, _ := pSelectObject.Call(hdc, br)
	op, _, _ := pSelectObject.Call(hdc, pen)
	pEllipse.Call(hdc, uintptr(cx-r), uintptr(cy-r), uintptr(cx+r+1), uintptr(cy+r+1))
	pSelectObject.Call(hdc, ob)
	pSelectObject.Call(hdc, op)
	pDeleteObject.Call(br)
	pDeleteObject.Call(pen)
}

func star(hdc uintptr, cx, cy, r float64, col uintptr) {
	var pts [10]point
	for i := range pts {
		rr := r
		if i%2 == 1 {
			rr = r * 0.4
		}
		a := -math.Pi/2 + float64(i)*math.Pi/5
		pts[i] = point{int32(math.Round(cx + rr*math.Cos(a))), int32(math.Round(cy + rr*math.Sin(a)))}
	}
	br, _, _ := pCreateSolidBrush.Call(col)
	pen, _, _ := pCreatePen.Call(0, 1, col)
	ob, _, _ := pSelectObject.Call(hdc, br)
	op, _, _ := pSelectObject.Call(hdc, pen)
	pPolygon.Call(hdc, uintptr(unsafe.Pointer(&pts[0])), 10)
	pSelectObject.Call(hdc, ob)
	pSelectObject.Call(hdc, op)
	pDeleteObject.Call(br)
	pDeleteObject.Call(pen)
}

// drawFlag : drapeau de la langue l dans le rectangle (x, y, w, h).
func drawFlag(hdc uintptr, l Lang, x, y, w, h int32) {
	saved, _, _ := pSaveDC.Call(hdc)
	pIntersectClipRect.Call(hdc, uintptr(x), uintptr(y), uintptr(x+w), uintptr(y+h))
	white := rgb(255, 255, 255)
	switch l {
	case LFR:
		bands(hdc, x, y, w, h, true, rgb(0, 35, 149), white, rgb(237, 41, 57))
	case LEN: // Royaume-Uni
		fillBox(hdc, x, y, x+w, y+h, rgb(1, 33, 105))
		d := max(2, h/5)
		line(hdc, d, white, x, y, x+w, y+h)
		line(hdc, d, white, x+w, y, x, y+h)
		line(hdc, max(1, d/2), rgb(200, 16, 46), x, y, x+w, y+h)
		line(hdc, max(1, d/2), rgb(200, 16, 46), x+w, y, x, y+h)
		c := max(3, h/3)
		fillBox(hdc, x+w/2-c/2-1, y, x+w/2+c/2+1, y+h, white)
		fillBox(hdc, x, y+h/2-c/2-1, x+w, y+h/2+c/2+1, white)
		fillBox(hdc, x+w/2-c/2, y, x+w/2+c/2, y+h, rgb(200, 16, 46))
		fillBox(hdc, x, y+h/2-c/2, x+w, y+h/2+c/2, rgb(200, 16, 46))
	case LES:
		fillBox(hdc, x, y, x+w, y+h, rgb(170, 21, 27))
		fillBox(hdc, x, y+h/4, x+w, y+h*3/4, rgb(241, 191, 0))
	case LDE:
		bands(hdc, x, y, w, h, false, rgb(0, 0, 0), rgb(221, 0, 0), rgb(255, 206, 0))
	case LIT:
		bands(hdc, x, y, w, h, true, rgb(0, 146, 70), white, rgb(206, 43, 55))
	case LPT:
		fillBox(hdc, x, y, x+w, y+h, rgb(255, 0, 0))
		fillBox(hdc, x, y, x+w*2/5, y+h, rgb(0, 102, 0))
		disc(hdc, x+w*2/5, y+h/2, max(2, h/4), rgb(255, 204, 0))
	case LNL:
		bands(hdc, x, y, w, h, false, rgb(174, 28, 40), white, rgb(33, 70, 139))
	case LZH:
		fillBox(hdc, x, y, x+w, y+h, rgb(238, 28, 37))
		u := float64(h) / 20
		star(hdc, float64(x)+5*u, float64(y)+5*u, 3.6*u, rgb(255, 255, 0))
		for _, p := range [][2]float64{{10, 2}, {12, 4}, {12, 7}, {10, 9}} {
			disc(hdc, x+int32(p[0]*u), y+int32(p[1]*u), max(1, int32(u)), rgb(255, 255, 0))
		}
	case LJA:
		fillBox(hdc, x, y, x+w, y+h, white)
		disc(hdc, x+w/2, y+h/2, h*3/10, rgb(188, 0, 45))
	}
	pRestoreDC.Call(hdc, saved)
	// liseré gris pour les drapeaux clairs (Japon…)
	pen, _, _ := pCreatePen.Call(0, 1, rgb(190, 190, 190))
	op, _, _ := pSelectObject.Call(hdc, pen)
	nb, _, _ := pGetStockObject.Call(5) // NULL_BRUSH
	ob, _, _ := pSelectObject.Call(hdc, nb)
	pRectangle.Call(hdc, uintptr(x), uintptr(y), uintptr(x+w), uintptr(y+h))
	pSelectObject.Call(hdc, ob)
	pSelectObject.Call(hdc, op)
	pDeleteObject.Call(pen)
}

var pRectangle = gdi32.NewProc("Rectangle")

// drawLangItem : une ligne de la liste des langues (drapeau + nom).
func (a *app) drawLangItem(di *drawItem) {
	if int32(di.itemID) < 0 || di.itemID >= uint32(nLang) {
		return
	}
	hdc, r := di.hDC, di.rcItem
	bg, fg := uintptr(5), uintptr(8) // COLOR_WINDOW, COLOR_WINDOWTEXT
	if di.itemState&odsSelected != 0 && di.itemState&odsComboEdit == 0 {
		bg, fg = 13, 14 // COLOR_HIGHLIGHT, COLOR_HIGHLIGHTTEXT
	}
	cb, _, _ := pGetSysColor.Call(bg)
	ct, _, _ := pGetSysColor.Call(fg)
	fillBox(hdc, r.l, r.t, r.r, r.b, cb)
	fh := int32(a.s(13))
	fw := fh * 3 / 2
	fx := r.l + int32(a.s(5))
	fy := r.t + (r.b-r.t-fh)/2
	drawFlag(hdc, Lang(di.itemID), fx, fy, fw, fh)
	pSetBkMode.Call(hdc, 1) // TRANSPARENT
	pSetTextColor.Call(hdc, ct)
	tr := rect{fx + fw + int32(a.s(8)), r.t, r.r - 2, r.b}
	pDrawTextW.Call(hdc, wp(langNames[di.itemID]), ^uintptr(0), uintptr(unsafe.Pointer(&tr)), dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
	if di.itemState&odsFocus != 0 && di.itemState&odsComboEdit == 0 {
		pDrawFocusRect.Call(hdc, uintptr(unsafe.Pointer(&r)))
	}
}
