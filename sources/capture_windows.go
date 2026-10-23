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
		// thème imposé pour la série de captures, sans l'enregistrer
		if os.Getenv("VAP_CAPTURE_THEME") == "clair" {
			setPaletteDark(false)
			a.applyTheme()
		} else if os.Getenv("VAP_CAPTURE_THEME") == "sombre" {
			setPaletteDark(true)
			a.applyTheme()
		}
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
		keyTest := func(label, key string, vk uintptr, ask func() int) func() {
			return func() {
				capSteps = append(capSteps[:capStep], append([]func(){func() {
					h, _, _ := pLastPopup.Call(A.hwnd)
					f, _, _ := pGetFocus.Call()
					kb(fmt.Sprintf("%s : focus initial sur « %s », touche %s", label, getText(f), key))
					if h != 0 && h != A.hwnd {
						pPostMessageW.Call(f, 0x0100, vk, 0) // WM_KEYDOWN
					}
				}}, capSteps[capStep:]...)...)
				r := ask()
				names := map[int]string{idYes: "action 1 (IDYES)", idNo: "action 2 (IDNO)", idOK: "OK"}
				kb("  -> réponse : " + names[r])
			}
		}
		replace := func() int {
			return msgChoice(a.hwnd, Tf("ask.replace", "MINI24 40 watts"), T("msg.title"), mbIconWarning, "btn.replace", "btn.cancel", false)
		}
		finish := func() int {
			return msgChoice(a.hwnd, Tf("msg.done", 111, "MINI24 40 watts"), T("msg.title"), mbIconInfo, "btn.open", "btn.close", true)
		}
		longDat := `C:\Users\Atelier\Documents\Epilog\Engraving_Settings\Anciens réglages de l'atelier de gravure\Mini 24 40 watts - réglages validés par l'atelier`
		longFolder := "Mini 24 40 watts – atelier principal – réglages validés 2026"
		for _, l := range langs {
			l := l
			p := langCodes[l] + "-"
			capSteps = append(capSteps,
				func() {
					for _, id := range []int{idDatEdit, idBkEdit, idFolderName} { // état initial
						setText(a.ctl[id], "")
					}
					send(a.ctl[idMachine], cbResetContent, 0, 0)
					enable(a.ctl[idOpen], false)
					a.setProgress(0)
					cur = l
					a.applyLanguage(false)
					a.showWelcome()
					pSetFocus.Call(a.ctl[idDatBrowse])
				},
				func() { shot(a.hwnd, p+"1-accueil"); a.checkFit(p + "fenêtre principale") },
				func() { // états remplis : chemins longs, journal coloré et défilé en bas, progression
					a.fillDemo(longDat, longFolder)
					pSetFocus.Call(a.ctl[idConvert])
				},
				func() { shot(a.hwnd, p+"2-rempli"); a.checkFit(p + "fenêtre remplie") },
				func() { // pendant une conversion : contrôles verrouillés (comme convert())
					for _, id := range lockedIDs {
						enable(a.ctl[id], false)
					}
				},
				func() { shot(a.hwnd, p+"3-verrouille"); a.checkFit(p + "fenêtre verrouillée") },
				func() {
					for _, id := range lockedIDs {
						enable(a.ctl[id], id != idOpen)
					}
					enable(a.ctl[idOpen], true)
				},
				func() { a.showHelp() },
				func() { shot(helpHwnd, p+"4-notice") },
				func() { pPostMessageW.Call(helpHwnd, wmClose, 0, 0) },
				func() { kb("[" + langCodes[l] + "]") },
				keyTest("Remplacer / Annuler", "Entrée", 0x0D, replace),
				keyTest("Remplacer / Annuler", "Échap", 0x1B, replace),
				keyTest("Ouvrir le dossier / Fermer", "Entrée", 0x0D, finish),
				keyTest("Ouvrir le dossier / Fermer", "Échap", 0x1B, finish),
				func() {
					dialogShot(p + "5-remplacer")
					msgChoice(a.hwnd, Tf("ask.replace", longFolder), T("msg.title"), mbIconWarning, "btn.replace", "btn.cancel", false)
				},
				func() {
					dialogShot(p + "6-termine")
					done := withSummary(Tf("msg.done", 111, longFolder), Tf("msg.summary", 108, 3, 1))
					msgChoice(a.hwnd, done, T("msg.title"), mbIconInfo, "btn.open", "btn.close", true)
				},
			)
		}
		capSteps = append(capSteps,
			func() {
				cur = langs[0]
				a.applyLanguage(false)
				a.fillDemo(longDat, longFolder)
			},
			func() { dialogShot("7-erreur"); msgBox(a.hwnd, T("err.convFailed")+"exemple\nsur plusieurs lignes"+Tf("ask.sendReport", `C:\Exemple\rapport d'erreur.txt`)+T("rep.privacy"), T("msg.title"), mbYesNo|mbIconError) },
			func() { dialogShot("8-suggestion"); msgBox(a.hwnd, suggestText(`C:\Exemple\rapport d'erreur.txt`), T("msg.title"), mbOK|mbIconInfo) },
		)
		if os.Getenv("VAP_CAPTURE_BASCULE") != "" {
			capSteps = append(capSteps, a.toggleSteps()...)
		}
		// clavier : ordre de tabulation, puis Entrée et Échap dans les deux questions
		capSteps = append(capSteps, func() { pSetFocus.Call(a.ctl[idHelp]); kb("Tabulation depuis « Notice » :") })
		for i := 0; i < 22; i++ {
			capSteps = append(capSteps, func() {
				f, _, _ := pGetFocus.Call()
				id, _, _ := pGetDlgCtrlID.Call(f)
				kb(fmt.Sprintf("  %2d. %s (%q)", i+1, ctlName(int(id)), getText(f)))
				pPostMessageW.Call(f, 0x0100, 0x09, 0) // WM_KEYDOWN Tab
			})
		}
		capSteps = append(capSteps,
			func() {
				os.WriteFile(filepath.Join(capDir, capPrefix+"clavier.txt"), []byte(strings.Join(kbLines, "\r\n")+"\r\n"), 0o644)
				pPostMessageW.Call(a.hwnd, wmClose, 0, 0)
			},
		)
		next(700)
	}
}

func next(ms int) { pSetTimer.Call(A.hwnd, 77, uintptr(ms), capTimerCB) }

// dialogShot : prochaine étape = capture de la boîte ouverte puis fermeture (Échap).
func dialogShot(name string) {
	capSteps = append(capSteps[:capStep], append([]func(){
		func() {
			h, _, _ := pLastPopup.Call(A.hwnd) // la boîte ouverte par le programme, jamais une autre fenêtre
			if h == 0 || h == A.hwnd {
				return
			}
			shot(h, name)
			if te, _, _ := pGetDlgItem.Call(h, idMsgText); te != 0 {
				var r rect
				pGetClientRect.Call(te, uintptr(unsafe.Pointer(&r)))
				kb(fmt.Sprintf("%s : %d lignes, zone de texte %d px, ligne %d px, première ligne visible %d", name,
					send(te, 0x00BA, 0, 0), r.b, textHeight(h, A.fontUI), send(te, 0x00CE, 0, 0)))
			}
			pPostMessageW.Call(h, wmCommand, idCancel, 0)
		},
	}, capSteps[capStep:]...)...)
}

var (
	kbLines       []string
	pGetDlgCtrlID = user32.NewProc("GetDlgCtrlID")
)

func kb(s string) { kbLines = append(kbLines, s) }

func ctlName(id int) string {
	names := map[int]string{idHelp: "Notice", idLang: "Langue", idDatEdit: "Dossier .DAT", idDatBrowse: "Parcourir 1",
		idRecursive: "Sous-dossiers", idBkEdit: "Sauvegarde", idBkBrowse: "Parcourir 2", idMachine: "Machine",
		idFolderName: "Nom du dossier", idAir: "Air", idSort: "Tri", idTop: "En tête", idLaser: "Source laser",
		idOpen: "Ouvrir le dossier", idConvert: "Convertir", idQuit: "Quitter", idDonate: "Don", idSuggest: "Suggestion",
		idTheme: "Mode clair / sombre"}
	if n, ok := names[id]; ok {
		return n
	}
	return fmt.Sprintf("id %d", id)
}

// checkFit : textes de la fenêtre principale qui ne tiennent pas dans leur place (unités : pixels).
func (a *app) checkFit(label string) {
	var bad []string
	w := func(id int) int32 { r := childRect(a.hwnd, a.ctl[id]); return r.r - r.l }
	h := func(id int) int32 { r := childRect(a.hwnd, a.ctl[id]); return r.b - r.t }
	fontOf := func(id int) uintptr { f, _, _ := pSendMessageW.Call(a.ctl[id], wmGetFont, 0, 0); return f }
	text := func(id int) string { return getText(a.ctl[id]) }
	for _, id := range []int{idHelp, idDatBrowse, idBkBrowse, idOpen, idConvert, idQuit, idDonate, idSuggest} {
		if need := int32(a.measurePx(fontOf(id), text(id))) + int32(a.s(12)); need > w(id) {
			bad = append(bad, fmt.Sprintf("bouton %s : %d > %d", ctlName(id), need, w(id)))
		}
	}
	for _, k := range []string{"btn.themeLight", "btn.themeDark"} { // les deux libellés, icône comprise
		if need := int32(a.measurePx(fontOf(idTheme), T(k))) + int32(a.s(12+14+7)); need > w(idTheme) {
			bad = append(bad, fmt.Sprintf("bouton %q : %d > %d", T(k), need, w(idTheme)))
		}
	}
	for _, id := range []int{idLblMachine, idLblFolder, idLblSort, idLblLaser, idDatInfo} {
		if need := int32(a.measurePx(fontOf(id), text(id))); need > w(id) {
			bad = append(bad, fmt.Sprintf("libellé %q : %d > %d", text(id), need, w(id)))
		}
	}
	for _, id := range []int{idRecursive, idAir, idTop} {
		need := int32(a.measurePx(fontOf(id), text(id))) + int32(a.s(24))
		if id == idAir {
			need += int32(a.measurePx(a.fontSmall, T("lbl.required"))) + int32(a.s(26))
		}
		if need > w(id) && 2*textHeight(a.hwnd, fontOf(id)) > h(id) {
			bad = append(bad, fmt.Sprintf("case %q : %d > %d", text(id), need, w(id)))
		}
	}
	for _, c := range a.cards {
		if need := a.measure(a.fontSection, c.title) + 16 + 22 + 10 + 16; need > c.x1-c.x0 {
			bad = append(bad, fmt.Sprintf("titre %q : %d > %d", c.title, need, c.x1-c.x0))
		}
	}
	if need := a.titleLeft + a.measure(a.fontTitle, T("hdr.title")); need > a.titleRight {
		bad = append(bad, fmt.Sprintf("titre de la fenêtre : %d > %d", need, a.titleRight))
	}
	for _, id := range []int{idSort, idLaser, idLang} {
		n := int(send(a.ctl[id], cbGetCount, 0, 0))
		for i := 0; i < n; i++ {
			l := send(a.ctl[id], cbGetLBTextLen, uintptr(i), 0)
			buf := make([]uint16, l+1)
			send(a.ctl[id], cbGetLBText, uintptr(i), uintptr(unsafe.Pointer(&buf[0])))
			s := syscall.UTF16ToString(buf)
			extra := int32(a.s(46))
			if id == idLang {
				extra += int32(a.s(28))
			}
			if need := int32(a.measurePx(fontOf(id), s)) + extra; need > w(id) {
				bad = append(bad, fmt.Sprintf("liste %q : %d > %d", s, need, w(id)))
			}
		}
	}
	var rc rect
	pGetClientRect.Call(a.hwnd, uintptr(unsafe.Pointer(&rc)))
	for id, hw := range a.ctl {
		if r := childRect(a.hwnd, hw); isVisible(hw) && (r.r > rc.r || r.b > rc.b || r.l < 0 || r.t < 0) && id != idMachine && id != idSort && id != idLaser && id != idLang {
			bad = append(bad, fmt.Sprintf("contrôle %s hors fenêtre", ctlName(id)))
		}
	}
	ids := []int{idHelp, idTheme, idLang, idDatEdit, idDatBrowse, idRecursive, idDatInfo, idBkEdit, idBkBrowse, idLblMachine, idMachine,
		idLblFolder, idFolderName, idAir, idLblSort, idSort, idTop, idLblLaser, idLaser, idLog, idOpen, idConvert, idQuit, idDonate, idSuggest, idUpdate}
	vis := func(id int) (rect, bool) {
		r := childRect(a.hwnd, a.ctl[id])
		if f, ok := a.frames[id]; ok { // champ : son cadre dessiné
			r = f
		}
		return r, isVisible(a.ctl[id])
	}
	for i, x := range ids {
		for _, y := range ids[i+1:] {
			rx, vx := vis(x)
			ry, vy := vis(y)
			if vx && vy && rx.l < ry.r && ry.l < rx.r && rx.t < ry.b && ry.t < rx.b {
				bad = append(bad, fmt.Sprintf("chevauchement %s / %s", ctlName(x), ctlName(y)))
			}
		}
	}
	if len(bad) == 0 {
		kb(label + " : aucun débordement ni chevauchement")
		return
	}
	for _, b := range bad {
		kb(label + " : DÉBORDEMENT " + b)
	}
}

var pIsWindowVisible = user32.NewProc("IsWindowVisible")

func isVisible(h uintptr) bool { r, _, _ := pIsWindowVisible.Call(h); return r != 0 }

const cbGetCount = 0x0146

// fillDemo : données d'exemple (chemins longs, journal de tous les types, progression réelle simulée).
func (a *app) fillDemo(dat, folder string) {
	setText(a.ctl[idDatEdit], dat)
	setText(a.ctl[idBkEdit], `C:\Users\Atelier\Documents\Sauvegardes Job Manager\EpilogDatabase_Backup_10-22-2026 - atelier.epidb`)
	setText(a.ctl[idFolderName], folder)
	setText(a.ctl[idDatInfo], Tf("info.count", 3125))
	send(a.ctl[idMachine], cbResetContent, 0, 0)
	for _, k := range []string{"Fusion 40 Fibre 50 watts – atelier principal (laser n° 2)", "MINI24 40 watts"} {
		send(a.ctl[idMachine], cbAddString, 0, wp(k))
	}
	send(a.ctl[idMachine], cbSetCurSel, 0, 0)
	a.clearLog()
	a.log(Tf("log.backupLoaded", 2))
	a.log("")
	a.log(Tf("log.convStart", folder, "MINI24 40 watts"))
	a.log(T("log.copy"))
	a.log(Tf("log.read", 112))
	a.log(Tf("log.skipped", "Essai.DAT", T("skip.testGridLog")))
	a.log(Tf("log.note", "Bois 3 mm.DAT", "Tampon"))
	a.log(Tf("log.warnings", 3))
	a.log(Tf("log.done", 111, 1))
	a.log(T("log.outDir") + dat + T("out.suffix"))
	a.log(T("log.fail") + "exemple d'erreur\nsur deux lignes")
	a.setProgress(62)
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

// uiState : tout ce qu'une bascule de thème ne doit pas changer.
func (a *app) uiState() string {
	var b strings.Builder
	for _, id := range []int{idDatEdit, idBkEdit, idFolderName, idDatInfo} {
		fmt.Fprintf(&b, "%s=%q;", ctlName(id), getText(a.ctl[id]))
	}
	for _, id := range []int{idMachine, idSort, idLaser, idLang} {
		fmt.Fprintf(&b, "%s=%d;", ctlName(id), int32(send(a.ctl[id], cbGetCurSel, 0, 0)))
	}
	for _, id := range []int{idRecursive, idAir, idTop} {
		fmt.Fprintf(&b, "%s=%v;", ctlName(id), a.checked(id))
	}
	for _, id := range []int{idOpen, idConvert, idDatEdit, idAir} {
		fmt.Fprintf(&b, "%s actif=%v;", ctlName(id), isEnabled(a.ctl[id]))
	}
	// texte du rapport (logBuf) et texte réellement affiché dans le journal
	fmt.Fprintf(&b, "progression=%d;journal=%q;affiché=%q;entrées=%d", a.progress, a.logText(), getText(a.ctl[idLog]), len(a.logEntries))
	return b.String()
}

// toggleSteps : bascules répétées (sans enregistrer le choix) avec fenêtre remplie, notice
// ouverte et contrôles verrouillés ; l'état doit rester identique, le journal recoloré.
func (a *app) toggleSteps() []func() {
	var before string
	var helpPos point
	flip := func() { setPaletteDark(!isDark); a.applyTheme() }
	check := func(label string) func() {
		return func() {
			if after := a.uiState(); after == before {
				kb(label + " : état inchangé (champs, options, journal, progression)")
			} else {
				kb(label + " : ÉTAT MODIFIÉ\n  avant " + before + "\n  après " + after)
			}
		}
	}
	steps := []func(){
		func() {
			kb("Bascules de thème :")
			a.fillDemo(`C:\Exemple\Réglages Epilog\Helix 75 watts`, "Helix")
			send(a.ctl[idTop], bmSetCheck, 0, 0)
			send(a.ctl[idSort], cbSetCurSel, 2, 0)
			before = a.uiState()
		},
	}
	for i := 0; i < 10; i++ {
		steps = append(steps, flip)
	}
	steps = append(steps, check("10 bascules, fenêtre remplie"),
		func() { shot(a.hwnd, "bascule-fenetre") },
		func() { a.showHelp() },
		func() { send(helpEdit, 0x0115, 3, 0); send(helpEdit, emGetScrollPos, 0, uintptr(unsafe.Pointer(&helpPos))) }, // une page plus bas
	)
	for i := 0; i < 5; i++ {
		steps = append(steps, flip)
	}
	steps = append(steps,
		func() {
			var p point
			send(helpEdit, emGetScrollPos, 0, uintptr(unsafe.Pointer(&p)))
			kb(fmt.Sprintf("5 bascules, notice ouverte : position de lecture %d -> %d, notice %v", helpPos.y, p.y, map[bool]string{true: "toujours ouverte", false: "FERMÉE"}[helpHwnd != 0]))
			shot(helpHwnd, "bascule-notice")
		},
		check("5 bascules, notice ouverte"),
		func() { pPostMessageW.Call(helpHwnd, wmClose, 0, 0) },
		func() { // comme pendant une conversion : contrôles verrouillés, curseur occupé
			for _, id := range lockedIDs {
				enable(a.ctl[id], false)
			}
			a.busy = true
			before = a.uiState()
		},
	)
	for i := 0; i < 5; i++ {
		steps = append(steps, flip)
	}
	steps = append(steps,
		func() {
			kb(fmt.Sprintf("5 bascules, contrôles verrouillés : bouton de thème actif=%v", isEnabled(a.ctl[idTheme])))
			shot(a.hwnd, "bascule-verrouille")
		},
		check("5 bascules, contrôles verrouillés"),
		func() {
			a.busy = false
			for _, id := range lockedIDs {
				enable(a.ctl[id], true)
			}
		},
	)
	return steps
}
