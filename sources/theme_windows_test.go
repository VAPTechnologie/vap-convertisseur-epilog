//go:build windows

package main

import (
	"math"
	"strings"
	"testing"
)

// luminance relative (WCAG 2.x) d'une couleur COLORREF (0x00BBGGRR)
func luminance(c uintptr) float64 {
	ch := func(v uintptr) float64 {
		x := float64(v&0xFF) / 255
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*ch(c) + 0.7152*ch(c>>8) + 0.0722*ch(c>>16)
}

func contrast(a, b uintptr) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// Contrastes mesurés : 4,5:1 pour le texte, 3:1 pour les contours et indicateurs utiles.
func testContrastes(t *testing.T) {
	cases := []struct {
		nom      string
		fg, bg   uintptr
		min      float64
	}{
		{"texte sur panneau", colTexte, colPanneau, 4.5},
		{"libellés secondaires (Machine, Nom du dossier, Source laser, compteur)", colDiscret, colPanneau, 4.5},
		{"sous-titre de l'en-tête", colDiscret, colFond, 4.5},
		{"texte d'un contrôle verrouillé (fond sombre)", colInactif, colFond, 4.5},
		{"texte d'un bouton désactivé", colInactif, colFond, 4.5},
		{"texte blanc de « Convertir »", colTexteAccent, colPrincipal, 4.5},
		{"texte blanc de « Convertir » enfoncé", colTexteAccent, colPrincipalOn, 4.5},
		{"texte d'un bouton secondaire", colTexte, colDoux, 4.5},
		{"texte d'un bouton survolé", colTexte, colSurvol, 4.5},
		{"liens (don, suggestion)", colAccentClair, colPanneau, 4.5},
		{"journal : vert (réussite)", colVert, colPanneau, 4.5},
		{"journal : orange (avertissement)", colOrange, colPanneau, 4.5},
		{"journal : rouge (erreur)", colRouge, colPanneau, 4.5},
		{"journal : bleu (titres)", colAccentClair, colPanneau, 4.5},
		{"badge « Obligatoire »", colTexte, colDoux, 4.5},
		{"contour des champs et listes", colBord, colPanneau, 3},
		{"contour d'une case non cochée", colBordSurvol, colPanneau, 3},
		{"contour au focus", colAccent, colPanneau, 3},
		{"contour d'un bouton secondaire", colBord, colPanneau, 3},
		{"case cochée", colAccent, colPanneau, 3},
		{"curseur de la barre de défilement", colBord, colDoux, 3},
		{"barre de progression", colAccent, colDoux, 3},
		{"coche blanche dans la case", colTexteAccent, colAccent, 3},
		{"coche d'une case verrouillée", colDiscret, colImpose, 3},
	}
	for _, c := range cases {
		r := contrast(c.fg, c.bg)
		t.Logf("[%s] %-70s %5.2f:1 (minimum %.1f)", themeName(), c.nom, r, c.min)
		if r < c.min {
			t.Errorf("[%s] %s : %.2f:1 < %.1f:1", themeName(), c.nom, r, c.min)
		}
	}
}

func themeName() string {
	if curPal == &lightPal {
		return "clair"
	}
	return "sombre"
}

// Les deux thèmes, et le bandeau (toujours sombre) avec ses boutons.
func TestContrastesDeuxThemes(t *testing.T) {
	defer usePalette(curPal)
	for _, p := range []*palette{&darkPal, &lightPal} {
		usePalette(p)
		t.Run(themeName(), testContrastes)
	}
}

// Nouveaux libellés présents dans les 9 langues.
func TestNouveauxLibelles(t *testing.T) {
	for _, k := range []string{"btn.yes", "btn.no", "btn.ok", "btn.replace", "btn.cancel", "lbl.required", "btn.themeLight", "btn.themeDark"} {
		for l := Lang(0); l < nLang; l++ {
			if tr[k][l] == "" {
				t.Errorf("%s [%s] manquant", k, langCodes[l])
			}
		}
	}
}

// Message de fin : résultat, bilan, consignes, question ; aucun texte perdu, dans les 9 langues.
func TestMessageDeFin(t *testing.T) {
	defer func(l Lang) { cur = l }(cur)
	for l := Lang(0); l < nLang; l++ {
		cur = l
		done := Tf("msg.done", 111, "Mini 24")
		sum := Tf("msg.summary", 108, 3, 1)
		got := withSummary(done, sum)
		parts := strings.Split(done, "\n\n")
		if len(parts) < 4 {
			t.Fatalf("[%s] msg.done : %d paragraphes", langCodes[l], len(parts))
		}
		bilan := strings.TrimPrefix(sum, "\n\n")
		iRes, iBil := strings.Index(got, parts[1]), strings.Index(got, bilan)
		iCons, iQ := strings.Index(got, parts[2]), strings.LastIndex(got, parts[len(parts)-1])
		if !(iRes >= 0 && iRes < iBil && iBil < iCons && iCons < iQ) {
			t.Errorf("[%s] ordre incorrect :\n%s", langCodes[l], got)
		}
		if len(got) != len(done)+len(sum) {
			t.Errorf("[%s] texte modifié (%d au lieu de %d caractères)", langCodes[l], len(got), len(done)+len(sum))
		}
		if !strings.HasSuffix(got, parts[len(parts)-1]) {
			t.Errorf("[%s] la question n'est plus à la fin", langCodes[l])
		}
	}
}
