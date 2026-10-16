//go:build windows

package main

// Informations système du rapport d'erreur : version de Windows, programme,
// bibliothèque SQLite utilisée (celle de Windows ou un sqlite3.dll à côté du
// programme, qui passe en priorité).

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	pRtlGetVersion      = syscall.NewLazyDLL("ntdll.dll").NewProc("RtlGetVersion")
	pGetModuleFileNameW = kernel32.NewProc("GetModuleFileNameW")
)

type osVersionInfo struct {
	size, major, minor, build, platform uint32
	csd                                 [128]uint16
}

func windowsVersion() string {
	v := osVersionInfo{size: uint32(unsafe.Sizeof(osVersionInfo{}))}
	if r, _, _ := pRtlGetVersion.Call(uintptr(unsafe.Pointer(&v))); r != 0 {
		return "inconnue"
	}
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.build)
}

func sqliteInfo() string {
	if sqliteDLL == nil {
		if _, err := loadSQLite(); err != nil {
			return "non chargée : " + err.Error()
		}
	}
	path := sqliteDLL.Name
	buf := make([]uint16, 1024)
	if n, _, _ := pGetModuleFileNameW.Call(sqliteDLL.Handle(), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf))); n > 0 {
		path = syscall.UTF16ToString(buf[:n])
	}
	ver := "?"
	if p := sqliteDLL.NewProc("sqlite3_libversion"); p.Find() == nil {
		r, _, _ := p.Call()
		ver = goStrZ(r)
	}
	return fmt.Sprintf("%s (%s)", path, ver)
}

func reportSystem() []string {
	exe, _ := os.Executable()
	local := "non"
	if _, err := os.Stat(filepath.Join(filepath.Dir(exe), "sqlite3.dll")); err == nil {
		local = "OUI (utilisé en priorité)"
	}
	return []string{
		"Windows : " + windowsVersion(),
		"Programme : " + exe,
		"SQLite : " + sqliteInfo(),
		"sqlite3.dll à côté du programme : " + local,
	}
}
