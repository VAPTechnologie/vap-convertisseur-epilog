package main

import "testing"

func TestNewerVersion(t *testing.T) {
	for _, c := range []struct {
		latest, current string
		want            bool
	}{
		{"v1.6", "1.5", true},
		{"v1.10", "1.9", true},
		{"v1.9", "1.10", false},
		{"v1.5", "1.5", false},
		{"v1.5.1", "1.5", true},
		{"v1.5", "1.5.0", false},
		{"v2.0", "1.12", true},
		{"v1.4", "1.5", false},
		{"V1.6", "VAP  V1.5", true},
		{"", "1.5", false},
		{"nightly", "1.5", false},
		{"v1.6", "", false},
	} {
		if got := NewerVersion(c.latest, c.current); got != c.want {
			t.Errorf("NewerVersion(%q, %q) = %v", c.latest, c.current, got)
		}
	}
}

func TestParseRelease(t *testing.T) {
	tag, page, err := parseRelease([]byte(`{"tag_name":"v1.6","html_url":"` + releasesURL + `/tag/v1.6","body":"..."}`))
	if err != nil || tag != "v1.6" || page != releasesURL+"/tag/v1.6" {
		t.Fatalf("%q %q %v", tag, page, err)
	}
	if _, page, _ := parseRelease([]byte(`{"tag_name":"v1.6","html_url":"https://exemple.invalid/"}`)); page != releasesURL+"/latest" {
		t.Errorf("page hors projet acceptée : %q", page)
	}
	if _, _, err := parseRelease([]byte(`<html>`)); err == nil {
		t.Error("réponse non JSON acceptée")
	}
}
