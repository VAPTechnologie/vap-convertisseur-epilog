//go:build windows

package main

// Nouvelle version : la recherche tourne dans une goroutine, qui poste wmUpdate
// à la fenêtre ; l'interface n'est modifiée que dans wndProc.

import "strings"

const (
	wmUpdate   = 0x8000 + 1 // WM_APP + 1
	ssNotify   = 0x0100
	idcHand    = 32649
	stnClicked = 0
)

var pPostMessageW = user32.NewProc("PostMessageW")

type release struct{ tag, page string }

var updateCh = make(chan release, 1)

func (a *app) checkUpdate() {
	hwnd := a.hwnd
	go func() {
		tag, page, err := LatestRelease(latestURL)
		if err != nil || !NewerVersion(tag, appVersion()) {
			return
		}
		updateCh <- release{"V" + strings.TrimLeft(tag, "vV"), page}
		pPostMessageW.Call(hwnd, wmUpdate, 0, 0)
	}()
}

// onUpdate : sur le fil de la fenêtre.
func (a *app) onUpdate() {
	select {
	case r := <-updateCh:
		a.newRelease = r
	default:
		return
	}
	a.applyUpdateText()
	a.layout()
	if a.welcomeOnly {
		a.showWelcome()
	} else {
		a.log(a.updateLogLine())
	}
}

func (a *app) updateLogLine() string {
	return Tf("log.update", a.newRelease.tag, a.newRelease.page)
}

// applyUpdateText : texte du lien du pied de fenêtre, visible seulement s'il y a une version plus récente.
func (a *app) applyUpdateText() {
	h := a.ctl[idUpdate]
	if a.newRelease.tag == "" {
		pShowWindow.Call(h, 0)
		return
	}
	setText(h, Tf("upd.link", a.newRelease.tag))
	pShowWindow.Call(h, 5)
}
