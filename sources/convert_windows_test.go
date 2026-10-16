package main

// Conversion complète sans interface, pour contrôler une version avant diffusion :
//   set VAP_DAT=<dossier .DAT>  &  set VAP_BACKUP=<sauvegarde .epidb>  &  go test -run TestConvertDossier -v
// Le dossier de sortie est créé à côté du dossier .DAT, comme avec le programme.

import (
	"errors"
	"os"
	"testing"
)

func TestConvertDossier(t *testing.T) {
	dat, bk := os.Getenv("VAP_DAT"), os.Getenv("VAP_BACKUP")
	if dat == "" || bk == "" {
		t.Skip("VAP_DAT et VAP_BACKUP non définis")
	}
	cur = LFR
	db, err := openRO(bk)
	if err != nil {
		t.Fatal(err)
	}
	ms, err := LoadMachines(db)
	db.Close()
	if err != nil || len(ms) == 0 {
		t.Fatal("aucune machine", err)
	}
	m := ms[0]
	for _, x := range ms {
		if x.Name == os.Getenv("VAP_MACHINE") {
			m = x
		}
	}
	name := os.Getenv("VAP_FOLDER")
	if name == "" {
		name = m.Name
	}
	sortOpt := SortFromDat
	if v := os.Getenv("VAP_SORT"); v != "" { // 0 aucun, 1 intérieur/extérieur, 2 optimisé, 3 comme le .DAT
		sortOpt = int(v[0] - '0')
	}
	o := Options{DatDir: dat, Machine: m, FolderName: name, AirAssistCut: true, VectorSort: sortOpt, OnTop: true, ReplaceFolder: true}
	out, res, err := Run(openRW, bk, o, func(s string) { t.Log(s) }, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d réglages, %d ignorés -> %s", len(res.Converted), len(res.Skipped), out)
}

// Rapport d'erreur sur les mêmes fichiers, avec une erreur simulée :
//
//	go test -run TestRapport -v   (mêmes variables que TestConvertDossier)
func TestRapport(t *testing.T) {
	dat, bk := os.Getenv("VAP_DAT"), os.Getenv("VAP_BACKUP")
	if dat == "" || bk == "" {
		t.Skip("VAP_DAT et VAP_BACKUP non définis")
	}
	cur = LFR
	db, err := openRO(bk)
	if err != nil {
		t.Fatal(err)
	}
	ms, err := LoadMachines(db)
	db.Close()
	if err != nil || len(ms) == 0 {
		t.Fatal("aucune machine", err)
	}
	o := Options{DatDir: dat, Machine: ms[0], FolderName: ms[0].Name, AirAssistCut: true, OnTop: true}
	p, err := WriteReport(openRO, ReportInput{Version: appVersion(), System: reportSystem(), Backup: bk,
		Options: o, Log: "ligne 1\r\nligne 2", Err: errors.New("erreur simulée")})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(p)
}
