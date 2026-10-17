package main

// Conversions complètes sans fichiers réels : .DAT synthétiques (dat_test.go)
// et petite base Job Manager créée ici avec les seules tables utilisées.
// Ces tests tournent à chaque modification sur GitHub.

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	mCO2   = Machine{ID: 1, Name: "Mini 24", LaserType: `<list><element val="CO2"/></list>`}
	mDual  = Machine{ID: 2, Name: "Fusion 32", LaserType: `<list><element val="CO2"/><element val="Fiber"/></list>`}
	mFiber = Machine{ID: 3, Name: "Fusion Fibre", LaserType: `<list><element val="Fiber"/></list>`}
)

// testBase : sauvegarde minimale avec les trois machines ci-dessus.
func testBase(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "Sauvegarde.epidb")
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	db, err := openTestDB(p)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, q := range []string{
		`CREATE TABLE Machines (ID INTEGER PRIMARY KEY AUTOINCREMENT, Name TEXT, LaserType TEXT, TableSize TEXT, SortID INTEGER)`,
		`CREATE TABLE MaterialFolders (ID INTEGER PRIMARY KEY AUTOINCREMENT, SortID INTEGER, Name TEXT, Subfolders TEXT, MachineID INTEGER, ForJobSettings INTEGER)`,
		`CREATE TABLE MaterialSubfolders (ID INTEGER PRIMARY KEY AUTOINCREMENT, SortID INTEGER, Name TEXT)`,
		`CREATE TABLE Settings (ID INTEGER PRIMARY KEY AUTOINCREMENT, Category TEXT, ColorSettings TEXT, CreationDate INTEGER, GlobalSettings TEXT,
			MachineID INTEGER, Name TEXT, Notes TEXT, PrintDate INTEGER, SortID INTEGER, SubCategory TEXT, OriginalColorSettings TEXT, FolderID INTEGER,
			SubFolderID INTEGER, ProcessSettings TEXT, OriginalProcessSettings TEXT, EpilogProvided INTEGER)`,
	} {
		if err := db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	for _, m := range []Machine{mCO2, mDual, mFiber} {
		if err := db.Exec(`INSERT INTO Machines (ID, Name, LaserType, TableSize, SortID) VALUES (?,?,?,?,?)`,
			m.ID, m.Name, m.LaserType, `<map><element key="width" val="24"/><element key="height" val="12"/></map>`, m.ID); err != nil {
			t.Fatal(err)
		}
	}
	return p
}

func machineIn(t *testing.T, base string, id int64) Machine {
	t.Helper()
	db, err := openTestDB(base)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ms, err := LoadMachines(db)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range ms {
		if m.ID == id {
			return m
		}
	}
	t.Fatalf("machine %d absente", id)
	return Machine{}
}

type convCase struct {
	t    *testing.T
	base string
	dats string
}

func newConv(t *testing.T) *convCase {
	cur = LFR
	root := t.TempDir()
	return &convCase{t, testBase(t), filepath.Join(root, "Reglages")}
}

func (c *convCase) add(name string, b []byte) {
	c.t.Helper()
	os.MkdirAll(c.dats, 0o755)
	if err := os.WriteFile(filepath.Join(c.dats, name+".DAT"), b, 0o644); err != nil {
		c.t.Fatal(err)
	}
}

// run : conversion complète, avec contrôle que la sauvegarde et les .DAT
// d'origine ne sont pas modifiés. Renvoie les procédés par nom de réglage.
func (c *convCase) run(m Machine, fiber bool) (Result, map[string]string, error) {
	c.t.Helper()
	sums := map[string][32]byte{}
	files, _ := filepath.Glob(filepath.Join(c.dats, "*"))
	for _, p := range append(files, c.base) {
		b, _ := os.ReadFile(p)
		sums[p] = sha256.Sum256(b)
	}
	o := Options{DatDir: c.dats, Machine: machineIn(c.t, c.base, m.ID), FolderName: m.Name, VectorSort: 1, OnTop: true, Fiber: fiber}
	out, res, err := Run(openTestDB, c.base, o, func(string) {}, nil)
	for p, s := range sums {
		if b, _ := os.ReadFile(p); sha256.Sum256(b) != s {
			c.t.Errorf("%s a été modifié", filepath.Base(p))
		}
	}
	if err != nil {
		if out != "" {
			c.t.Errorf("dossier de sortie rendu malgré l'erreur : %s", out)
		}
		return res, nil, err
	}
	dbs, _ := filepath.Glob(filepath.Join(out, "*.epidb"))
	if len(dbs) != 1 {
		c.t.Fatalf("base convertie introuvable dans %s", out)
	}
	db, err := openTestDB(dbs[0])
	if err != nil {
		c.t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(`SELECT s.Name, s.ProcessSettings FROM Settings s JOIN MaterialFolders f ON f.ID=s.FolderID WHERE f.MachineID=?`, m.ID)
	if err != nil {
		c.t.Fatal(err)
	}
	got := map[string]string{}
	for _, r := range rows {
		got[toStr(r[0])] = toStr(r[1])
	}
	return res, got, nil
}

func set32(b []byte, off, v int) []byte {
	c := append([]byte(nil), b...)
	binary.LittleEndian.PutUint32(c[off:], uint32(int32(v)))
	return c
}

func TestConversionFamilles(t *testing.T) {
	c := newConv(t)
	fams := map[string][]byte{
		"Mini":   fakeDat(2924, "Epilog Engraver Win32 Mini"),
		"Zing":   fakeDat(2992, "Epilog Engraver Win32 Zing"),
		"Fusion": fakeDat(3156, "Epilog Engraver Win32 Fusion"),
	}
	for f, b := range fams {
		c.add(f+" decoupe 3 mm", set32(b, 272, 0))
		c.add(f+" gravure", set32(b, 272, 1))
		c.add(f+" combine", b)
	}
	res, got, err := c.run(mCO2, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Converted) != 9 || len(res.Skipped) != 0 || len(got) != 9 {
		t.Fatalf("%d convertis, %d ignorés, %d en base : %v", len(res.Converted), len(res.Skipped), len(got), res.Skipped)
	}
	for f := range fams {
		for name, types := range map[string][]string{"decoupe 3 mm": {`processType="2"`}, "gravure": {`processType="1"`}, "combine": {`processType="1"`, `processType="2"`}} {
			ps := got[f+" "+name]
			for _, ty := range types {
				if !strings.Contains(ps, ty) {
					t.Errorf("%s %s : %s absent", f, name, ty)
				}
			}
			if !strings.Contains(ps, `laserType="0"`) || strings.Contains(ps, `laserType="1"`) {
				t.Errorf("%s %s : source autre que CO2", f, name)
			}
		}
	}
	if ps := got["Fusion decoupe 3 mm"]; !strings.Contains(ps, `airAssist="1"`) || !strings.Contains(ps, `vectorPower="80"`) || !strings.Contains(ps, `vectorSpeed="20"`) {
		t.Errorf("découpe Fusion : %s", ps)
	}
}

func TestConversionSources(t *testing.T) {
	fus := fakeDat(3156, "Epilog Engraver Win32 Fusion")
	fib := set32(fus, fusLaser, 1)

	t.Run("Fibre vers machine CO2 refusée", func(t *testing.T) {
		c := newConv(t)
		c.add("Inox fibre", fib)
		c.add("Bois CO2", fus)
		res, got, err := c.run(mCO2, false)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := got["Inox fibre"]; ok || len(res.Skipped) != 1 || !strings.Contains(res.Skipped[0], "pas de source Fibre") {
			t.Errorf("réglage Fibre non refusé : %v", res.Skipped)
		}
	})
	t.Run("choix Fibre sur machine CO2", func(t *testing.T) {
		c := newConv(t)
		c.add("Bois", fus)
		if _, _, err := c.run(mCO2, true); err == nil || !strings.Contains(err.Error(), "n'a pas de source Fibre") {
			t.Errorf("erreur attendue : %v", err)
		}
	})
	t.Run("choix CO2 sur machine Fibre", func(t *testing.T) {
		c := newConv(t)
		c.add("Inox", fib)
		if _, _, err := c.run(mFiber, false); err == nil || !strings.Contains(err.Error(), "n'a pas de source CO2") {
			t.Errorf("erreur attendue : %v", err)
		}
	})
	t.Run("machine CO2 et Fibre", func(t *testing.T) {
		c := newConv(t)
		c.add("Inox fibre", fib)
		c.add("Bois CO2", fus)
		_, got, err := c.run(mDual, false)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got["Inox fibre"], `laserType="1"`) || strings.Contains(got["Bois CO2"], `laserType="1"`) {
			t.Errorf("sources : fibre %q / co2 %q", got["Inox fibre"], got["Bois CO2"])
		}
	})
}

func TestConversionCouleurs(t *testing.T) {
	c := newConv(t)
	b := set32(fakeDat(3156, "Epilog Engraver Win32 Fusion"), 792, 1)
	o := colorTable
	b[o], b[o+3] = 0xff, 70 // rouge, puissance 70
	b = set32(b, o+4, 25)   // vitesse
	b = set32(b, o+8, 40)   // fréquence 40 %
	b = set32(b, o+20, 1)   // découpe
	c.add("Acrylique couleurs", b)
	_, got, err := c.run(mDual, false)
	if err != nil {
		t.Fatal(err)
	}
	ps := got["Acrylique couleurs"]
	if !strings.Contains(ps, `processName="Acrylique couleurs #FF0000"`) || !strings.Contains(ps, `value="#ffff0000"`) ||
		!strings.Contains(ps, `vectorPower="70"`) || !strings.Contains(ps, `vectorSpeed="25"`) || !strings.Contains(ps, `vectorFreq="40"`) {
		t.Errorf("procédé de la couleur : %s", ps)
	}
}

func TestConversionRefus(t *testing.T) {
	c := newConv(t)
	mini := fakeDat(2924, "Epilog Engraver Win32 Mini")
	c.add("Bon", mini)
	c.add("Taille inconnue", fakeDat(3000, "Epilog Engraver Win32 Mini"))
	c.add("Puissance 150", set32(mini, 256, 150))
	c.add("Mode 7", set32(mini, 272, 7))
	res, got, err := c.run(mCO2, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(res.Skipped) != 3 {
		t.Errorf("%d en base, ignorés : %v", len(got), res.Skipped)
	}
	// rien d'exploitable : erreur, et aucun dossier de sortie laissé
	c2 := newConv(t)
	c2.add("Taille inconnue", fakeDat(3000, "Epilog Engraver Win32 Mini"))
	if _, _, err := c2.run(mCO2, false); err == nil {
		t.Error("erreur attendue")
	}
	if ents, _ := os.ReadDir(filepath.Dir(c2.dats)); len(ents) != 1 {
		t.Errorf("dossier de sortie laissé : %d entrées", len(ents))
	}
}

func TestConversionDoublons(t *testing.T) {
	c := newConv(t)
	mini := fakeDat(2924, "Epilog Engraver Win32 Mini")
	os.MkdirAll(filepath.Join(c.dats, "A"), 0o755)
	os.MkdirAll(filepath.Join(c.dats, "B"), 0o755)
	os.MkdirAll(filepath.Join(c.dats, "C"), 0o755)
	os.WriteFile(filepath.Join(c.dats, "A", "Bois 3 mm.DAT"), mini, 0o644)
	os.WriteFile(filepath.Join(c.dats, "B", "BOIS 3 MM.DAT"), mini, 0o644)                 // même contenu
	os.WriteFile(filepath.Join(c.dats, "C", "Bois 3 mm.DAT"), set32(mini, 264, 90), 0o644) // autre puissance
	db0 := c.base
	o := Options{DatDir: c.dats, Recursive: true, Machine: machineIn(t, db0, mCO2.ID), FolderName: "Doublons", VectorSort: 1}
	_, res, err := Run(openTestDB, db0, o, func(string) {}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Converted) != 2 || len(res.Skipped) != 1 {
		t.Fatalf("%d convertis / %d ignorés : %v", len(res.Converted), len(res.Skipped), res.Skipped)
	}
	names := res.Converted[0].Name + " | " + res.Converted[1].Name
	if !strings.Contains(names, "Bois 3 mm (2)") {
		t.Errorf("second réglage non renommé : %s", names)
	}
}

func TestConversionRemplacement(t *testing.T) {
	c := newConv(t)
	c.add("Bois 3 mm", fakeDat(2924, "Epilog Engraver Win32 Mini"))
	// première conversion, dont la base sert ensuite de sauvegarde
	_, _, err := c.run(mCO2, false)
	if err != nil {
		t.Fatal(err)
	}
	outs, _ := filepath.Glob(filepath.Join(filepath.Dir(c.dats), "*", "*.epidb"))
	if len(outs) != 1 {
		t.Fatalf("%d bases", len(outs))
	}
	c.base = outs[0]
	c.add("Carton 2 mm", fakeDat(2924, "Epilog Engraver Win32 Mini"))
	if _, _, err := c.run(mCO2, false); !errors.Is(err, ErrFolderExists) {
		t.Fatalf("dossier existant : %v", err)
	}
	o := Options{DatDir: c.dats, Machine: machineIn(t, c.base, mCO2.ID), FolderName: mCO2.Name, VectorSort: 1, ReplaceFolder: true}
	out, res, err := Run(openTestDB, c.base, o, func(string) {}, nil)
	if err != nil || len(res.Converted) != 2 {
		t.Fatalf("remplacement : %v, %d", err, len(res.Converted))
	}
	dbs, _ := filepath.Glob(filepath.Join(out, "*.epidb"))
	db, _ := openTestDB(dbs[0])
	defer db.Close()
	r, _ := db.Query(`SELECT COUNT(*) FROM Settings`)
	f, _ := db.Query(`SELECT COUNT(*) FROM MaterialFolders`)
	if toInt(r[0][0]) != 2 || toInt(f[0][0]) != 1 {
		t.Errorf("après remplacement : %v réglages, %v dossiers (attendu 2 / 1)", r[0][0], f[0][0])
	}
}

// execIn : requête sur la sauvegarde de test avant conversion.
func execIn(t *testing.T, base, q string, args ...any) {
	t.Helper()
	db, err := openTestDB(base)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Exec(q, args...); err != nil {
		t.Fatal(err)
	}
}

func TestConversionDossiers(t *testing.T) {
	mini := fakeDat(2924, "Epilog Engraver Win32 Mini")

	t.Run("même nom sur une autre machine", func(t *testing.T) {
		c := newConv(t)
		c.add("Bois 3 mm", mini)
		execIn(t, c.base, `INSERT INTO MaterialFolders (ID, SortID, Name, Subfolders, MachineID, ForJobSettings) VALUES (50, 1, ?, '', ?, 1)`, mCO2.Name, mDual.ID)
		execIn(t, c.base, `INSERT INTO Settings (ID, Name, FolderID) VALUES (70, 'Réglage Fusion', 50)`)
		res, got, err := c.run(mCO2, false)
		if err != nil || len(res.Converted) != 1 || len(got) != 1 {
			t.Fatalf("%v %d %v", err, len(res.Converted), got)
		}
		if res.FolderID == 50 {
			t.Error("le dossier de l'autre machine a été repris")
		}
	})
	t.Run("plusieurs dossiers identiques", func(t *testing.T) {
		c := newConv(t)
		c.add("Bois 3 mm", mini)
		for _, id := range []int{50, 51} {
			execIn(t, c.base, `INSERT INTO MaterialFolders (ID, SortID, Name, Subfolders, MachineID, ForJobSettings) VALUES (?, 1, ?, '', ?, 1)`, id, mCO2.Name, mCO2.ID)
		}
		if _, _, err := c.run(mCO2, false); err == nil || !strings.Contains(err.Error(), "2 dossiers") {
			t.Errorf("erreur attendue : %v", err)
		}
	})
	t.Run("identifiants 1 2 3 déjà pris", func(t *testing.T) {
		c := newConv(t)
		c.add("Inox brossé", mini)
		for i, n := range []string{"Verre", "Pierre", "Cuir"} {
			execIn(t, c.base, `INSERT INTO MaterialSubfolders (ID, SortID, Name) VALUES (?,?,?)`, i+1, i+1, n)
		}
		_, _, err := c.run(mCO2, false)
		if err != nil {
			t.Fatal(err)
		}
		outs, _ := filepath.Glob(filepath.Join(filepath.Dir(c.dats), "*", "*.epidb"))
		db, _ := openTestDB(outs[0])
		defer db.Close()
		r, _ := db.Query(`SELECT f.Name FROM Settings s JOIN MaterialSubfolders f ON f.ID=s.SubFolderID WHERE s.Name='Inox brossé'`)
		if len(r) != 1 || toStr(r[0][0]) != "Métal" {
			t.Errorf("sous-dossier de « Inox brossé » : %v", r)
		}
	})
}
