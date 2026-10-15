//go:build linux

package main

// Implémentation de test (Linux uniquement) : pont vers le module sqlite3 de Python.

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
)

const pyBridge = `
import sys, json, sqlite3
db = sqlite3.connect(sys.argv[1], isolation_level=None)
for line in sys.stdin:
    req = json.loads(line)
    try:
        cur = db.execute(req["sql"], req["args"])
        rows = []
        for r in cur.fetchall():
            rows.append([x.decode("utf-8","replace") if isinstance(x, bytes) else x for x in r])
        out = {"rows": rows}
    except Exception as e:
        out = {"error": str(e)}
    sys.stdout.write(json.dumps(out) + "\n"); sys.stdout.flush()
`

type pyDB struct {
	cmd *exec.Cmd
	in  io.WriteCloser
	out *bufio.Reader
}

func openPyDB(path string) (DB, error) {
	cmd := exec.Command("python3", "-c", pyBridge, path)
	in, _ := cmd.StdinPipe()
	o, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &pyDB{cmd, in, bufio.NewReaderSize(o, 1<<20)}, nil
}

func (p *pyDB) call(sql string, args []any) ([][]any, error) {
	if args == nil {
		args = []any{}
	}
	b, _ := json.Marshal(map[string]any{"sql": sql, "args": args})
	p.in.Write(append(b, '\n'))
	line, err := p.out.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	var r struct {
		Rows  [][]any
		Error string
	}
	d := json.NewDecoder(bytesReader(line))
	d.UseNumber()
	if err := d.Decode(&r); err != nil {
		return nil, err
	}
	if r.Error != "" {
		return nil, errors.New(r.Error)
	}
	for _, row := range r.Rows {
		for i, v := range row {
			if n, ok := v.(json.Number); ok {
				if iv, err := n.Int64(); err == nil {
					row[i] = iv
				} else {
					f, _ := n.Float64()
					row[i] = f
				}
			}
		}
	}
	return r.Rows, nil
}

func (p *pyDB) Query(sql string, args ...any) ([][]any, error) { return p.call(sql, args) }
func (p *pyDB) Exec(sql string, args ...any) error             { _, err := p.call(sql, args); return err }
func (p *pyDB) Close() error                                   { p.in.Close(); return p.cmd.Wait() }

type br struct {
	b []byte
	i int
}

func (r *br) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}
func bytesReader(b []byte) io.Reader { return &br{b: b} }

func main() {
	dat := flag.String("dat", "", "dossier DAT")
	bk := flag.String("backup", "", "backup")
	mid := flag.Int64("machine", 0, "machine id")
	lg := flag.String("lang", "fr", "langue (fr, en, es, de, it, pt, nl, zh, ja)")
	flag.Parse()
	if l, ok := langFromCode(*lg); ok {
		cur = l
	}
	ro, err := openPyDB(*bk)
	if err != nil {
		panic(err)
	}
	ms, err := LoadMachines(ro)
	ro.Close()
	if err != nil {
		panic(err)
	}
	var m Machine
	for _, x := range ms {
		fmt.Printf("machine %d %q %q %.2fx%.2f\n", x.ID, x.Name, x.LaserType, x.W, x.H)
		if x.ID == *mid {
			m = x
		}
	}
	o := Options{DatDir: *dat, Machine: m, FolderName: m.Name, AirAssistCut: true, VectorSort: 1, OnTop: true}
	out, res, err := Run(openPyDB, *bk, o, func(s string) { fmt.Println(s) }, nil)
	fmt.Println("OUT", out, len(res.Converted), len(res.Skipped), err)
}
