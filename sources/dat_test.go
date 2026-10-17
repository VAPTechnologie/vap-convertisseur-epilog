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
