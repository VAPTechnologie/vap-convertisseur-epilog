package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// Faux serveur local : aucun appel réseau réel.
func TestLatestRelease(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/repos/x/releases/latest" ||
			r.Header.Get("User-Agent") != "VAP-Convertisseur-Epilog" || r.Header.Get("Accept") != "application/vnd.github+json" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Write([]byte(`{"tag_name":"v1.6","html_url":"` + releasesURL + `/tag/v1.6","body":"..."}`))
	}))
	defer srv.Close()
	tag, page, err := LatestRelease(srv.URL + "/repos/x/releases/latest")
	if err != nil || tag != "v1.6" || page != releasesURL+"/tag/v1.6" {
		t.Fatalf("%q %q %v", tag, page, err)
	}

	limited := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer limited.Close()
	if _, _, err := LatestRelease(limited.URL); err == nil {
		t.Error("limite de requêtes : aucune erreur")
	}

	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"tag_name":"v1.6","html_url":"https://exemple.invalid/"}`))
	}))
	defer other.Close()
	if _, page, _ := LatestRelease(other.URL); page != releasesURL+"/latest" {
		t.Errorf("page hors projet acceptée : %q", page)
	}

	big := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"tag_name":"v1.6","body":"` + strings.Repeat("x", 2<<20) + `"}`))
	}))
	defer big.Close()
	if _, _, err := LatestRelease(big.URL); err == nil {
		t.Error("réponse de plus de 1 Mo acceptée")
	}

	srv.Close()
	if _, _, err := LatestRelease(srv.URL); err == nil {
		t.Error("serveur arrêté : aucune erreur")
	}
	if _, _, err := LatestRelease("ftp://exemple.invalid/"); err == nil {
		t.Error("schéma ftp accepté")
	}
}

// Vraie requête à GitHub, seulement avec VAP_RESEAU=1.
func TestLatestReleaseGitHub(t *testing.T) {
	if os.Getenv("VAP_RESEAU") == "" {
		t.Skip("VAP_RESEAU non défini")
	}
	tag, page, err := LatestRelease(latestURL)
	if err != nil || versionNums(tag) == nil || !strings.HasPrefix(page, releasesURL+"/") {
		t.Fatalf("%q %q %v", tag, page, err)
	}
	t.Logf("dernière release : %s %s", tag, page)
}
