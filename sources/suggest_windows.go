//go:build windows

package main

// Bouton « Suggestion / bug » : ouvre la messagerie avec un e-mail déjà préparé
// (version, langue, questions), qui rappelle aussi la page GitHub du projet.

import (
	"net/url"
	"strings"
)

func suggestMailto() string {
	ver := strings.TrimSpace(strings.TrimPrefix(appSignature, "VAP"))
	subject := "VAP Convertisseur " + ver + " - suggestion / bug"
	body := strings.Join([]string{
		T("mail.q1"), "", "",
		T("mail.q2"), "", "",
		T("mail.q3"), "",
		"----",
		"Version : " + ver + " · " + langCodes[cur],
		T("mail.gh") + " " + issuesURL,
	}, "\r\n")
	esc := func(s string) string { return strings.ReplaceAll(url.QueryEscape(s), "+", "%20") }
	return "mailto:" + suggestMail + "?subject=" + esc(subject) + "&body=" + esc(body)
}
