//go:build windows

package main

// Logo VAP (symbole + mot, blanc avec la facette bleue) dessiné dans les bandeaux.
// L'image est intégrée au programme (res/logo_entete.png), réduite à la taille voulue
// puis affichée avec sa transparence (AlphaBlend).

import (
	"bytes"
	_ "embed"
	"image"
	"image/draw"
	"image/png"
	"syscall"
	"unsafe"
)

//go:embed res/logo_entete.png
var logoPNG []byte

var (
	msimg32                = syscall.NewLazyDLL("msimg32.dll")
	pAlphaBlend            = msimg32.NewProc("AlphaBlend")
	pCreateDIBSection      = gdi32.NewProc("CreateDIBSection")
	pCreateCompatibleDC    = gdi32.NewProc("CreateCompatibleDC")
	pDeleteDC              = gdi32.NewProc("DeleteDC")
	logoSrc                *image.NRGBA
	logoCache              = map[int]logoBmp{} // par hauteur en pixels
)

type logoBmp struct {
	hbm  uintptr
	w, h int
}

type bitmapInfoHeader struct {
	biSize          uint32
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

// logoBitmap : le logo réduit à h pixels de haut, en BGRA prémultiplié (une fois par taille).
func logoBitmap(h int) (logoBmp, bool) {
	if b, ok := logoCache[h]; ok {
		return b, b.hbm != 0
	}
	if logoSrc == nil {
		img, err := png.Decode(bytes.NewReader(logoPNG))
		if err != nil {
			logoCache[h] = logoBmp{}
			return logoBmp{}, false
		}
		logoSrc = image.NewNRGBA(img.Bounds())
		draw.Draw(logoSrc, logoSrc.Bounds(), img, img.Bounds().Min, draw.Src)
	}
	sw, sh := logoSrc.Bounds().Dx(), logoSrc.Bounds().Dy()
	w := max(1, (sw*h+sh/2)/sh)
	bi := bitmapInfoHeader{biSize: 40, biWidth: int32(w), biHeight: -int32(h), biPlanes: 1, biBitCount: 32}
	var bits uintptr
	hbm, _, _ := pCreateDIBSection.Call(0, uintptr(unsafe.Pointer(&bi)), 0, uintptr(unsafe.Pointer(&bits)), 0, 0)
	if hbm == 0 || bits == 0 {
		logoCache[h] = logoBmp{}
		return logoBmp{}, false
	}
	dst := unsafe.Slice((*byte)(unsafe.Pointer(bits)), w*h*4)
	// réduction par moyenne de surface, en couleurs prémultipliées
	for y := 0; y < h; y++ {
		y0, y1 := y*sh/h, max(y*sh/h+1, (y+1)*sh/h)
		for x := 0; x < w; x++ {
			x0, x1 := x*sw/w, max(x*sw/w+1, (x+1)*sw/w)
			var r, g, b, a, n uint32
			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					o := logoSrc.PixOffset(sx, sy)
					p := logoSrc.Pix[o : o+4]
					al := uint32(p[3])
					r += uint32(p[0]) * al
					g += uint32(p[1]) * al
					b += uint32(p[2]) * al
					a += al
					n++
				}
			}
			d := dst[(y*w+x)*4:]
			d[0] = byte(b / 255 / n)
			d[1] = byte(g / 255 / n)
			d[2] = byte(r / 255 / n)
			d[3] = byte(a / n)
		}
	}
	lb := logoBmp{hbm, w, h}
	logoCache[h] = lb
	return lb, true
}

// drawLogo : dessine le logo, bord gauche x, centré verticalement sur cy (pixels) ;
// renvoie sa largeur en pixels (0 si l'image n'a pas pu être préparée).
func drawLogo(hdc uintptr, x, cy, h int) int {
	lb, ok := logoBitmap(h)
	if !ok {
		return 0
	}
	mem, _, _ := pCreateCompatibleDC.Call(hdc)
	old, _, _ := pSelectObject.Call(mem, lb.hbm)
	const blend = 0 | 0<<8 | 255<<16 | 1<<24 // AC_SRC_OVER, opacité 255, AC_SRC_ALPHA
	pAlphaBlend.Call(hdc, uintptr(x), uintptr(cy-lb.h/2), uintptr(lb.w), uintptr(lb.h), mem, 0, 0, uintptr(lb.w), uintptr(lb.h), blend)
	pSelectObject.Call(mem, old)
	pDeleteDC.Call(mem)
	return lb.w
}

// logoTitleX : dessine logo + trait de séparation dans un bandeau ; renvoie l'abscisse
// (unités logiques) où commence le titre.
func (a *app) logoTitleX(hdc uintptr, x, cy, h int) int {
	w := drawLogo(hdc, a.s(x), a.s(cy), a.s(h))
	if w == 0 {
		return x
	}
	lw := w * 96 / a.dpi
	sep := x + lw + 11
	line(hdc, int32(max(1, a.s(1))), rgb(70, 90, 120), int32(a.s(sep)), int32(a.s(cy-h)), int32(a.s(sep)), int32(a.s(cy+h)))
	return sep + 12
}
