//go:build windows

package main

// Accès SQLite via winsqlite3.dll, fournie avec Windows 10/11 et chargée
// par son chemin complet dans le dossier système : un sqlite3.dll ou
// winsqlite3.dll déposé à côté du programme n'est jamais chargé.

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"unsafe"
)

type sqliteLib struct {
	openV2, prepareV2, step, finalize, close2, errmsg                *syscall.LazyProc
	bindInt64, bindText, bindNull, colCount, colType                 *syscall.LazyProc
	colInt64, colText, colBytes, colBlob, busyTimeout, extendedCodes *syscall.LazyProc
}

var sqlib *sqliteLib

var pGetSystemDirectoryW = kernel32.NewProc("GetSystemDirectoryW")

// DLL SQLite effectivement chargée (indiquée dans le rapport d'erreur)
var sqliteDLL *syscall.LazyDLL

func loadSQLite() (*sqliteLib, error) {
	if sqlib != nil {
		return sqlib, nil
	}
	buf := make([]uint16, 260)
	n, _, _ := pGetSystemDirectoryW.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 || int(n) >= len(buf) {
		return nil, errors.New(T("err.noSqlite"))
	}
	dll := syscall.NewLazyDLL(filepath.Join(syscall.UTF16ToString(buf[:n]), "winsqlite3.dll"))
	if dll.Load() != nil {
		return nil, errors.New(T("err.noSqlite"))
	}
	l := &sqliteLib{
		openV2: dll.NewProc("sqlite3_open_v2"), prepareV2: dll.NewProc("sqlite3_prepare_v2"),
		step: dll.NewProc("sqlite3_step"), finalize: dll.NewProc("sqlite3_finalize"),
		close2: dll.NewProc("sqlite3_close_v2"), errmsg: dll.NewProc("sqlite3_errmsg"),
		bindInt64: dll.NewProc("sqlite3_bind_int64"), bindText: dll.NewProc("sqlite3_bind_text"),
		bindNull: dll.NewProc("sqlite3_bind_null"), colCount: dll.NewProc("sqlite3_column_count"),
		colType: dll.NewProc("sqlite3_column_type"), colInt64: dll.NewProc("sqlite3_column_int64"),
		colText: dll.NewProc("sqlite3_column_text"), colBytes: dll.NewProc("sqlite3_column_bytes"),
		colBlob: dll.NewProc("sqlite3_column_blob"), busyTimeout: dll.NewProc("sqlite3_busy_timeout"),
		extendedCodes: dll.NewProc("sqlite3_extended_result_codes"),
	}
	for _, p := range []*syscall.LazyProc{l.openV2, l.prepareV2, l.step, l.finalize, l.close2, l.errmsg,
		l.bindInt64, l.bindText, l.bindNull, l.colCount, l.colType, l.colInt64, l.colText, l.colBytes, l.colBlob} {
		if err := p.Find(); err != nil {
			return nil, fmt.Errorf(T("err.sqliteIncomplete"), err)
		}
	}
	sqlib, sqliteDLL = l, dll
	return l, nil
}

type winDB struct {
	l  *sqliteLib
	db uintptr
}

func cstr(s string) []byte { return append([]byte(s), 0) }

func goStrN(p uintptr, n int) string {
	if p == 0 || n <= 0 {
		return ""
	}
	return string(unsafe.Slice((*byte)(unsafe.Pointer(p)), n))
}

func goStrZ(p uintptr) string {
	if p == 0 {
		return ""
	}
	n := 0
	for *(*byte)(unsafe.Pointer(p + uintptr(n))) != 0 {
		n++
	}
	return goStrN(p, n)
}

const (
	sqliteOK      = 0
	sqliteRow     = 100
	sqliteDone    = 101
	openReadOnly  = 0x1
	openReadWrite = 0x2
	transient     = ^uintptr(0)
	typeInteger   = 1
	typeFloat     = 2
	typeText      = 3
	typeBlob      = 4
	typeNull      = 5
)

func openSQLite(path string, readonly bool) (DB, error) {
	l, err := loadSQLite()
	if err != nil {
		return nil, err
	}
	var h uintptr
	p := cstr(path)
	flags := uintptr(openReadWrite)
	if readonly {
		flags = openReadOnly
	}
	rc, _, _ := l.openV2.Call(uintptr(unsafe.Pointer(&p[0])), uintptr(unsafe.Pointer(&h)), flags, 0)
	runtime.KeepAlive(p)
	if int32(rc) != sqliteOK {
		msg := "ouverture impossible"
		if h != 0 {
			r, _, _ := l.errmsg.Call(h)
			msg = goStrZ(r)
			l.close2.Call(h)
		}
		return nil, fmt.Errorf(T("err.database"), msg)
	}
	if l.busyTimeout.Find() == nil {
		l.busyTimeout.Call(h, 5000)
	}
	return &winDB{l, h}, nil
}

func openRW(path string) (DB, error) { return openSQLite(path, false) }
func openRO(path string) (DB, error) { return openSQLite(path, true) }

func (w *winDB) err(what string) error {
	r, _, _ := w.l.errmsg.Call(w.db)
	return fmt.Errorf("%s : %s", what, goStrZ(r))
}

func (w *winDB) run(sql string, args []any, collect bool) ([][]any, error) {
	q := cstr(sql)
	var st uintptr
	rc, _, _ := w.l.prepareV2.Call(w.db, uintptr(unsafe.Pointer(&q[0])), uintptr(len(q)), uintptr(unsafe.Pointer(&st)), 0)
	runtime.KeepAlive(q)
	if int32(rc) != sqliteOK {
		return nil, w.err("requête")
	}
	defer w.l.finalize.Call(st)
	keep := make([][]byte, 0, len(args))
	for i, a := range args {
		idx := uintptr(i + 1)
		var rc uintptr
		switch v := a.(type) {
		case nil:
			rc, _, _ = w.l.bindNull.Call(st, idx)
		case int:
			rc, _, _ = w.l.bindInt64.Call(st, idx, uintptr(int64(v)))
		case int64:
			rc, _, _ = w.l.bindInt64.Call(st, idx, uintptr(v))
		case string:
			b := []byte(v)
			if len(b) == 0 {
				b = []byte{0}
				rc, _, _ = w.l.bindText.Call(st, idx, uintptr(unsafe.Pointer(&b[0])), 0, transient)
			} else {
				rc, _, _ = w.l.bindText.Call(st, idx, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), transient)
			}
			keep = append(keep, b)
		default:
			s := fmt.Sprint(v)
			b := append([]byte(s), 0)
			rc, _, _ = w.l.bindText.Call(st, idx, uintptr(unsafe.Pointer(&b[0])), uintptr(len(s)), transient)
			keep = append(keep, b)
		}
		if int32(rc) != sqliteOK {
			return nil, w.err("paramètre")
		}
	}
	runtime.KeepAlive(keep)
	var rows [][]any
	for {
		rc, _, _ := w.l.step.Call(st)
		switch int32(rc) & 0xff {
		case sqliteRow:
			if !collect {
				continue
			}
			n, _, _ := w.l.colCount.Call(st)
			row := make([]any, int(int32(n)))
			for c := range row {
				ci := uintptr(c)
				t, _, _ := w.l.colType.Call(st, ci)
				switch int32(t) {
				case typeInteger:
					v, _, _ := w.l.colInt64.Call(st, ci)
					row[c] = int64(v)
				case typeNull:
					row[c] = nil
				case typeFloat:
					p, _, _ := w.l.colText.Call(st, ci)
					nb, _, _ := w.l.colBytes.Call(st, ci)
					f, _ := strconv.ParseFloat(goStrN(p, int(int32(nb))), 64)
					row[c] = f
				case typeBlob:
					p, _, _ := w.l.colBlob.Call(st, ci)
					nb, _, _ := w.l.colBytes.Call(st, ci)
					row[c] = []byte(goStrN(p, int(int32(nb))))
				default:
					p, _, _ := w.l.colText.Call(st, ci)
					nb, _, _ := w.l.colBytes.Call(st, ci)
					row[c] = goStrN(p, int(int32(nb)))
				}
			}
			rows = append(rows, row)
		case sqliteDone:
			return rows, nil
		default:
			return nil, w.err("exécution")
		}
	}
}

func (w *winDB) Query(sql string, args ...any) ([][]any, error) { return w.run(sql, args, true) }
func (w *winDB) Exec(sql string, args ...any) error             { _, err := w.run(sql, args, false); return err }
func (w *winDB) Close() error {
	w.l.close2.Call(w.db)
	return nil
}
