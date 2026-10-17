package main

import (
	"regexp"
	"sort"
	"strings"
	"testing"
)

var verbRe = regexp.MustCompile(`%(\[\d\])?[-0-9]*[dsvwX]`)
var idxRe = regexp.MustCompile(`\[\d\]`)

func verbs(s string) string {
	v := verbRe.FindAllString(s, -1)
	for i := range v {
		v[i] = idxRe.ReplaceAllString(v[i], "")
	}
	sort.Strings(v)
	return strings.Join(v, ",")
}

func TestVerbs(t *testing.T) {
	for k, v := range tr {
		ref := verbs(v[LFR])
		for l := Lang(0); l < nLang; l++ {
			if v[l] == "" {
				continue
			}
			if verbs(v[l]) != ref {
				t.Errorf("%s [%s]: %q vs %q", k, langCodes[l], verbs(v[l]), ref)
			}
		}
		if k == "csv.cols" {
			for l := Lang(0); l < nLang; l++ {
				if n := strings.Count(v[l], "|"); n != 11 {
					t.Errorf("csv.cols [%s] %d cols", langCodes[l], n+1)
				}
			}
		}
	}
	for l := Lang(0); l < nLang; l++ {
		if n := len(verbRe.FindAllString(readmes[l], -1)); n != 6 {
			t.Errorf("readme [%s] %d verbs", langCodes[l], n)
		}
		if strings.Contains(notices[l], "%") {
			t.Errorf("notice [%s] contains %%", langCodes[l])
		}
	}
}
