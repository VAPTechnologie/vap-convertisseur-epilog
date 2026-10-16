package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// schemaDB : fausse base qui ne répond qu'à PRAGMA table_info.
type schemaDB map[string][]string

func (s schemaDB) Query(sql string, args ...any) ([][]any, error) {
	t := strings.TrimSuffix(strings.TrimPrefix(sql, "PRAGMA table_info("), ")")
	if t == sql {
		return nil, errors.New("requête inattendue")
	}
	var rows [][]any
	for i, c := range s[t] {
		rows = append(rows, []any{int64(i), c, "TEXT"})
	}
	return rows, nil
}
func (s schemaDB) Exec(sql string, args ...any) error { return errors.New("lecture seule") }
func (s schemaDB) Close() error                       { return nil }

func fullSchema() schemaDB {
	s := schemaDB{}
	for _, t := range requiredSchema {
		s[t.Table] = append([]string{}, t.Cols...)
	}
	return s
}

func TestCheckSchema(t *testing.T) {
	cur = LFR
	if err := CheckSchema(fullSchema()); err != nil {
		t.Fatal(err)
	}
	s := fullSchema()
	s["Settings"] = s["Settings"][:0:0]
	for _, c := range fullSchema()["Settings"] {
		if c != "Category" {
			s["Settings"] = append(s["Settings"], strings.ToLower(c))
		}
	}
	delete(s, "MaterialSubfolders")
	err := CheckSchema(s)
	if err == nil {
		t.Fatal("aucune erreur")
	}
	for _, w := range []string{"Settings.Category", "MaterialSubfolders", "Refaites une sauvegarde"} {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("%q absent de %q", w, err)
		}
	}
	if strings.Contains(err.Error(), "Settings.ID") {
		t.Errorf("colonne présente signalée : %q", err)
	}
	if _, err := Convert(s, Options{DatDir: t.TempDir()}, func(string) {}, nil); err == nil || !strings.Contains(err.Error(), "Settings.Category") {
		t.Errorf("Convert : %v", err)
	}
}

func TestSourceDirFor(t *testing.T) {
	cur = LFR
	root := t.TempDir()
	mk := func(dir string, dat bool) string {
		p := filepath.Join(root, dir)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		if dat {
			os.WriteFile(filepath.Join(p, "Bois.DAT"), []byte{0}, 0o644)
		}
		return p
	}
	src := mk("Helix 75 watts", true)
	out := mk("Helix 75 watts - Conversion Job Manager", false)
	out2 := mk("Helix 75 watts - Conversion Job Manager (2)", false)
	outEN := mk("Helix 75 watts - Job Manager conversion (3)", false)
	orphan := mk("Zing - Conversion Job Manager", false)
	other := mk("Divers", false)
	full := mk("Mini - Conversion Job Manager", true)
	mk("Mini", true)

	for _, c := range []struct{ dir, want string }{
		{out, src}, {out2, src}, {outEN, src}, {src, ""}, {orphan, ""}, {other, ""}, {full, ""},
		{filepath.Join(root, "absent - Conversion Job Manager"), ""},
	} {
		if got := SourceDirFor(c.dir, false); got != c.want {
			t.Errorf("%s : %q, attendu %q", filepath.Base(c.dir), got, c.want)
		}
	}
}
