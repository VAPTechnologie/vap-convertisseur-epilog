package main

// Contrôle de non-régression : chaque .DAT d'un ou plusieurs dossiers est
// converti dans une copie de la sauvegarde, puis chaque valeur du réglage
// Job Manager obtenu est comparée au .DAT, décodé ici indépendamment du
// convertisseur (table des octets de « Decodage drivers Epilog.md »).
// Rien n'est écrit à côté des .DAT : tout se passe dans un dossier temporaire.
//
//   set VAP_REF=<dossier contenant les dossiers de .DAT>  (ex. « Réglages Epilog OLD »)
//   set VAP_BACKUP=<sauvegarde Job Manager .epidb / .db3>
//   go test -run TestVerifReglages -v
//
// La machine de destination est choisie d'après le format du .DAT :
// Mini / Helix -> machine dont le nom contient « mini », Zing -> « zing »,
// Fusion -> « fusion » (VAP_MINI, VAP_ZING, VAP_FUSION pour imposer un nom).

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"
)

type refLayout struct{ af, sortOff, sortMode, pos int }

var refLayouts = map[int]refLayout{
	2924: {844, 840, 2888, 1852},  // Mini, Helix
	2992: {-1, 840, 2956, 1920},   // Zing
	3156: {844, -1, -1, 2056},     // Fusion
}

var (
	reProc   = regexp.MustCompile(`<process ([^>]*)><query><container>(.*?)</container>`)
	reAttr   = regexp.MustCompile(`(\w+)="([^"]*)"`)
	reGlobal = regexp.MustCompile(`key="(\w+)" val="([^"]*)"`)
	reMMRef  = regexp.MustCompile(`(?i)(\d+(?:[.,]\d+)?)\s*mm`)
	reParRef = regexp.MustCompile(`\([^)]*\)`)
	rePasRef = regexp.MustCompile(`(?i)(\d+)\s*pass`)
)

type refDat struct {
	name string
	b    []byte
}

func (d refDat) i(off int) int   { return int(int32(binary.LittleEndian.Uint32(d.b[off:]))) }
func (d refDat) h(off int) int   { return int(int16(binary.LittleEndian.Uint16(d.b[off:]))) }
func (d refDat) fusion() bool    { return len(d.b) == 3156 }
func (d refDat) header() string {
	u := make([]uint16, 32)
	for k := range u {
		u[k] = binary.LittleEndian.Uint16(d.b[k*2:])
	}
	s := string(utf16.Decode(u))
	if k := strings.IndexRune(s, 0); k >= 0 {
		s = s[:k]
	}
	return s
}

// valid : le .DAT doit être converti (sinon il doit être ignoré).
func (d refDat) valid(m Machine) bool {
	if len(d.b) < 900 || !strings.HasPrefix(d.header(), "Epilog") || string(d.b[220:224]) != "MDPE" {
		return false
	}
	// seules les structures relevées sur les pilotes sont converties
	h, want := strings.ToLower(d.header()), 2924
	switch {
	case strings.Contains(h, "zing"):
		want = 2992
	case strings.Contains(h, "fusion"):
		want = 3156
	}
	if len(d.b) != want {
		return false
	}
	// réglage Fibre (général ou d'une couleur) vers une machine sans Fibre : refusé
	if d.needsFiber() && !m.HasFiber() {
		return false
	}
	if m := d.i(272); m < 0 || m > 2 {
		return false
	}
	for _, o := range []int{256, 260, 264, 268} {
		if v := d.i(o); v < 0 || v > 100 {
			return false
		}
	}
	low := foldKey(d.name)
	return !strings.HasPrefix(low, "param1") && !strings.HasPrefix(low, "param2")
}

func (d refDat) needsFiber() bool {
	if !d.fusion() {
		return false
	}
	if d.i(3096) != 0 {
		return true
	}
	if d.i(792) != 0 {
		for k := 0; k < 16; k++ {
			o := 1376 + k*40
			if d.b[o] == 0 && d.b[o+1] == 0 && d.b[o+2] == 0 && d.i(o+4) == 0 && d.i(o+8) == 0 {
				break
			}
			if d.i(o+28) != 0 {
				return true
			}
		}
	}
	return false
}

func refThickMM(name string) float64 {
	m := reMMRef.FindStringSubmatch(reParRef.ReplaceAllString(name, ""))
	if m == nil {
		return 0
	}
	v, _ := strconv.ParseFloat(strings.ReplaceAll(m[1], ",", "."), 64)
	if v <= 0 || v > 100 {
		return 0
	}
	return v
}

func refPasses(name string) int {
	if m := rePasRef.FindStringSubmatch(name); m != nil {
		if n, _ := strconv.Atoi(m[1]); n >= 2 && n <= 10 {
			return n
		}
	}
	return 1
}

// check : liste des écarts entre le .DAT et le réglage Job Manager.
func (d refDat) check(global, procs string, m Machine) []string {
	var errs []string
	expect := func(what string, got string, want any) {
		if got != fmt.Sprint(want) {
			errs = append(errs, fmt.Sprintf("%s %s au lieu de %v", what, got, want))
		}
	}
	near := func(what string, got string, want float64) {
		g, err := strconv.ParseFloat(got, 64)
		if err != nil || math.Abs(g-want) > 1e-3 {
			errs = append(errs, fmt.Sprintf("%s %s au lieu de %.4f", what, got, want))
		}
	}
	L := refLayouts[len(d.b)]
	mode, fus := d.i(272), d.fusion()
	dpi := d.h(90)
	if dpi <= 0 || dpi > 2400 {
		dpi = 600
	}
	dither := map[int]int{2: 2, 3: 6, 5: 3, 6: 4}[d.i(848)]
	if dither == 0 {
		dither = 1
	}
	sortWant := 1
	if L.sortOff >= 0 {
		switch {
		case d.i(L.sortOff) != 0:
			sortWant = 0
		case d.i(L.sortMode) != 0:
			sortWant = 2
		}
	}
	laser := 0
	// machine uniquement Fibre : source Fibre imposée ; sinon celle du .DAT Fusion
	if !m.HasCO2() || fus && d.i(3096) != 0 {
		laser = 1
	}
	thick := refThickMM(d.name)
	if fus && d.h(3104) > 0 {
		thick = float64(d.h(3104)) / 100
	}
	offset := 0.0
	if fus {
		offset = float64(d.h(3148)) / 100
	}
	passes := refPasses(d.name)
	low := foldKey(d.name)
	passOnEngrave := mode == 1 || (mode == 2 && (strings.Contains(low, "grav") || strings.Contains(low, "def")))

	var nGrav, nCut int
	for _, m := range reProc.FindAllStringSubmatch(procs, -1) {
		a := map[string]string{}
		for _, kv := range reAttr.FindAllStringSubmatch(m[1], -1) {
			a[kv[1]] = kv[2]
		}
		if a["processType"] == "2" {
			expect("assistance d'air", a["airAssist"], 1) // décision VAP : air sur toutes les découpes
		}
		if strings.Contains(m[2], `name="color"`) {
			continue // procédé d'une couleur du pilote : vérifié par l'air ci-dessus
		}
		expect("trame", a["dithering"], dither)
		expect("sens de gravure", a["rasterOrder"], d.i(852))
		expect("tri des vecteurs", a["vectorSort"], sortWant)
		expect("source laser", a["laserType"], laser)
		near("décalage de focale", a["offsetFocus"], offset/25.4*1000)
		switch a["processType"] {
		case "1":
			nGrav++
			expect("puissance gravure", a["rasterPower"], d.i(256))
			expect("vitesse gravure", a["rasterSpeed"], d.i(260))
			expect("résolution", a["rasterResolution"], dpi)
			wantP := 1
			if passOnEngrave {
				wantP = passes
			}
			expect("passes gravure", a["passes"], wantP)
			if laser == 1 && d.i(3100) > 0 {
				expect("fréquence gravure fibre", a["rasterFreq"], d.i(3100))
			}
		case "2":
			nCut++
			expect("puissance découpe", a["vectorPower"], d.i(264))
			expect("vitesse découpe", a["vectorSpeed"], d.i(268))
			f := float64(d.i(836))
			if !fus {
				f = 1 + (f-10)*99/4990
			}
			near("fréquence découpe", a["vectorFreq"], math.Max(1, math.Min(100, f)))
			near("épaisseur", a["materialThickness"], thick/25.4*1000)
			wantP := 1
			if !passOnEngrave {
				wantP = passes
			}
			expect("passes découpe", a["passes"], wantP)
		}
	}
	wantG, wantC := map[int][2]int{0: {0, 1}, 1: {1, 0}, 2: {1, 1}}[mode][0], map[int][2]int{0: {0, 1}, 1: {1, 0}, 2: {1, 1}}[mode][1]
	if nGrav != wantG || nCut != wantC {
		errs = append(errs, fmt.Sprintf("procédés %d gravure / %d découpe au lieu de %d / %d", nGrav, nCut, wantG, wantC))
	}

	g := map[string]string{}
	for _, kv := range reGlobal.FindAllStringSubmatch(global, -1) {
		g[kv[1]] = kv[2]
	}
	af := 0
	if L.af >= 0 && d.i(L.af) != 0 {
		af = 1
	}
	expect("autofocus", g["autoFocus"], af)
	center := 0
	if d.i(788) != 0 {
		center = map[int]int{0: 3, 1: 5, 2: 2}[d.i(L.pos)]
	}
	expect("position du travail", g["centerEngraving"], center)
	rt := 0
	switch {
	case mode == 0:
	case d.i(808) != 0:
		rt = 1
	case d.i(804) != 0:
		rt = 2
	}
	expect("type de gravure", g["rasterType"], rt)
	expect("résolution globale", g["rasterResolution"], dpi)
	if fus {
		expect("compensation vitesse", g["vectorSpeedComp"], b2i(d.i(832) != 0))
		expect("compensation puissance", g["vectorPowerComp"], b2i(d.i(3116) != 0))
		expect("grille de découpe", g["vectorGrid"], b2i(d.i(3112) != 0))
	}
	return errs
}

func pickMachine(ms []Machine, env, key string) (Machine, bool) {
	want := strings.ToLower(os.Getenv(env))
	for _, m := range ms {
		n := strings.ToLower(m.Name)
		if (want != "" && n == want) || (want == "" && strings.Contains(n, key)) {
			return m, true
		}
	}
	return Machine{}, false
}

func TestVerifReglages(t *testing.T) {
	ref, bk := os.Getenv("VAP_REF"), os.Getenv("VAP_BACKUP")
	if ref == "" || bk == "" {
		t.Skip("VAP_REF et VAP_BACKUP non définis")
	}
	cur = LFR
	ro, err := openRO(bk)
	if err != nil {
		t.Fatal(err)
	}
	ms, err := LoadMachines(ro)
	ro.Close()
	if err != nil {
		t.Fatal(err)
	}
	machines := map[int]Machine{}
	for size, k := range map[int][2]string{2924: {"VAP_MINI", "mini"}, 2992: {"VAP_ZING", "zing"}, 3156: {"VAP_FUSION", "fusion"}} {
		if m, ok := pickMachine(ms, k[0], k[1]); ok {
			machines[size] = m
		}
	}

	// chaque dossier contenant des .DAT est converti à part, comme le ferait un utilisateur
	var dirs []string
	filepath.WalkDir(ref, func(p string, f os.DirEntry, err error) error {
		if err == nil && f.IsDir() {
			if l, _ := ListDats(p, false); len(l) > 0 {
				dirs = append(dirs, p)
			}
		}
		return nil
	})
	total, totalOK := 0, 0
	for _, dir := range dirs {
		label, _ := filepath.Rel(ref, dir)
		t.Run(label, func(t *testing.T) {
			var dats []refDat
			paths, _ := ListDats(dir, false)
			for _, p := range paths {
				b, _ := os.ReadFile(p)
				dats = append(dats, refDat{strings.TrimSpace(strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))), b})
			}
			if len(dats) == 0 {
				t.Skip("aucun .DAT")
			}
			m, ok := machines[len(dats[0].b)]
			if !ok {
				t.Skipf("pas de machine pour les .DAT de %d octets", len(dats[0].b))
			}
			db0 := filepath.Join(t.TempDir(), "copie.epidb")
			if err := CopyFile(bk, db0); err != nil {
				t.Fatal(err)
			}
			db, err := openRW(db0)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			folder := "TEST " + filepath.Base(dir)
			o := Options{DatDir: dir, Machine: m, FolderName: folder, AirAssistCut: true, VectorSort: SortFromDat, OnTop: true, ReplaceFolder: true, Fiber: !m.HasCO2()}
			res, err := Convert(db, o, func(string) {}, nil)
			if err != nil {
				t.Fatal(err)
			}
			fid, _ := FindJobFolder(db, folder)
			rows, err := db.Query(`SELECT Name, GlobalSettings, ProcessSettings FROM Settings WHERE FolderID=?`, fid)
			if err != nil {
				t.Fatal(err)
			}
			got := map[string][2]string{}
			for _, r := range rows {
				got[foldKey(toStr(r[0]))] = [2]string{toStr(r[1]), toStr(r[2])}
			}
			// même ordre que le convertisseur : le premier de deux noms identiques est gardé
			sort.SliceStable(dats, func(i, j int) bool { return foldKey(dats[i].name) < foldKey(dats[j].name) })
			seen := map[string]bool{}
			nOK, nBad := 0, 0
			for _, d := range dats {
				k := foldKey(d.name)
				s, present := got[k]
				switch {
				case !d.valid(m):
					if present && !seen[k] {
						t.Errorf("%s : aurait dû être ignoré", d.name)
						nBad++
					}
					continue
				case seen[k]:
					continue // doublon de nom, ignoré par le convertisseur
				case !present:
					t.Errorf("%s : absent de Job Manager", d.name)
					nBad++
					continue
				}
				seen[k] = true
				if errs := d.check(s[0], s[1], m); len(errs) > 0 {
					t.Errorf("%s : %s", d.name, strings.Join(errs, " ; "))
					nBad++
				} else {
					nOK++
				}
			}
			if len(res.Converted) != len(got) {
				t.Errorf("%d réglages convertis mais %d dans le dossier", len(res.Converted), len(got))
			}
			total += nOK + nBad
			totalOK += nOK
			why := map[string]int{}
			for _, s := range res.Skipped {
				if k := strings.LastIndex(s, "("); k >= 0 {
					s = s[k:]
				}
				why[s]++
			}
			t.Logf("%-40s machine %-16s %d réglages conformes, %d écarts, %d ignorés %v", label, m.Name, nOK, nBad, len(res.Skipped), why)
		})
	}
	t.Logf("TOTAL : %d / %d réglages conformes", totalOK, total)
}
