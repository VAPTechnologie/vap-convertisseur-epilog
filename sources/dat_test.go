package main

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"
)

// fakeDat : .DAT synthétique (en-tête, marqueur MDPE, valeurs principales)
// à la taille du pilote donné.
func fakeDat(size int, header string) []byte {
	b := make([]byte, size)
	for i, u := range utf16.Encode([]rune(header)) {
		binary.LittleEndian.PutUint16(b[i*2:], u)
	}
	copy(b[220:], "MDPE")
	put := func(off, v int) { binary.LittleEndian.PutUint32(b[off:], uint32(int32(v))) }
	put(256, 50)                                       // puissance gravure
	put(260, 40)                                       // vitesse gravure
	put(264, 80)                                       // puissance découpe
	put(268, 20)                                       // vitesse découpe
	put(272, 2)                                        // combiné
	put(836, 50)                                       // fréquence (50 Hz Mini/Zing, 50 % Fusion)
	binary.LittleEndian.PutUint16(b[90:], uint16(600)) // dpi
	return b
}

func writeFake(t *testing.T, name string, b []byte) string {
	p := filepath.Join(t.TempDir(), name+".DAT")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestParseDatFormats(t *testing.T) {
	cur = LFR
	for _, c := range []struct {
		size   int
		header string
		layout string
	}{
		{2924, "Epilog Engraver Win32 Mini", "mini"},
		{2924, "Epilog Engraver WinX64", "mini"}, // en-tête réel (Mini 24, plateau 24 × 12) : sans nom de modèle
		{2992, "Epilog Engraver Win32 Zing", "zing"},
		{3156, "Epilog Engraver Win32 Fusion", "fusion"},
	} {
		d, err := ParseDat(writeFake(t, c.layout, fakeDat(c.size, c.header)))
		if err != nil {
			t.Errorf("%s : %v", c.layout, err)
			continue
		}
		if !d.Known() || d.Layout.Name != c.layout {
			t.Errorf("%s : structure %v", c.layout, d.Layout)
		}
	}
	// tailles ou combinaisons non relevées sur les pilotes d'origine : refusées
	for _, c := range []struct {
		size   int
		header string
	}{
		{1478, "Epilog Engraver Win32 Mini"},   // coupé en deux
		{3000, "Epilog Engraver Win32 Mini"},   // taille inconnue
		{2924, "Epilog Engraver Win32 Zing"},   // en-tête Zing, taille Mini
		{3156, "Epilog Engraver Win32 Zing"},   // en-tête Zing, taille Fusion
		{2992, "Epilog Engraver Win32 Fusion"}, // en-tête Fusion, taille Zing
	} {
		_, err := ParseDat(writeFake(t, "x", fakeDat(c.size, c.header)))
		if err == nil || !strings.Contains(err.Error(), "non reconnue") {
			t.Errorf("%d octets %q : erreur attendue, obtenu %v", c.size, c.header, err)
		}
	}
}

func TestCheckLaser(t *testing.T) {
	cur = LFR
	co2 := Machine{Name: "Fusion CO2", LaserType: `<list><element val="CO2"/></list>`}
	fib := Machine{Name: "Fusion Fibre", LaserType: `<list><element val="Fiber"/></list>`}
	dual := Machine{Name: "Fusion Dual", LaserType: `<list><element val="CO2"/><element val="Fiber"/></list>`}
	for _, c := range []struct {
		m     Machine
		fiber bool
		ok    bool
	}{
		{co2, false, true}, {co2, true, false},
		{fib, true, true}, {fib, false, false},
		{dual, false, true}, {dual, true, true},
	} {
		err := CheckLaser(Options{Machine: c.m, Fiber: c.fiber})
		if (err == nil) != c.ok {
			t.Errorf("%s fibre=%v : %v", c.m.Name, c.fiber, err)
		}
	}
}

func TestNeedsFiber(t *testing.T) {
	b := fakeDat(3156, "Epilog Engraver Win32 Fusion")
	d, err := ParseDat(writeFake(t, "co2", b))
	if err != nil || d.NeedsFiber() {
		t.Fatalf("Fusion CO2 : %v %v", d.NeedsFiber(), err)
	}
	binary.LittleEndian.PutUint32(b[fusLaser:], 1)
	d, err = ParseDat(writeFake(t, "fibre", b))
	if err != nil || !d.NeedsFiber() {
		t.Fatalf("Fusion Fibre : %v %v", d.NeedsFiber(), err)
	}
	if !(DatFile{Colors: []ColorEntry{{}, {Fiber: true}}}).NeedsFiber() {
		t.Error("couleur Fibre non détectée")
	}
}

func TestParseDatRemarques(t *testing.T) {
	cur = LFR
	put := func(b []byte, off, v int) []byte {
		c := append([]byte(nil), b...)
		binary.LittleEndian.PutUint32(c[off:], uint32(int32(v)))
		return c
	}
	fus := fakeDat(3156, "Epilog Engraver Win32 Fusion")
	put16 := func(b []byte, off, v int) []byte {
		c := append([]byte(nil), b...)
		binary.LittleEndian.PutUint16(c[off:], uint16(int16(v)))
		return c
	}
	// une couleur : rouge, puissance 60, vitesse 30, fréquence 50 %, découpe
	color := func(b []byte, power, speed, freq int) []byte {
		c := put(b, 792, 1)
		o := colorTable
		c[o], c[o+3] = 0xff, byte(power)
		c = put(c, o+4, speed)
		c = put(c, o+8, freq)
		return put(c, o+20, 1)
	}
	for _, c := range []struct {
		name string
		b    []byte
		want string // début de la remarque attendue ("" : aucune)
	}{
		{"normal", fus, ""},
		{"resolution 0", put16(fus, 90, 0), "résolution 0 invalide"},
		{"frequence 250", put(fus, 836, 250), "fréquence 250 hors plage"},
		{"page centre", put(put(fus, 788, 1), 2056, 3), "position « Page-Centre »"},
		{"position 9", put(put(fus, 788, 1), 2056, 9), "position du travail inconnue"},
		{"epaisseur negative", put16(fus, fusThickness, -300), "épaisseur autofocus négative"},
		{"Acrylique 2 passes", fus, "2 passes"},
		{"couleur frequence 300", color(fus, 60, 30, 300), "couleur #FF0000 : fréquence 300"},
	} {
		d, err := ParseDat(writeFake(t, c.name, c.b))
		if err != nil {
			t.Errorf("%s : %v", c.name, err)
			continue
		}
		got := strings.Join(d.Notes, " | ")
		if c.want == "" && got != "" || c.want != "" && !strings.HasPrefix(got, c.want) {
			t.Errorf("%s : remarques %q, attendu %q", c.name, got, c.want)
		}
	}
	if d, _ := ParseDat(writeFake(t, "r", put16(fus, 90, 0))); d.DPI != 600 {
		t.Errorf("résolution remplacée : %d", d.DPI)
	}
	// gestion des couleurs hors limites : réglage refusé
	for _, b := range [][]byte{color(fus, 150, 30, 50), color(fus, 60, 120, 50), color(fus, 60, -1, 50)} {
		if _, err := ParseDat(writeFake(t, "c", b)); err == nil || !strings.Contains(err.Error(), "hors limites") {
			t.Errorf("couleur hors limites acceptée : %v", err)
		}
	}
}
