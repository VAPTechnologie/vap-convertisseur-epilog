//go:build windows

package main

import "testing"

// Type des lignes du journal dans les 9 langues : le texte n'est jamais modifié,
// seule la couleur dépend du type.
func TestLogKind(t *testing.T) {
	defer func(l Lang) { cur = l }(cur)
	for l := Lang(0); l < nLang; l++ {
		cur = l
		cases := []struct {
			s    string
			want int
		}{
			{T("welcome.0"), logHead},
			{T("welcome.1"), logInfo},
			{"", logInfo},
			{Tf("log.backupLoaded", 3), logInfo},
			{Tf("log.convStart", "Mini 24", "MINI24 40 watts"), logHead},
			{T("log.copy"), logInfo},
			{Tf("log.read", 12), logInfo},
			{Tf("log.skipped", "a.DAT", "x"), logWarn},
			{Tf("log.note", "a.DAT", "x"), logNote},
			{Tf("log.warnings", 4), logWarn},
			{"  ⚠ " + "remplacée", logWarn},
			{Tf("log.done", 10, 2), logOK},
			{T("log.outDir") + `C:\x`, logOK},
			{T("log.fail") + "erreur\nsur deux lignes", logError},
			{T("log.report") + `C:\rapport.txt`, logError},
			{T("log.recapFail") + "x", logWarn},
			{Tf("log.update", "V1.9", "https://github.com/x"), logLink},
		}
		for _, c := range cases {
			if got := logKindOf(c.s); got != c.want {
				t.Errorf("[%s] %q : type %d, attendu %d", langCodes[l], c.s, got, c.want)
			}
		}
	}
}

// Le tampon du journal reproduit exactement l'ancien champ texte (texte du rapport d'erreur) :
// lignes séparées par CR LF, pas de séparateur avant la première ligne non vide.
func TestLogBuffer(t *testing.T) {
	a := &app{ctl: map[int]uintptr{}} // sans fenêtre : seuls les appels Windows échouent sans effet
	a.clearLog()
	for _, s := range []string{"", "un", "", "deux\nlignes", "trois"} {
		a.log(s)
	}
	if want := "un\r\n\r\ndeux\nlignes\r\ntrois"; a.logText() != want {
		t.Fatalf("%q, attendu %q", a.logText(), want)
	}
	a.clearLog()
	a.log("x")
	if a.logText() != "x" {
		t.Fatalf("après effacement : %q", a.logText())
	}
}
