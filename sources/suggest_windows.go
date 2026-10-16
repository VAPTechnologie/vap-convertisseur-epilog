//go:build windows

package main

// Bouton « Suggestion / bug » et envoi du rapport d'erreur. Pas de lien
// mailto : sans logiciel de messagerie (Gmail dans le navigateur), Windows
// demande une application et rien ne s'ouvre. L'adresse est copiée dans le
// presse-papiers et un message explique quoi envoyer ; après un échec,
// l'explorateur s'ouvre sur le rapport à joindre.

import (
	"path/filepath"
	"strings"
	"unsafe"
)

var (
	pOpenClipboard    = user32.NewProc("OpenClipboard")
	pEmptyClipboard   = user32.NewProc("EmptyClipboard")
	pSetClipboardData = user32.NewProc("SetClipboardData")
	pCloseClipboard   = user32.NewProc("CloseClipboard")
	pGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	pGlobalLock       = kernel32.NewProc("GlobalLock")
	pGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
)

func appVersion() string { return strings.TrimSpace(strings.TrimPrefix(appSignature, "VAP")) }

func copyToClipboard(owner uintptr, s string) bool {
	u := w16(s)
	n := 0
	for p := u; *p != 0; p = (*uint16)(unsafe.Add(unsafe.Pointer(p), 2)) {
		n++
	}
	size := uintptr(n+1) * 2
	if r, _, _ := pOpenClipboard.Call(owner); r == 0 {
		return false
	}
	defer pCloseClipboard.Call()
	pEmptyClipboard.Call()
	h, _, _ := pGlobalAlloc.Call(0x2, size) // GMEM_MOVEABLE
	if h == 0 {
		return false
	}
	dst, _, _ := pGlobalLock.Call(h)
	if dst == 0 {
		return false
	}
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(dst)), n+1), unsafe.Slice(u, n+1))
	pGlobalUnlock.Call(h)
	r, _, _ := pSetClipboardData.Call(13, h) // CF_UNICODETEXT
	return r != 0
}

// suggestText : message affiché, adresse en tête, puis ce qu'il faut écrire.
func suggestText(report string) string {
	lines := []string{Tf("msg.sendMail", suggestMail), ""}
	if report != "" {
		lines = append(lines, Tf("msg.attachReport", filepath.Base(report)), "")
	}
	lines = append(lines,
		"• "+T("mail.q1"),
		"• "+T("mail.q2"),
		"• "+T("mail.q3"),
		"",
		"Version : "+appVersion()+" · "+langCodes[cur],
		"",
		T("mail.gh")+"\n"+issuesURL,
	)
	return strings.Join(lines, "\n")
}
