//go:build windows

package main

// Bouton « Suggestion / bug » : ouvre la messagerie avec un e-mail déjà préparé
// (version, langue, questions), qui rappelle aussi la page GitHub du projet.
// Après un échec, l'e-mail demande de joindre le rapport d'erreur.

import (
	"net/url"
	"path/filepath"
	"strings"
)

func appVersion() string { return strings.TrimSpace(strings.TrimPrefix(appSignature, "VAP")) }

func suggestMailto(report string) string {
	ver := appVersion()
	subject := "VAP Convertisseur " + ver + " - suggestion / bug"
	lines := []string{T("mail.q1"), "", "", T("mail.q2"), "", ""}
	if report != "" {
		subject = "VAP Convertisseur " + ver + " - " + T("mail.reportSubject")
		lines = append(lines, Tf("mail.report", filepath.Base(report)), "")
	}
	lines = append(lines,
		T("mail.q3"), "",
		"----",
		"Version : "+ver+" · "+langCodes[cur],
		T("mail.gh")+" "+issuesURL,
	)
	body := strings.Join(lines, "\r\n")
	esc := func(s string) string { return strings.ReplaceAll(url.QueryEscape(s), "+", "%20") }
	return "mailto:" + suggestMail + "?subject=" + esc(subject) + "&body=" + esc(body)
}
