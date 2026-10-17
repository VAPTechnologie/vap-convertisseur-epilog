package main

// Cas limites (sur des copies, dans un dossier temporaire) :
//   set VAP_BACKUP=<sauvegarde Job Manager>  &  set VAP_SAMPLE=<un .DAT valide>
//   go test -run TestLimites -v
// Vérifie à chaque fois que la sauvegarde d'origine et les .DAT ne sont pas
// modifiés, et qu'aucun dossier de sortie à moitié rempli ne reste en cas d'erreur.

import (
	"crypto/sha256"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func sum(t *testing.T, p string) [32]byte {
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return sha256.Sum256(b)
}

type limEnv struct {
	backup string
	sample []byte
	mach   Machine
}

func setupLimites(t *testing.T) limEnv {
	bk, smp := os.Getenv("VAP_BACKUP"), os.Getenv("VAP_SAMPLE")
	if bk == "" || smp == "" {
		t.Skip("VAP_BACKUP et VAP_SAMPLE non définis")
	}
	cur = LFR
	b, err := os.ReadFile(smp)
	if err != nil {
		t.Fatal(err)
	}
	// copie de la sauvegarde : l'original n'est jamais utilisé directement
	dir := t.TempDir()
	cp := filepath.Join(dir, "Sauvegarde test.epidb")
	if err := CopyFile(bk, cp); err != nil {
		t.Fatal(err)
	}
	db, err := openRO(cp)
	if err != nil {
		t.Fatal(err)
	}
	ms, err := LoadMachines(db)
	db.Close()
	if err != nil || len(ms) == 0 {
		t.Fatal("aucune machine", err)
	}
	return limEnv{cp, b, ms[0]}
}

func (e limEnv) opts(dat string) Options {
	return Options{DatDir: dat, Machine: e.mach, FolderName: "TEST limites", AirAssistCut: true, VectorSort: 1, OnTop: true, ReplaceFolder: true}
}

func writeDat(t *testing.T, dir, name string, b []byte) string {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name+".DAT")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func patch(b []byte, off int, v int32) []byte {
	c := append([]byte(nil), b...)
	binary.LittleEndian.PutUint32(c[off:], uint32(v))
	return c
}

// run : conversion complète + contrôles communs.
func (e limEnv) run(t *testing.T, dat string) (string, Result, error) {
	before := sum(t, e.backup)
	datSums := map[string][32]byte{}
	paths, _ := ListDats(dat, false)
	for _, p := range paths {
		datSums[p] = sum(t, p)
	}
	out, res, err := Run(openRW, e.backup, e.opts(dat), func(string) {}, nil)
	if sum(t, e.backup) != before {
		t.Error("la sauvegarde d'origine a été modifiée")
	}
	for p, s := range datSums {
		if sum(t, p) != s {
			t.Errorf("%s a été modifié", filepath.Base(p))
		}
	}
	if err != nil {
		if _, e2 := os.Stat(OutputDir(dat)); e2 == nil && out != "" {
			t.Error("dossier de sortie laissé après une erreur")
		}
	}
	return out, res, err
}

func settingNames(t *testing.T, out string, mid int64) map[string]string {
	m, _ := filepath.Glob(filepath.Join(out, "*.epidb"))
	if len(m) != 1 {
		t.Fatalf("base convertie introuvable dans %s", out)
	}
	db, err := openRO(m[0])
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	fid, _ := FindJobFolder(db, "TEST limites", mid)
	rows, err := db.Query(`SELECT Name, ProcessSettings FROM Settings WHERE FolderID=?`, fid)
	if err != nil {
		t.Fatal(err)
	}
	r := map[string]string{}
	for _, x := range rows {
		r[toStr(x[0])] = toStr(x[1])
	}
	return r
}

func noLeftover(t *testing.T, parent string) {
	ents, _ := os.ReadDir(parent)
	for _, e := range ents {
		if strings.Contains(e.Name(), strings.TrimSpace(T("out.suffix"))) {
			t.Errorf("dossier de sortie laissé : %s", e.Name())
		}
	}
}

func TestLimites(t *testing.T) {
	e := setupLimites(t)

	t.Run("dossier vide", func(t *testing.T) {
		root := t.TempDir()
		dat := filepath.Join(root, "Vide")
		os.MkdirAll(dat, 0o755)
		if _, _, err := e.run(t, dat); err == nil {
			t.Error("une erreur était attendue")
		}
		noLeftover(t, root)
	})

	t.Run("fichiers abimes", func(t *testing.T) {
		dat := filepath.Join(t.TempDir(), "Abimes")
		writeDat(t, dat, "Bon 3 mm", e.sample)
		writeDat(t, dat, "Vide", nil)
		writeDat(t, dat, "Trop petit", e.sample[:100])
		writeDat(t, dat, "Coupe en deux", e.sample[:len(e.sample)/2])
		bad := append([]byte(nil), e.sample...)
		copy(bad, make([]byte, 64))
		writeDat(t, dat, "Pas Epilog", bad)
		bad2 := append([]byte(nil), e.sample...)
		copy(bad2[220:], "XXXX")
		writeDat(t, dat, "Format inconnu", bad2)
		writeDat(t, dat, "Mode 7", patch(e.sample, 272, 7))
		writeDat(t, dat, "Puissance 150", patch(e.sample, 256, 150))
		writeDat(t, dat, "Vitesse negative", patch(e.sample, 268, -5))
		writeDat(t, dat, "Resolution 0", patch(e.sample, 88, 0)) // DPI (octet 90) mis à 0 : 600 par défaut
		_, res, err := e.run(t, dat)
		if err != nil {
			t.Fatal(err)
		}
		// "Coupe en deux" garde une taille > 900 octets mais n'a plus une structure connue : refusé
		if len(res.Converted) != 2 || len(res.Skipped) != 8 {
			t.Errorf("%d convertis / %d ignorés, attendu 2 / 8 : %v", len(res.Converted), len(res.Skipped), res.Skipped)
		}
	})

	t.Run("noms accentues chinois japonais", func(t *testing.T) {
		dat := filepath.Join(t.TempDir(), "Langues é 木材")
		names := []string{"Érable vernis 3 mm", "木材 3 mm", "アクリル 5mm", "Plexi d'1,5 mm & co", "Œuvre ½ ° ü"}
		for _, n := range names {
			writeDat(t, dat, n, e.sample)
		}
		out, _, err := e.run(t, dat)
		if err != nil {
			t.Fatal(err)
		}
		got := settingNames(t, out, e.mach.ID)
		for _, n := range names {
			ps, ok := got[n]
			if !ok {
				t.Errorf("%q absent ou mal écrit dans Job Manager", n)
				continue
			}
			if strings.Contains(n, "&") && !strings.Contains(ps, "&amp; co") {
				t.Errorf("%q : & non échappé dans le XML", n)
			}
		}
		if th := ThicknessMils("木材 3 mm"); th < 118 || th > 119 {
			t.Errorf("épaisseur lue dans un nom chinois : %v", th)
		}
	})

	t.Run("noms en double", func(t *testing.T) {
		dat := filepath.Join(t.TempDir(), "Doublons")
		writeDat(t, dat, "Bois élevé 3 mm", e.sample)
		writeDat(t, dat, "BOIS ELEVE 3 mm", e.sample)
		_, res, err := e.run(t, dat)
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Converted) != 1 || len(res.Skipped) != 1 {
			t.Errorf("%d convertis / %d ignorés, attendu 1 / 1", len(res.Converted), len(res.Skipped))
		}
	})

	t.Run("chemin tres long", func(t *testing.T) {
		dat := t.TempDir()
		for len(dat) < 300 {
			dat = filepath.Join(dat, "Sous-dossier avec un nom assez long")
		}
		writeDat(t, dat, "Acrylique coulé 3 mm", e.sample)
		_, res, err := e.run(t, dat)
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Converted) != 1 {
			t.Errorf("%d convertis, attendu 1", len(res.Converted))
		}
	})

	t.Run("deux conversions de suite", func(t *testing.T) {
		dat := filepath.Join(t.TempDir(), "Deux fois")
		writeDat(t, dat, "Carton 2 mm", e.sample)
		o1, _, err := e.run(t, dat)
		if err != nil {
			t.Fatal(err)
		}
		o2, _, err := e.run(t, dat)
		if err != nil {
			t.Fatal(err)
		}
		if o1 == o2 || !strings.HasSuffix(o2, "(2)") {
			t.Errorf("la deuxième conversion devrait aller dans un nouveau dossier : %s / %s", o1, o2)
		}
	})

	t.Run("sauvegarde absente", func(t *testing.T) {
		root := t.TempDir()
		dat := filepath.Join(root, "Sans sauvegarde")
		writeDat(t, dat, "Bois 3 mm", e.sample)
		e2 := e
		e2.backup = filepath.Join(root, "introuvable.epidb")
		if _, _, err := Run(openRW, e2.backup, e2.opts(dat), func(string) {}, nil); err == nil {
			t.Error("une erreur était attendue")
		}
		noLeftover(t, root)
	})

	t.Run("sauvegarde abimee", func(t *testing.T) {
		root := t.TempDir()
		dat := filepath.Join(root, "Base abimee")
		writeDat(t, dat, "Bois 3 mm", e.sample)
		e2 := e
		e2.backup = filepath.Join(root, "abimee.epidb")
		os.WriteFile(e2.backup, []byte(strings.Repeat("pas une base SQLite ", 500)), 0o644)
		if _, _, err := e2.run(t, dat); err == nil {
			t.Error("une erreur était attendue")
		}
		noLeftover(t, root)
	})

	t.Run("sauvegarde verrouillee", func(t *testing.T) {
		root := t.TempDir()
		dat := filepath.Join(root, "Verrou")
		writeDat(t, dat, "Bois 3 mm", e.sample)
		e2 := e
		e2.backup = filepath.Join(root, "verrou.epidb")
		CopyFile(e.backup, e2.backup)
		p, _ := syscall.UTF16PtrFromString(e2.backup)
		h, err := syscall.CreateFile(p, syscall.GENERIC_READ, 0, nil, syscall.OPEN_EXISTING, 0, 0) // aucun partage
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.CloseHandle(h)
		if _, _, err := Run(openRW, e2.backup, e2.opts(dat), func(string) {}, nil); err == nil {
			t.Error("une erreur était attendue")
		}
		noLeftover(t, root)
	})
}
