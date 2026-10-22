//go:build windows && capture

package main

// Captures d'écran de contrôle de l'interface (jamais dans le programme livré :
// compilé seulement avec « go build -tags capture »). Variables d'environnement :
//   VAP_CAPTURE_DIR  dossier des images (obligatoire)
//   VAP_CAPTURE_DPI  échelle simulée (96 = 100 %, 120 = 125 %, 144 = 150 %)
//   VAP_CAPTURE_ECRAN zone de travail simulée en pixels, ex. 1366x720
//   VAP_CAPTURE_LANGUES liste de codes (fr,de,ja…) ; défaut : toutes
// Aucune conversion n'est lancée et aucun fichier utilisateur n'est lu ni écrit.

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

var (
	capDir     string
	capPrefix  string
	capSteps   []func()
	capStep    int
	capTimerCB uintptr
	pSetTimer  = user32.NewProc("SetTimer")
	pKillTimer = user32.NewProc("KillTimer")
	pLastPopup = user32.NewProc("GetLastActivePopup")
	pPrintWnd  = user32.NewProc("PrintWindow")
	pGetDIBits = gdi32.NewProc("GetDIBits")
)

func init() {
	capDir = os.Getenv("VAP_CAPTURE_DIR")
	if capDir == "" {
		return
	}
	if d, err := strconv.Atoi(os.Getenv("VAP_CAPTURE_DPI")); err == nil && d > 0 {
		testDpi = d
	}
	if e := os.Getenv("VAP_CAPTURE_ECRAN"); e != "" {
		var w, h int
		if _, err := fmt.Sscanf(e, "%dx%d", &w, &h); err == nil {
			testWorkArea = rect{0, 0, int32(w), int32(h)}
		}
	}
	capPrefix = os.Getenv("VAP_CAPTURE_PREFIXE")
	os.MkdirAll(capDir, 0o755)
	capTimerCB = syscall.NewCallback(func(hwnd, msg, id, t uintptr) uintptr {
		pKillTimer.Call(hwnd, id)
		if capStep < len(capSteps) {
			f := capSteps[capStep]
			capStep++
			next(400)
			f()
		}
		return 0
	})
	afterShow = func() {
		a := A
		langs := []Lang{}
		for _, c := range strings.Split(os.Getenv("VAP_CAPTURE_LANGUES"), ",") {
			if l, ok := langFromCode(strings.TrimSpace(c)); ok {
				langs = append(langs, l)
			}
		}
		if len(langs) == 0 {
			for l := Lang(0); l < nLang; l++ {
				langs = append(langs, l)
			}
		}
		const wmChangeUIState = 0x0127
		send(a.hwnd, wmChangeUIState, 2|1<<16, 0) // UIS_CLEAR, UISF_HIDEFOCUS : focus visible
		for _, l := range langs {
			l := l
			capSteps = append(capSteps,
				func() {
					cur = l
					a.applyLanguage(false)
					a.showWelcome()
				},
				func() { shot(a.hwnd, "1-accueil-"+langCodes[l]) },
			)
		}
		capSteps = append(capSteps,
			func() { // états remplis : champs, journal coloré, progression, focus sur Convertir
				cur = langs[0]
				a.applyLanguage(false)
				setText(a.ctl[idDatEdit], `C:\Exemple\Réglages Epilog\MINI24 40 watts`)
				setText(a.ctl[idBkEdit], `C:\Exemple\Sauvegarde Job Manager.epidb`)
				setText(a.ctl[idFolderName], "MINI24 40 watts")
				setText(a.ctl[idDatInfo], Tf("info.count", 112))
				for _, k := range []string{"MINI24 40 watts", "Fusion 32 Fibre"} {
					send(a.ctl[idMachine], cbAddString, 0, wp(k))
				}
				send(a.ctl[idMachine], cbSetCurSel, 0, 0)
				a.clearLog()
				a.log(Tf("log.backupLoaded", 2))
				a.log("")
				a.log(Tf("log.convStart", "MINI24 40 watts", "MINI24 40 watts"))
				a.log(T("log.copy"))
				a.log(Tf("log.read", 112))
				a.log(Tf("log.skipped", "Essai.DAT", T("skip.testGridLog")))
				a.log(Tf("log.note", "Bois 3 mm.DAT", "Tampon"))
				a.log(Tf("log.warnings", 3))
				a.log(Tf("log.done", 111, 1))
				a.log(T("log.outDir") + `C:\Exemple\MINI24 40 watts - Conversion Job Manager`)
				a.log(T("log.fail") + "exemple d'erreur")
				a.setProgress(62)
				pSetFocus.Call(a.ctl[idConvert])
			},
			func() { shot(a.hwnd, "2-rempli") },
			func() { // pendant une conversion : contrôles verrouillés
				for _, id := range []int{idConvert, idDatBrowse, idBkBrowse, idOpen, idDatEdit, idBkEdit, idRecursive, idMachine, idFolderName, idSort, idTop, idLaser, idLang} {
					enable(a.ctl[id], false)
				}
			},
			func() { shot(a.hwnd, "3-verrouille") },
			func() {
				for _, id := range []int{idConvert, idDatBrowse, idBkBrowse, idDatEdit, idBkEdit, idRecursive, idMachine, idFolderName, idSort, idTop, idLaser, idLang} {
					enable(a.ctl[id], true)
				}
				enable(a.ctl[idOpen], true)
				pSetFocus.Call(a.ctl[idDatEdit])
			},
			func() { shot(a.hwnd, "4-actif") },
			func() { a.showHelp() },
			func() { shot(helpHwnd, "5-notice") },
			func() { pPostMessageW.Call(helpHwnd, wmClose, 0, 0) },
			func() { dialogShot("6-question", 0); msgBox(a.hwnd, Tf("ask.replace", "MINI24 40 watts"), T("msg.title"), mbYesNo|mbIconWarning) },
			func() { dialogShot("7-termine", 0); msgBox(a.hwnd, Tf("msg.done", 111, "MINI24 40 watts")+Tf("msg.summary", 108, 3, 1), T("msg.title"), mbYesNo|mbIconInfo) },
			func() { dialogShot("8-erreur", 0); msgBox(a.hwnd, T("err.openBackup")+"exemple", T("msg.title"), mbIconError) },
			func() { dialogShot("9-suggestion", 0); msgBox(a.hwnd, suggestText(""), T("msg.title"), mbOK|mbIconInfo) },
			func() { pPostMessageW.Call(a.hwnd, wmClose, 0, 0) },
		)
		next(700)
	}
}

func next(ms int) { pSetTimer.Call(A.hwnd, 77, uintptr(ms), capTimerCB) }

// dialogShot : prochaine étape = capture de la boîte ouverte puis fermeture (Échap).
func dialogShot(name string, _ int) {
	capSteps = append(capSteps[:capStep], append([]func(){
		func() {
			h, _, _ := pLastPopup.Call(A.hwnd) // la boîte ouverte par le programme, jamais une autre fenêtre
			if h == 0 || h == A.hwnd {
				return
			}
			shot(h, name)
			pPostMessageW.Call(h, wmCommand, idCancel, 0)
		},
	}, capSteps[capStep:]...)...)
}

// shot : capture de la fenêtre telle qu'elle est à l'écran (barre de titre comprise).
func shot(hwnd uintptr, name string) {
	pUpdateWindow.Call(hwnd)
	var r rect
	pGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	w, h := int(r.r-r.l), int(r.b-r.t)
	scr, _, _ := pGetDC.Call(0)
	mem, _, _ := pCreateCompatibleDC.Call(scr)
	bmp, _, _ := pCreateCompatibleBitmap.Call(scr, uintptr(w), uintptr(h))
	ob, _, _ := pSelectObject.Call(mem, bmp)
	if ok, _, _ := pPrintWnd.Call(hwnd, mem, 2); ok == 0 { // PW_RENDERFULLCONTENT, même si la fenêtre est cachée
		pBitBlt.Call(mem, 0, 0, uintptr(w), uintptr(h), scr, uintptr(r.l), uintptr(r.t), 0x00CC0020)
	}
	pSelectObject.Call(mem, ob)
	bi := bitmapInfoHeader{biSize: 40, biWidth: int32(w), biHeight: -int32(h), biPlanes: 1, biBitCount: 32}
	buf := make([]byte, w*h*4)
	pGetDIBits.Call(mem, bmp, 0, uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&bi)), 0)
	pDeleteObject.Call(bmp)
	pDeleteDC.Call(mem)
	pReleaseDC.Call(0, scr)
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < w*h; i++ {
		img.Pix[i*4], img.Pix[i*4+1], img.Pix[i*4+2], img.Pix[i*4+3] = buf[i*4+2], buf[i*4+1], buf[i*4], 255
	}
	f, err := os.Create(filepath.Join(capDir, capPrefix+name+".png"))
	if err != nil {
		return
	}
	png.Encode(f, img)
	f.Close()
}
