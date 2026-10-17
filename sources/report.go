package main

// Rapport d'erreur : écrit quand une conversion échoue, pour comprendre le
// problème à distance sans la sauvegarde complète (45 Mo). Il contient les
// versions, la structure de la base, les noms des machines et des dossiers
// Job Manager, la liste des .DAT et le journal ; le dossier personnel Windows
// (C:\Users\<nom>) y est remplacé par %USERPROFILE%. Toujours en français,
// il est lu par VAP.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ReportInput struct {
	Version string
	System  []string // lignes propres à la plateforme (Windows, bibliothèque SQLite…)
	Backup  string
	Options Options
	Log     string
	Err     error
}

func yesNo(b bool) string {
	if b {
		return "oui"
	}
	return "non"
}

func fileSHA256(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "illisible : " + err.Error()
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "illisible : " + err.Error()
	}
	return hex.EncodeToString(h.Sum(nil))
}

// reportBackup : structure et contenu utile de la sauvegarde d'origine,
// ouverte en lecture seule (la copie de travail est supprimée après l'échec).
func reportBackup(open func(string) (DB, error), path string, w func(string, ...any)) {
	w("Fichier : %s", path)
	st, err := os.Stat(path)
	if err != nil {
		w("Introuvable : %v", err)
		return
	}
	w("Taille : %d octets, modifié le %s", st.Size(), st.ModTime().Format("2006-01-02 15:04:05"))
	w("SHA-256 : %s", fileSHA256(path))
	db, err := open(path)
	if err != nil {
		w("Ouverture impossible : %v", err)
		return
	}
	defer db.Close()
	one := func(sql string) string {
		r, err := db.Query(sql)
		if err != nil {
			return "erreur : " + err.Error()
		}
		var vals []string
		for _, row := range r {
			if len(row) > 0 {
				vals = append(vals, toStr(row[0]))
			}
		}
		return strings.Join(vals, " / ")
	}
	w("Vérification rapide : %s", one(`PRAGMA quick_check(5)`))
	w("user_version : %s", one(`PRAGMA user_version`))
	w("Réglages (Settings) : %s", one(`SELECT COUNT(*) FROM Settings`))

	w("")
	w("Machines :")
	if rows, err := db.Query(`SELECT ID, Name, LaserType, TableSize FROM Machines ORDER BY SortID, ID`); err != nil {
		w("  erreur : %v", err)
	} else {
		for _, r := range rows {
			w("  %d · %s · laser %s · table %s", toInt(r[0]), toStr(r[1]), toStr(r[2]), toStr(r[3]))
		}
	}

	w("")
	w("Dossiers « Paramètres du travail » :")
	if rows, err := db.Query(`SELECT f.ID, f.Name, f.MachineID, (SELECT COUNT(*) FROM Settings s WHERE s.FolderID=f.ID) FROM MaterialFolders f WHERE f.ForJobSettings=1 ORDER BY f.SortID`); err != nil {
		w("  erreur : %v", err)
	} else {
		for _, r := range rows {
			w("  %d · %s · machine %d · %d réglages", toInt(r[0]), toStr(r[1]), toInt(r[2]), toInt(r[3]))
		}
	}

	w("")
	w("Structure des tables :")
	if rows, err := db.Query(`SELECT name, sql FROM sqlite_master WHERE type IN ('table','view','trigger') ORDER BY type, name`); err != nil {
		w("  erreur : %v", err)
	} else {
		for _, r := range rows {
			w("  %s", strings.Join(strings.Fields(toStr(r[1])), " "))
		}
	}
}

func reportDats(o Options, w func(string, ...any)) {
	paths, err := ListDats(o.DatDir, o.Recursive)
	if err != nil {
		w("Lecture du dossier impossible : %v", err)
		return
	}
	w("%d fichiers", len(paths))
	for _, p := range paths {
		rel, err := filepath.Rel(o.DatDir, p)
		if err != nil {
			rel = filepath.Base(p)
		}
		size := int64(-1)
		if st, err := os.Stat(p); err == nil {
			size = st.Size()
		}
		d, err := ParseDat(p)
		if err != nil {
			w("  %s · %d octets · ERREUR : %v", rel, size, err)
			continue
		}
		layout := "inconnu"
		if d.Known() {
			layout = d.Layout.Name
		}
		w("  %s · %d octets · driver %s · %s · mode %d · %d dpi · %d couleurs", rel, size, layout, d.Header, d.Mode, d.DPI, len(d.Colors))
	}
}

func BuildReport(open func(string) (DB, error), in ReportInput) string {
	var sb strings.Builder
	w := func(f string, a ...any) {
		fmt.Fprintf(&sb, f, a...)
		sb.WriteString("\r\n")
	}
	section := func(t string) {
		w("")
		w("== %s", t)
	}
	w("VAP Convertisseur – rapport d'erreur")
	w("Date : %s", time.Now().Format("2006-01-02 15:04:05"))
	w("Version : %s · langue %s", in.Version, langCodes[cur])
	for _, l := range in.System {
		w("%s", l)
	}

	section("Erreur")
	if in.Err != nil {
		w("%s", in.Err.Error())
	}

	o := in.Options
	section("Options")
	w("Dossier .DAT : %s", o.DatDir)
	w("Sous-dossiers : %s", yesNo(o.Recursive))
	m := o.Machine
	size := fnum(m.W) + " × " + fnum(m.H)
	if !m.SizeKnown {
		size += " (par défaut)"
	}
	w("Machine : %s (ID %d, laser %s, table %s)", m.Name, m.ID, m.LaserType, size)
	w("Nom du dossier : %s", o.FolderName)
	w("Tri des vecteurs : %d · en tête de liste : %s · source Fibre : %s · remplacer le dossier : %s",
		o.VectorSort, yesNo(o.OnTop), yesNo(o.Fiber), yesNo(o.ReplaceFolder))

	section("Sauvegarde Job Manager")
	reportBackup(open, in.Backup, w)

	section("Fichiers .DAT")
	reportDats(o, w)

	section("Journal")
	for _, l := range strings.Split(strings.ReplaceAll(in.Log, "\r\n", "\n"), "\n") {
		w("%s", l)
	}
	return maskHome(sb.String())
}

// maskHome : le dossier personnel (qui contient le nom d'utilisateur Windows)
// est remplacé par %USERPROFILE% dans tout le rapport.
func maskHome(s string) string {
	home, err := os.UserHomeDir()
	if err != nil || len(home) < 4 {
		return s
	}
	home = strings.TrimRight(home, `\/`)
	return regexp.MustCompile(`(?i)`+regexp.QuoteMeta(home)).ReplaceAllString(s, "%USERPROFILE%")
}

// WriteReport : "<dossier DAT> - rapport d'erreur.txt" à côté du dossier .DAT,
// ou dans le dossier temporaire si cet emplacement n'est pas accessible.
func WriteReport(open func(string) (DB, error), in ReportInput) (string, error) {
	data := []byte(string(rune(0xFEFF)) + BuildReport(open, in))
	name := safeFileName(filepath.Base(filepath.Clean(in.Options.DatDir)) + " - " + T("file.report"))
	p := filepath.Join(filepath.Dir(filepath.Clean(in.Options.DatDir)), name)
	if err := os.WriteFile(p, data, 0o644); err == nil {
		return p, nil
	}
	p = filepath.Join(os.TempDir(), name)
	return p, os.WriteFile(p, data, 0o644)
}
