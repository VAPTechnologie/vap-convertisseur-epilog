package main

// Recherche d'une version plus récente sur GitHub (dernière release publiée).
// Aucune donnée n'est envoyée ; en cas d'échec, rien n'est signalé.

import (
	"encoding/json"
	"strconv"
	"strings"
)

const (
	latestURL   = "https://api.github.com/repos/VAPTechnologie/vap-convertisseur-epilog/releases/latest"
	releasesURL = "https://github.com/VAPTechnologie/vap-convertisseur-epilog/releases"
)

// versionNums : "v1.10" ou "VAP  V1.5" -> [1 10] ; nil si aucun numéro.
func versionNums(s string) []int {
	s = strings.TrimSpace(s)
	if i := strings.LastIndexAny(s, "vV "); i >= 0 {
		s = s[i+1:]
	}
	var n []int
	for _, p := range strings.Split(s, ".") {
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		n = append(n, v)
	}
	return n
}

// NewerVersion : latest est-elle plus récente que current ? (1.10 > 1.9)
func NewerVersion(latest, current string) bool {
	l, c := versionNums(latest), versionNums(current)
	if l == nil || c == nil {
		return false
	}
	for i := 0; i < len(l) || i < len(c); i++ {
		var a, b int
		if i < len(l) {
			a = l[i]
		}
		if i < len(c) {
			b = c[i]
		}
		if a != b {
			return a > b
		}
	}
	return false
}

// parseRelease : tag et page de la release, depuis la réponse JSON de GitHub.
func parseRelease(body []byte) (tag, page string, err error) {
	var r struct {
		Tag  string `json:"tag_name"`
		Page string `json:"html_url"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return "", "", err
	}
	// on n'ouvrira que la page des releases du projet
	if !strings.HasPrefix(r.Page, releasesURL+"/") {
		r.Page = releasesURL + "/latest"
	}
	return r.Tag, r.Page, nil
}
