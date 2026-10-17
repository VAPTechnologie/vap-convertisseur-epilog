package main

// Moteur de conversion : anciens fichiers .DAT Epilog (driver legacy)
// -> réglages "Paramètres du travail" (méthode Gamma-Tec) dans une
// sauvegarde de base Job Manager (.epidb, SQLite).

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf16"
)

// ---------------------------------------------------------------------------
// Base de données (implémentée par winsqlite3 sous Windows)

type DB interface {
	Query(sql string, args ...any) ([][]any, error)
	Exec(sql string, args ...any) error
	Close() error
}

// ---------------------------------------------------------------------------
// Lecture des .DAT

type DatFile struct {
	Path     string
	Name     string // nom sans extension
	Header   string // ex. "Epilog Engraver Win32 Zing"
	RPower   int
	RSpeed   int
	VPower   int
	VSpeed   int
	Mode     int // 0 vecteur, 1 raster, 2 combiné
	DPI      int
	FreqRaw  int
	IsFusion bool
	Dither   int
	Stamp    bool
	// Champs relevés dans les drivers d'origine (voir datLayouts),
	// lus seulement quand Layout est connu.
	Layout    *datLayout
	Bottom    bool // gravure de bas en haut
	AutoFocus bool
	SortOff   bool // tri des vecteurs décoché
	SortOpt   bool // tri "Optimiser" (sinon intérieur/extérieur)
	Raster3D  bool
	Position  int          // position du travail : -1 aucune, 0 Gauche-Centre, 1 Dessus-Centre, 2 Centre-Centre, 3 Page-Centre
	Colors    []ColorEntry // gestion des couleurs du driver (vide si inactive)
	Passes    int          // nombre de passes lu dans le nom ("2 passes"), 1 sinon
	PageW     float64      // taille de page du driver (pouces), 0 si inconnue
	PageH     float64
	// Fusion seulement
	Fiber      bool
	EngFreq    int     // fréquence de gravure (%), source Fibre
	Thickness  float64 // épaisseur matière (mm) pour l'autofocus
	Offset     float64 // décalage de focale (mm)
	SpeedComp  bool
	PowerComp  bool
	VectorGrid bool
}

// ColorEntry : une ligne de la table "Gest./Couleur" du driver
// (à partir de l'octet 1376 : R, G, B, puissance, puis en int32 vitesse,
// fréquence, gravure, découpe, et selon le driver air / épaisseur / source / offset).
type ColorEntry struct {
	RGB                string // "rrggbb"
	Power, Speed, Freq int
	Engrave, Cut, Air  bool
	HasAir             bool
	Fiber              bool
	Thick, Offset      float64 // millièmes de pouce (Fusion)
}

// datLayout : emplacement des réglages selon le driver d'origine. Relevé le
// 13/10/2026 en changeant une option à la fois dans chaque driver
// (dossiers "Tests driver Mini24 / Zing / Fusion"). -1 = absent.
type datLayout struct {
	Name                         string
	Size                         int
	AutoFocus, SortOff, SortMode int
	Position                     int
	ColStride, ColEng, ColCut    int
	ColAir, ColThick, ColLaser   int
	ColOffset                    int
	FreqPct                      bool // fréquences déjà en % (Fusion)
}

var datLayouts = []datLayout{
	{Name: "mini", Size: 2924, AutoFocus: 844, SortOff: 840, SortMode: 2888, Position: 1852,
		ColStride: 28, ColEng: 12, ColCut: 16, ColAir: 20, ColThick: -1, ColLaser: -1, ColOffset: -1},
	{Name: "zing", Size: 2992, AutoFocus: -1, SortOff: 840, SortMode: 2956, Position: 1920,
		ColStride: 32, ColEng: 12, ColCut: 16, ColAir: -1, ColThick: -1, ColLaser: -1, ColOffset: -1},
	{Name: "fusion", Size: 3156, AutoFocus: 844, SortOff: -1, SortMode: -1, Position: 2056,
		ColStride: 40, ColEng: 16, ColCut: 20, ColAir: 24, ColThick: 12, ColLaser: 28, ColOffset: 32, FreqPct: true},
}

// Réglages propres au driver Fusion (fichiers de 3156 octets).
const (
	fusSpeedComp  = 832
	fusLaser      = 3096 // 0 CO2, 1 Fibre
	fusEngFreq    = 3100
	fusThickness  = 3104 // int16, centièmes de mm
	fusVectorGrid = 3112
	fusPowerComp  = 3116
	fusOffset     = 3148 // int16, centièmes de mm
)

func findLayout(size int, header string) *datLayout {
	h := strings.ToLower(header)
	name := "mini"
	switch {
	case strings.Contains(h, "zing"):
		name = "zing"
	case strings.Contains(h, "fusion"):
		name = "fusion"
	}
	for i := range datLayouts {
		if datLayouts[i].Name == name && datLayouts[i].Size == size {
			return &datLayouts[i]
		}
	}
	return nil
}

// NeedsFiber : le .DAT Fusion demande la source Fibre (réglage général ou une couleur).
func (d DatFile) NeedsFiber() bool {
	if d.Fiber {
		return true
	}
	for _, c := range d.Colors {
		if c.Fiber {
			return true
		}
	}
	return false
}

// Known : la mise en page du .DAT a été relevée sur le driver d'origine.
func (d DatFile) Known() bool { return d.Layout != nil }

// JMCenter : point de centrage Job Manager (0 Aucun, 2 Centre-Centre,
// 3 Gauche-Centre, 5 Centre-Haut). Page-Centre n'a pas d'équivalent.
func (d DatFile) JMCenter() int {
	switch d.Position {
	case 0:
		return 3
	case 1:
		return 5
	case 2:
		return 2
	}
	return 0
}

const (
	colorTable    = 1376
	colorMaxCount = 16
)

func mmToMils(mm float64) float64 { return mm / 25.4 * 1000 }

var rePasses = regexp.MustCompile(`(?i)(\d+)\s*pass`)

// PassesFromName : "LaserGlow 1.6 mm (2 passes)", "2 pass", "2 passages" -> 2.
func PassesFromName(name string) int {
	if m := rePasses.FindStringSubmatch(name); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil && n >= 2 && n <= 10 {
			return n
		}
	}
	return 1
}

// PassesOnEngrave : pour un combiné, les passes vont sur la gravure si le nom
// parle de gravure ou de profondeur ("Déf."), sinon sur la découpe.
func (d DatFile) PassesOnEngrave() bool {
	switch d.Mode {
	case 1:
		return true
	case 0:
		return false
	}
	s := foldKey(d.Name)
	return strings.Contains(s, "grav") || strings.Contains(s, "def")
}

// FreqToPercent : fréquence Hz du driver -> % Job Manager (10 Hz = 1 %, 5000 Hz = 100 %).
func FreqToPercent(hz int) float64 {
	return math.Max(1, math.Min(100, 1+(float64(hz)-10)*99/4990))
}

// JMDither : trame du driver -> trame Job Manager
// (driver : 1 Standard, 2 Floyd, 3 Brighten, 4 Low Res, 5 Jarvis, 6 Stucki ;
// Job Manager : 0 Aucun, 1 Standard, 2 Floyd, 3 Jarvis, 4 Stucki, 5 Bayer, 6 Brighten).
func (d DatFile) JMDither() int {
	if !d.Known() {
		return 1
	}
	switch d.Dither {
	case 2:
		return 2
	case 3:
		return 6
	case 5:
		return 3
	case 6:
		return 4
	}
	return 1
}

// JMSort : tri des vecteurs Job Manager (0 aucun, 1 intérieur/extérieur, 2 optimisé).
// opt = SortFromDat reprend le réglage du .DAT.
func (d DatFile) JMSort(opt int) int {
	if opt != SortFromDat {
		return opt
	}
	switch {
	case !d.Known() || d.Layout.SortOff < 0:
		return 1
	case d.SortOff:
		return 0
	case d.SortOpt:
		return 2
	}
	return 1
}

// JMRasterType : type de gravure Job Manager (0 Standard, 1 Tampon, 2 3D).
func (d DatFile) JMRasterType() int {
	switch {
	case d.Known() && d.Stamp:
		return 1
	case d.Known() && d.Raster3D:
		return 2
	}
	return 0
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (d DatFile) ModeLabel() string {
	switch d.Mode {
	case 0:
		return T("mode.cut")
	case 1:
		return T("mode.engrave")
	}
	return T("mode.combined")
}

// FreqPercent : fréquence vecteur au format Job Manager (1-100 %).
// Mini / Helix / Zing : Hz (10 Hz = 1 %, 5000 Hz = 100 %). Fusion : déjà en %.
func (d DatFile) FreqPercent() float64 {
	if !d.IsFusion {
		return FreqToPercent(d.FreqRaw)
	}
	return math.Max(1, math.Min(100, float64(d.FreqRaw)))
}

func le32(b []byte, off int) int { return int(int32(binary.LittleEndian.Uint32(b[off:]))) }

func ParseDat(path string) (DatFile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return DatFile{}, err
	}
	if len(b) < 900 {
		return DatFile{}, errors.New(T("err.tooSmall"))
	}
	u := make([]uint16, 32)
	for i := range u {
		u[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	h := string(utf16.Decode(u))
	if i := strings.IndexRune(h, 0); i >= 0 {
		h = h[:i]
	}
	if !strings.HasPrefix(h, "Epilog") {
		return DatFile{}, errors.New(T("err.notEpilog"))
	}
	if string(b[220:224]) != "MDPE" {
		return DatFile{}, errors.New(T("err.unknownFormat"))
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	d := DatFile{
		Path:     path,
		Name:     strings.TrimSpace(base),
		Header:   h,
		RPower:   le32(b, 256),
		RSpeed:   le32(b, 260),
		VPower:   le32(b, 264),
		VSpeed:   le32(b, 268),
		Mode:     le32(b, 272),
		DPI:      int(int16(binary.LittleEndian.Uint16(b[90:]))),
		Stamp:    le32(b, 808) != 0,
		FreqRaw:  le32(b, 836),
		Dither:   le32(b, 848),
		IsFusion: strings.Contains(strings.ToLower(h), "fusion"),
	}
	d.Position = -1
	// Seules les structures relevées sur les pilotes d'origine sont converties :
	// une autre taille donnerait des valeurs de remplacement présentées comme fidèles.
	L := findLayout(len(b), h)
	if L == nil {
		return d, fmt.Errorf(T("err.unknownLayout"), len(b))
	}
	d.Layout = L
	d.Raster3D = le32(b, 804) != 0
	d.Bottom = le32(b, 852) != 0
	if L.AutoFocus >= 0 {
		d.AutoFocus = le32(b, L.AutoFocus) != 0
	}
	if L.SortOff >= 0 {
		d.SortOff = le32(b, L.SortOff) != 0
		d.SortOpt = le32(b, L.SortMode) != 0
	}
	if le32(b, 788) != 0 {
		d.Position = le32(b, L.Position)
	}
	if L.Name == "fusion" {
		d.SpeedComp = le32(b, fusSpeedComp) != 0
		d.PowerComp = le32(b, fusPowerComp) != 0
		d.Fiber = le32(b, fusLaser) != 0
		d.EngFreq = le32(b, fusEngFreq)
		d.Thickness = float64(int16(binary.LittleEndian.Uint16(b[fusThickness:]))) / 100
		// int16 : certains anciens .DAT ont autre chose dans les 2 octets suivants
		// (00 00 30 00 lu en int32 donnait 31457,28 mm).
		d.Offset = float64(int16(binary.LittleEndian.Uint16(b[fusOffset:]))) / 100
		d.VectorGrid = le32(b, fusVectorGrid) != 0
	}
	if le32(b, 792) != 0 {
		for k := 0; k < colorMaxCount; k++ {
			o := colorTable + k*L.ColStride
			if o+L.ColStride > len(b) || (b[o] == 0 && b[o+1] == 0 && b[o+2] == 0 && le32(b, o+4) == 0 && le32(b, o+8) == 0) {
				break
			}
			c := ColorEntry{
				RGB:   fmt.Sprintf("%02x%02x%02x", b[o], b[o+1], b[o+2]),
				Power: int(b[o+3]), Speed: le32(b, o+4), Freq: le32(b, o+8),
				Engrave: le32(b, o+L.ColEng) != 0, Cut: le32(b, o+L.ColCut) != 0,
			}
			if L.ColAir >= 0 {
				c.HasAir, c.Air = true, le32(b, o+L.ColAir) != 0
			}
			if L.ColThick >= 0 {
				c.Thick = float64(le32(b, o+L.ColThick))
				c.Fiber = le32(b, o+L.ColLaser) != 0
				c.Offset = float64(le32(b, o+L.ColOffset))
			}
			d.Colors = append(d.Colors, c)
		}
	}
	d.Passes = PassesFromName(d.Name)
	// taille de page du DEVMODE (dmPaperWidth / dmPaperLength, dixièmes de mm)
	if w, l := int16(binary.LittleEndian.Uint16(b[82:])), int16(binary.LittleEndian.Uint16(b[80:])); w > 0 && l > 0 {
		d.PageW, d.PageH = float64(w)/254, float64(l)/254
	}
	if d.Mode < 0 || d.Mode > 2 {
		return d, fmt.Errorf(T("err.unknownJob"), d.Mode)
	}
	for _, v := range []int{d.RPower, d.RSpeed, d.VPower, d.VSpeed} {
		if v < 0 || v > 100 {
			return d, errors.New(T("err.outOfRange"))
		}
	}
	if d.DPI <= 0 || d.DPI > 2400 {
		d.DPI = 600
	}
	return d, nil
}

// ListDats : fichiers .DAT du dossier (et des sous-dossiers si demandé).
func ListDats(dir string, recursive bool) ([]string, error) {
	var out []string
	if recursive {
		err := filepath.WalkDir(dir, func(p string, e os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !e.IsDir() && strings.EqualFold(filepath.Ext(p), ".dat") {
				out = append(out, p)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		ents, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, e := range ents {
			if !e.IsDir() && strings.EqualFold(filepath.Ext(e.Name()), ".dat") {
				out = append(out, filepath.Join(dir, e.Name()))
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return foldKey(out[i]) < foldKey(out[j]) })
	return out, nil
}

// ---------------------------------------------------------------------------
// Classement et épaisseur

var accents = strings.NewReplacer(
	"à", "a", "â", "a", "ä", "a", "á", "a", "ã", "a", "å", "a",
	"ç", "c", "é", "e", "è", "e", "ê", "e", "ë", "e",
	"î", "i", "ï", "i", "í", "i", "ì", "i",
	"ô", "o", "ö", "o", "ó", "o", "ò", "o", "õ", "o",
	"ù", "u", "û", "u", "ü", "u", "ú", "u", "ÿ", "y", "ñ", "n", "œ", "oe", "æ", "ae")

func foldKey(s string) string {
	return accents.Replace(strings.Map(unicode.ToLower, s))
}

// SortFromDat : option "tri des vecteurs comme dans le .DAT".
const SortFromDat = 3

const (
	SubRoot = iota
	SubMetal
	SubBois
	SubPlastique
)

var (
	kwMetal = []string{"alulaz", "alumamark", "durablack", "formalu", "inox", "aluminium", "aluminum",
		"thermark", "laiton", "titan", "acier", "cuivre", "cermark", "metal brut", "fibergrave"}
	kwBois  = []string{"bois", "melamin", "cpl", "colorshop", "flexiwood", "mdf", "peuplier", "planchette", "contreplaque", "liege"}
	kwRoot  = []string{"cuir", "papier", "carton plume", "granit-", "granit ", "granit&", "marbre", "verre", "miroir", "tissu", "ardoise"}
	reMM    = regexp.MustCompile(`(?i)(\d+(?:[.,]\d+)?)\s*mm`)
	reParen = regexp.MustCompile(`\([^)]*\)`)
)

func Classify(name string) int {
	s := foldKey(name)
	for _, k := range kwMetal {
		if strings.Contains(s, k) {
			return SubMetal
		}
	}
	if !strings.Contains(s, "formaglas") && !strings.Contains(s, "plastique") && !strings.Contains(s, "granitex") && !strings.Contains(s, "granulex") {
		for _, k := range kwRoot {
			if strings.Contains(s, k) {
				return SubRoot
			}
		}
	}
	for _, k := range kwBois {
		if strings.Contains(s, k) {
			return SubBois
		}
	}
	return SubPlastique
}

// ThicknessMils : épaisseur matière (millièmes de pouce) lue dans le nom,
// en ignorant ce qui est entre parenthèses (ex. "Déf. 2 mm" = profondeur).
func ThicknessMils(name string) float64 {
	o := reParen.ReplaceAllString(name, "")
	m := reMM.FindStringSubmatch(o)
	if m == nil {
		return 0
	}
	v, err := strconv.ParseFloat(strings.ReplaceAll(m[1], ",", "."), 64)
	if err != nil || v <= 0 || v > 100 {
		return 0
	}
	return v / 25.4 * 1000
}

// ---------------------------------------------------------------------------
// Construction XML (format identique aux réglages Gamma-Tec existants)

func xmlEsc(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}

func fnum(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

const swThreshold = "0.00299213"

type proc struct {
	name        string
	ptype       int
	thick       float64
	rp, rs, dpi int
	vp, vs      int
	vf          string
	air, sort   int
	query       string
	pulse       string
	laser       int
	objs        string
	dither      int
	order       int // sens de gravure : 0 haut vers bas, 1 bas vers haut
	passes      int
	rfreq       string  // fréquence de gravure (%), "" = 50
	focus       float64 // décalage de focale (millièmes de pouce)
}

func (p proc) xml() string {
	if p.passes < 1 {
		p.passes = 1
	}
	if p.rfreq == "" {
		p.rfreq = "50"
	}
	return fmt.Sprintf(`<process airAssist="%d" dithering="%d" engraveHighQuality="0" engraveUnidirectional="0" hatchBreak="0" hatches="&lt;list&gt;&lt;/list&gt;" laserType="%d" materialThickness="%s" offsetFocus="%s" passes="%d" processName="%s" processType="%d" rasterFreq="%s" rasterOrder="%d" rasterPower="%d" rasterPulse="%s" rasterResolution="%d" rasterSpeed="%d" registration="0" thumbnails="" timeEstimation="" transform="" vectorAllowBezier="1" vectorFreq="%s" vectorPower="%d" vectorPulse="%s" vectorSort="%d" vectorSpeed="%d"><query><container>%s</container></query><objects>%s</objects></process>`,
		p.air, p.dither, p.laser, fnum(p.thick), fnum(p.focus), p.passes, p.name, p.ptype, p.rfreq, p.order, p.rp, p.pulse, p.dpi, p.rs, p.vf, p.vp, p.pulse, p.sort, p.vs, p.query, p.objs)
}

// colorQuery : objets d'une couleur donnée, traits fins (découpe) ou épais (gravure),
// sur le modèle des réglages par couleur Gamma-Tec.
func colorQuery(op, rgb string) string {
	return `<intersection operator="eq"><attribute name="strokeWidth" operator="` + op + `" value="` + swThreshold +
		`"/><attribute name="color" operator="eq" value="#ff` + rgb + `"/></intersection>`
}

var (
	qGT  = `<attribute name="strokeWidth" operator="gt" value="` + swThreshold + `"/>`
	qLTE = `<attribute name="strokeWidth" operator="lte" value="` + swThreshold + `"/>`
	qAll = `<all/>`
)

func objPage(i int) string {
	return fmt.Sprintf(`<processPage page="0"><object id="obj%d" isWhite="0"/></processPage>`, i)
}

func origProcesses(mode int) string {
	def := func(name string, t int, q, objs string) string {
		return proc{name: name, ptype: t, rp: 50, rs: 50, dpi: 500, vp: 50, vs: 50, vf: "50", air: 0, sort: 2, query: q, pulse: "0", objs: objs, dither: 1}.xml()
	}
	if mode == 2 {
		return "<list>" + def("Gravure", 1, qGT, objPage(2)) + def("Découpe", 2, qLTE, objPage(1)) + "</list>\n"
	}
	t := 1
	if mode == 0 {
		t = 2
	}
	return "<list>" + def("Tout", t, qAll, objPage(1)) + "</list>\n"
}

const globalTpl = `<map><element key="" val=""/><element key="allowAdvancedImport" val=""/><element key="allowHighContrast" val="1"/><element key="autoFocus" val="%d"/><element key="autoFocusType" val="0"/><element key="centerEngraving" val="%d"/><element key="centerOffsetX" val="0"/><element key="centerOffsetY" val="0"/><element key="copies" val="1"/><element key="height" val="%s"/><element key="leftOffset" val="0"/><element key="rasterOrder" val="%d"/><element key="rasterResolution" val="%d"/><element key="rasterType" val="%d"/><element key="registrationSize" val=""/><element key="rotaryDiameter" val="7.000000000000001"/><element key="rotaryEnable" val="0"/><element key="rotaryType" val="0"/><element key="stampFence" val="0"/><element key="stampFenceMargins" val="0"/><element key="stampProfile" val="&lt;list/&gt;"/><element key="stampShoulder" val="0.02"/><element key="stampWidening" val="0"/><element key="topOffset" val="0"/><element key="vectorGrid" val="%d"/><element key="vectorPowerComp" val="%d"/><element key="vectorSpeedComp" val="%d"/><element key="width" val="%s"/></map>`

// ---------------------------------------------------------------------------
// Machines

type Machine struct {
	ID        int64
	Name      string
	LaserType string
	W, H      float64
	SizeKnown bool // taille de table enregistrée dans Job Manager (sinon 24 x 12 par défaut)
}

func (m Machine) HasCO2() bool   { return strings.Contains(m.LaserType, `"CO2"`) || m.LaserType == "" }
func (m Machine) HasFiber() bool { return strings.Contains(m.LaserType, `"Fiber"`) }
func (m Machine) String() string { return m.Name }

var reTable = regexp.MustCompile(`val="([0-9.]+)"`)

func toInt(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case float64:
		return int64(x)
	case string:
		i, _ := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
		return i
	case []byte:
		i, _ := strconv.ParseInt(strings.TrimSpace(string(x)), 10, 64)
		return i
	}
	return 0
}

func toStr(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return fnum(x)
	}
	return ""
}

func LoadMachines(db DB) ([]Machine, error) {
	rows, err := db.Query(`SELECT ID, Name, LaserType, TableSize FROM Machines ORDER BY SortID, ID`)
	if err != nil {
		return nil, err
	}
	var ms []Machine
	for _, r := range rows {
		m := Machine{ID: toInt(r[0]), Name: toStr(r[1]), LaserType: toStr(r[2]), W: 24, H: 12}
		if t := reTable.FindAllStringSubmatch(toStr(r[3]), -1); len(t) == 2 {
			w, _ := strconv.ParseFloat(t[0][1], 64)
			h, _ := strconv.ParseFloat(t[1][1], 64)
			if w > 0 && h > 0 {
				m.W, m.H, m.SizeKnown = w, h, true
			}
		}
		ms = append(ms, m)
	}
	return ms, nil
}

// ---------------------------------------------------------------------------
// Conversion

type Options struct {
	DatDir        string
	Recursive     bool
	Machine       Machine
	FolderName    string
	AirAssistCut  bool
	VectorSort    int // 0 aucun, 1 intérieur vers extérieur, 2 optimisé, 3 comme le .DAT
	OnTop         bool
	Fiber         bool
	ReplaceFolder bool
}

type Result struct {
	Converted []DatFile
	Skipped   []string
	FolderID  int64
}

var ErrFolderExists = errors.New("folder exists")

// CheckLaser : la source choisie (CO2 ou Fibre) doit exister sur la machine.
func CheckLaser(o Options) error {
	if o.Fiber && !o.Machine.HasFiber() {
		return fmt.Errorf(T("err.noFiber"), o.Machine.Name)
	}
	if !o.Fiber && !o.Machine.HasCO2() {
		return fmt.Errorf(T("err.noCO2"), o.Machine.Name)
	}
	return nil
}

func FindJobFolder(db DB, name string) (int64, error) {
	rows, err := db.Query(`SELECT ID FROM MaterialFolders WHERE ForJobSettings=1 AND Name=?`, name)
	if err != nil || len(rows) == 0 {
		return 0, err
	}
	return toInt(rows[0][0]), nil
}

func subfolderIDs(db DB) map[int]int64 {
	ids := map[int]int64{SubMetal: 1, SubBois: 2, SubPlastique: 3}
	rows, err := db.Query(`SELECT ID, Name FROM MaterialSubfolders`)
	if err != nil {
		return ids
	}
	found := map[string]int64{}
	for _, r := range rows {
		found[foldKey(toStr(r[1]))] = toInt(r[0])
	}
	// noms acceptés : toutes les langues de l'interface
	for sub, key := range map[int]string{SubMetal: "sub.metal", SubBois: "sub.wood", SubPlastique: "sub.plastic"} {
		for _, n := range tr[key] {
			if v, ok := found[foldKey(n)]; ok {
				ids[sub] = v
				break
			}
		}
	}
	return ids
}

func ensureSubfolders(db DB) map[int]int64 {
	ids := subfolderIDs(db)
	// crée les sous-dossiers manquants (base neuve)
	names := map[int]string{SubMetal: T("sub.metal"), SubBois: T("sub.wood"), SubPlastique: T("sub.plastic")}
	rows, _ := db.Query(`SELECT ID FROM MaterialSubfolders`)
	have := map[int64]bool{}
	for _, r := range rows {
		have[toInt(r[0])] = true
	}
	for k, id := range ids {
		if !have[id] {
			if r, err := db.Query(`SELECT IFNULL(MAX(ID),0)+1 FROM MaterialSubfolders`); err == nil && len(r) > 0 {
				nid := toInt(r[0][0])
				if db.Exec(`INSERT INTO MaterialSubfolders (ID, SortID, Name) VALUES (?,?,?)`, nid, nid, names[k]) == nil {
					ids[k] = nid
				}
			}
		}
	}
	return ids
}

// Tables et colonnes de la sauvegarde utilisées par la conversion.
var requiredSchema = []struct {
	Table string
	Cols  []string
}{
	{"Machines", []string{"ID", "Name", "LaserType", "TableSize", "SortID"}},
	{"MaterialFolders", []string{"ID", "SortID", "Name", "Subfolders", "MachineID", "ForJobSettings"}},
	{"MaterialSubfolders", []string{"ID", "SortID", "Name"}},
	{"Settings", []string{"ID", "Category", "ColorSettings", "CreationDate", "GlobalSettings", "MachineID", "Name", "Notes",
		"PrintDate", "SortID", "SubCategory", "OriginalColorSettings", "FolderID", "SubFolderID", "ProcessSettings",
		"OriginalProcessSettings", "EpilogProvided"}},
}

// MissingSchema : tables ("Table") et colonnes ("Table.Colonne") absentes de la sauvegarde.
func MissingSchema(db DB) ([]string, error) {
	var missing []string
	for _, t := range requiredSchema {
		rows, err := db.Query(`PRAGMA table_info(` + t.Table + `)`)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			missing = append(missing, t.Table)
			continue
		}
		have := map[string]bool{}
		for _, r := range rows {
			if len(r) > 1 {
				have[strings.ToLower(toStr(r[1]))] = true
			}
		}
		for _, c := range t.Cols {
			if !have[strings.ToLower(c)] {
				missing = append(missing, t.Table+"."+c)
			}
		}
	}
	return missing, nil
}

// CheckSchema : erreur explicite si la sauvegarde n'a pas la structure attendue.
func CheckSchema(db DB) error {
	missing, err := MissingSchema(db)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return fmt.Errorf(T("err.schema"), strings.Join(missing, ", "))
	}
	return nil
}

func Convert(db DB, o Options, logf func(string), progress func(done, total int)) (Result, error) {
	var res Result
	if err := CheckSchema(db); err != nil {
		return res, err
	}
	if err := CheckLaser(o); err != nil {
		return res, err
	}
	paths, err := ListDats(o.DatDir, o.Recursive)
	if err != nil {
		return res, fmt.Errorf(T("err.readDir"), err)
	}
	if len(paths) == 0 {
		return res, errors.New(T("err.noDat"))
	}
	var dats []DatFile
	seen := map[string]bool{}
	for _, p := range paths {
		d, err := ParseDat(p)
		if err != nil {
			res.Skipped = append(res.Skipped, Tf("skip.fmt", filepath.Base(p), err.Error()))
			logf(Tf("log.skipped", filepath.Base(p), err.Error()))
			continue
		}
		low := foldKey(d.Name)
		if strings.HasPrefix(low, "param1") || strings.HasPrefix(low, "param2") {
			res.Skipped = append(res.Skipped, Tf("skip.fmt", d.Name, T("skip.testGrid")))
			logf(Tf("log.skipped", d.Name, T("skip.testGridLog")))
			continue
		}
		if d.NeedsFiber() && !o.Machine.HasFiber() {
			res.Skipped = append(res.Skipped, Tf("skip.fmt", d.Name, T("skip.needsFiber")))
			logf(Tf("log.skipped", d.Name, T("skip.needsFiber")))
			continue
		}
		key := foldKey(d.Name)
		if seen[key] {
			res.Skipped = append(res.Skipped, Tf("skip.fmt", d.Name, T("skip.dup")))
			continue
		}
		seen[key] = true
		dats = append(dats, d)
	}
	if len(dats) == 0 {
		return res, errors.New(T("err.noUsable"))
	}
	logf(Tf("log.read", len(dats)))

	if err := db.Exec("BEGIN"); err != nil {
		return res, err
	}
	ok := false
	defer func() {
		if !ok {
			db.Exec("ROLLBACK")
		}
	}()

	// Dossier "Paramètres du travail"
	fid, err := FindJobFolder(db, o.FolderName)
	if err != nil {
		return res, err
	}
	if fid != 0 {
		if !o.ReplaceFolder {
			return res, ErrFolderExists
		}
		if err := db.Exec(`DELETE FROM Settings WHERE FolderID=?`, fid); err != nil {
			return res, err
		}
		logf(Tf("log.replaced", o.FolderName))
	}
	subs := ensureSubfolders(db)
	subList := fmt.Sprintf(`<list><element val="%d"/><element val="%d"/><element val="%d"/></list>`, subs[SubBois], subs[SubPlastique], subs[SubMetal])

	sortID := int64(0)
	if o.OnTop {
		r, err := db.Query(`SELECT IFNULL(MIN(SortID),1) FROM MaterialFolders WHERE ForJobSettings=1`)
		if err != nil {
			return res, err
		}
		sortID = toInt(r[0][0]) - 1
		if sortID > 0 {
			sortID = 0
		}
	} else {
		r, err := db.Query(`SELECT IFNULL(MAX(SortID),0)+1 FROM MaterialFolders`)
		if err != nil {
			return res, err
		}
		sortID = toInt(r[0][0])
	}
	if fid == 0 {
		r, err := db.Query(`SELECT IFNULL(MAX(ID),0)+1 FROM MaterialFolders`)
		if err != nil {
			return res, err
		}
		fid = toInt(r[0][0])
		if r2, err := db.Query(`SELECT IFNULL(seq,0) FROM sqlite_sequence WHERE name='MaterialFolders'`); err == nil && len(r2) > 0 && toInt(r2[0][0]) >= fid {
			fid = toInt(r2[0][0]) + 1
		}
		if err := db.Exec(`INSERT INTO MaterialFolders (ID, SortID, Name, Subfolders, MachineID, ForJobSettings) VALUES (?,?,?,?,?,1)`,
			fid, sortID, o.FolderName, subList, o.Machine.ID); err != nil {
			return res, fmt.Errorf(T("err.createFolder"), err)
		}
	} else {
		if err := db.Exec(`UPDATE MaterialFolders SET SortID=?, Subfolders=?, MachineID=? WHERE ID=?`, sortID, subList, o.Machine.ID, fid); err != nil {
			return res, err
		}
	}
	res.FolderID = fid

	r, err := db.Query(`SELECT IFNULL(MAX(ID),0) FROM Settings`)
	if err != nil {
		return res, err
	}
	sid := toInt(r[0][0])
	if r2, err := db.Query(`SELECT IFNULL(seq,0) FROM sqlite_sequence WHERE name='Settings'`); err == nil && len(r2) > 0 && toInt(r2[0][0]) > sid {
		sid = toInt(r2[0][0])
	}

	laser := 0
	if o.Fiber {
		laser = 1
	}
	air := 0
	if o.AirAssistCut {
		air = 1
	}
	H := fnum(o.Machine.H)
	W := fnum(o.Machine.W)
	now := time.Now().Unix()

	type row struct {
		sub int
		d   DatFile
	}
	var rows []row
	for _, d := range dats {
		rows = append(rows, row{Classify(d.Name), d})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].sub != rows[j].sub {
			return rows[i].sub < rows[j].sub
		}
		return foldKey(rows[i].d.Name) < foldKey(rows[j].d.Name)
	})

	for i, rw := range rows {
		d := rw.d
		en := xmlEsc(d.Name)
		th := ThicknessMils(d.Name)
		vf := fnum(math.Round(d.FreqPercent()*1e4) / 1e4)
		vsort, dith, ord := d.JMSort(o.VectorSort), d.JMDither(), b2i(d.Bottom)
		// Fusion : source Fibre du .DAT (les .DAT Fibre vers une machine sans Fibre
		// ont été écartés plus haut).
		dl := laser
		if d.Fiber {
			dl = 1
		}
		grav := proc{name: en, ptype: 1, thick: 0, rp: d.RPower, rs: d.RSpeed, dpi: d.DPI, vp: 50, vs: 50, vf: "50", air: 0, sort: vsort, pulse: "-1", laser: dl, dither: dith, order: ord}
		cut := proc{name: en, ptype: 2, thick: th, rp: 100, rs: 15, dpi: d.DPI, vp: d.VPower, vs: d.VSpeed, vf: vf, air: air, sort: vsort, pulse: "-1", laser: dl, dither: dith, order: ord}
		if dl == 1 && d.EngFreq > 0 {
			grav.rfreq = strconv.Itoa(d.EngFreq)
		}
		if d.Thickness > 0 { // épaisseur de l'autofocus Fusion, prioritaire sur le nom
			grav.thick = mmToMils(d.Thickness)
			cut.thick = grav.thick
		}
		if d.Offset != 0 {
			grav.focus = mmToMils(d.Offset)
			cut.focus = grav.focus
		}
		if d.PassesOnEngrave() {
			grav.passes = d.Passes
		} else {
			cut.passes = d.Passes
		}
		// Gestion des couleurs du driver : un procédé par couleur, placé avant
		// les procédés généraux qui reprennent les objets des autres couleurs.
		var cps string
		for _, c := range d.Colors {
			cn := en + " #" + strings.ToUpper(c.RGB)
			cl := dl
			if c.Fiber {
				cl = 1
			}
			setFusion := func(p *proc) {
				p.laser = cl
				if c.Thick > 0 {
					p.thick = c.Thick
				}
				if c.Offset != 0 {
					p.focus = c.Offset
				}
			}
			if c.Cut && d.Mode != 1 {
				p := cut
				p.name, p.vp, p.vs, p.passes = cn, c.Power, c.Speed, 1
				cf := FreqToPercent(c.Freq)
				if d.Layout.FreqPct {
					cf = math.Max(1, math.Min(100, float64(c.Freq)))
				}
				p.vf = fnum(math.Round(cf*1e4) / 1e4)
				if c.HasAir {
					p.air = max(air, b2i(c.Air))
				}
				p.query = colorQuery("lte", c.RGB)
				setFusion(&p)
				cps += p.xml()
			}
			if c.Engrave && d.Mode != 0 {
				p := grav
				p.name, p.rp, p.rs, p.passes = cn, c.Power, c.Speed, 1
				p.query = colorQuery("gt", c.RGB)
				setFusion(&p)
				if cl == 1 && d.Layout.FreqPct {
					p.rfreq = strconv.Itoa(c.Freq)
				}
				cps += p.xml()
			}
		}
		var ps string
		// rasterType = type de gravure (0 Standard, 1 Tampon, 2 3D) :
		// la V1.1 mettait 1 sur les combinés, ce qui les passait en Tampon.
		rt := d.JMRasterType()
		switch d.Mode {
		case 1:
			grav.query = qAll
			ps = "<list>" + cps + grav.xml() + "</list>\n"
		case 0:
			cut.query = qAll
			ps = "<list>" + cps + cut.xml() + "</list>\n"
			rt = 0
		default:
			grav.query, cut.query = qGT, qLTE
			ps = "<list>" + cps + grav.xml() + cut.xml() + "</list>\n"
		}
		// Compensations : valeurs habituelles des réglages Job Manager, sauf Fusion
		// où elles viennent du .DAT (Speed Comp. / Power Comp.).
		pcomp, scomp := 1, 0
		if d.Known() && d.Layout.Name == "fusion" {
			pcomp, scomp = b2i(d.PowerComp), b2i(d.SpeedComp)
		}
		// Taille de page : celle de la machine si Job Manager la connaît,
		// sinon celle enregistrée par le driver dans le .DAT.
		pw, ph := W, H
		if !o.Machine.SizeKnown && d.PageW > 0 {
			pw, ph = fnum(d.PageW), fnum(d.PageH)
		}
		g := fmt.Sprintf(globalTpl, b2i(d.AutoFocus), d.JMCenter(), ph, ord, d.DPI, rt, b2i(d.VectorGrid), pcomp, scomp, pw)
		sid++
		subID := int64(0)
		if rw.sub != SubRoot {
			subID = subs[rw.sub]
		}
		if err := db.Exec(`INSERT INTO Settings (ID, Category, ColorSettings, CreationDate, GlobalSettings, MachineID, Name, Notes, PrintDate, SortID, SubCategory, OriginalColorSettings, FolderID, SubFolderID, ProcessSettings, OriginalProcessSettings, EpilogProvided)
			VALUES (?, NULL, NULL, ?, ?, NULL, ?, '', NULL, ?, NULL, NULL, ?, ?, ?, ?, NULL)`,
			sid, now, g, d.Name, sid, fid, subID, ps, origProcesses(d.Mode)); err != nil {
			return res, fmt.Errorf(T("err.addSetting"), d.Name, err)
		}
		res.Converted = append(res.Converted, d)
		if progress != nil {
			progress(i+1, len(rows))
		}
	}
	for _, t := range []string{"Settings", "MaterialFolders", "MaterialSubfolders"} {
		db.Exec(`UPDATE sqlite_sequence SET seq=(SELECT MAX(ID) FROM `+t+`) WHERE name=?`, t)
	}
	if err := db.Exec("COMMIT"); err != nil {
		return res, err
	}
	ok = true
	ic, err := db.Query(`PRAGMA integrity_check`)
	if err != nil || len(ic) == 0 || toStr(ic[0][0]) != "ok" {
		return res, errors.New(T("err.integrity"))
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// Récapitulatif CSV (séparateur et décimale selon la langue : ; et , pour
// l'Excel européen, , et . pour l'anglais, le chinois et le japonais)

func WriteRecap(path string, o Options, res Result) error {
	sep, dec := csvSep(), csvDec()
	esc := func(f string) string {
		if strings.ContainsAny(f, sep+"\"") {
			return `"` + strings.ReplaceAll(f, `"`, `""`) + `"`
		}
		return f
	}
	var sb strings.Builder
	sb.WriteString("\ufeff")
	sb.WriteString(strings.ReplaceAll(T("csv.cols"), "|", sep) + "\r\n")
	subName := map[int]string{SubRoot: T("sub.root"), SubMetal: T("sub.metal"), SubBois: T("sub.wood"), SubPlastique: T("sub.plastic")}
	for _, d := range res.Converted {
		rem := []string{}
		if d.Dither == 6 && !d.Known() {
			rem = append(rem, T("rem.photo"))
		}
		if d.Stamp {
			rem = append(rem, T("rem.stamp"))
		}
		if d.Passes > 1 {
			rem = append(rem, Tf("rem.passes", d.Passes))
		}
		if len(d.Colors) > 0 {
			rem = append(rem, Tf("rem.colors", len(d.Colors)))
		}
		gs, gp, cs, cp, cf := "", "", "", "", ""
		if d.Mode != 0 {
			gs, gp = strconv.Itoa(d.RSpeed), strconv.Itoa(d.RPower)
		}
		if d.Mode != 1 {
			cs, cp = strconv.Itoa(d.VSpeed), strconv.Itoa(d.VPower)
			cf = strings.ReplaceAll(strconv.FormatFloat(d.FreqPercent(), 'f', 1, 64), ".", dec)
		}
		th := ""
		if t := ThicknessMils(d.Name); t > 0 && d.Mode != 1 {
			th = strings.ReplaceAll(strconv.FormatFloat(t*25.4/1000, 'f', 2, 64), ".", dec)
		}
		f := []string{subName[Classify(d.Name)], d.Name, d.ModeLabel(), strconv.Itoa(d.DPI), gs, gp, cs, cp, cf, th, strings.Join(rem, " / ")}
		for i := range f {
			f[i] = esc(f[i])
		}
		sb.WriteString(strings.Join(f, sep) + "\r\n")
	}
	for _, s := range res.Skipped {
		f := []string{"", esc(s), T("csv.skipped"), "", "", "", "", "", "", "", ""}
		sb.WriteString(strings.Join(f, sep) + "\r\n")
	}
	return os.WriteFile(path, []byte(sb.String()), 0o644)
}

func CopyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

// OutputDir : "<dossier DAT> - Conversion Job Manager" à côté du dossier source.
func OutputDir(datDir string) string {
	parent := filepath.Dir(filepath.Clean(datDir))
	base := filepath.Base(filepath.Clean(datDir)) + T("out.suffix")
	p := filepath.Join(parent, base)
	for i := 2; ; i++ {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return p
		}
		p = filepath.Join(parent, fmt.Sprintf("%s (%d)", base, i))
	}
}

var reCopyNum = regexp.MustCompile(` \(\d+\)$`)

// SourceDirFor : si dir est un dossier de résultat sans .DAT (suffixe de sortie,
// dans n'importe quelle langue, éventuellement suivi de « (2) »…), renvoie le
// dossier source voisin s'il contient des .DAT ; sinon "".
func SourceDirFor(dir string, recursive bool) string {
	dir = filepath.Clean(dir)
	if ps, err := ListDats(dir, recursive); err != nil || len(ps) > 0 {
		return ""
	}
	base := reCopyNum.ReplaceAllString(filepath.Base(dir), "")
	for _, suf := range tr["out.suffix"] {
		if suf == "" || len(base) <= len(suf) || !strings.EqualFold(base[len(base)-len(suf):], suf) {
			continue
		}
		src := filepath.Join(filepath.Dir(dir), base[:len(base)-len(suf)])
		if ps, err := ListDats(src, recursive); err == nil && len(ps) > 0 {
			return src
		}
	}
	return ""
}

func safeFileName(s string) string {
	r := strings.NewReplacer(`\`, "-", "/", "-", ":", "-", "*", "", "?", "", `"`, "", "<", "", ">", "", "|", "")
	return strings.TrimSpace(r.Replace(s))
}

// Run : toute la chaîne (copie de la sauvegarde, conversion, fichiers annexes).
func Run(open func(path string) (DB, error), backup string, o Options, logf func(string), progress func(int, int)) (string, Result, error) {
	out := OutputDir(o.DatDir)
	if err := os.MkdirAll(out, 0o755); err != nil {
		return "", Result{}, fmt.Errorf(T("err.mkOut"), err)
	}
	bname := strings.TrimSuffix(filepath.Base(backup), filepath.Ext(backup))
	dbName := safeFileName(bname+" - "+o.FolderName) + ".epidb"
	dst := filepath.Join(out, dbName)
	logf(T("log.copy"))
	if err := CopyFile(backup, dst); err != nil {
		os.RemoveAll(out)
		return "", Result{}, fmt.Errorf(T("err.copy"), err)
	}
	db, err := open(dst)
	if err != nil {
		os.RemoveAll(out)
		return "", Result{}, err
	}
	res, err := Convert(db, o, logf, progress)
	db.Close()
	if err != nil {
		os.RemoveAll(out)
		return "", res, err
	}
	recap := T("file.recap")
	if err := WriteRecap(filepath.Join(out, recap), o, res); err != nil {
		logf(T("log.recapFail") + err.Error())
	}
	txt := fmt.Sprintf(readmeText(), dbName, o.FolderName, recap, o.FolderName, o.Machine.Name, time.Now().Format(T("fmt.date")))
	os.WriteFile(filepath.Join(out, T("file.readme")), []byte("\ufeff"+strings.ReplaceAll(txt, "\n", "\r\n")), 0o644)
	return out, res, nil
}
