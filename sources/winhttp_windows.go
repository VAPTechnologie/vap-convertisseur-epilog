//go:build windows

package main

// GET par WinHTTP (winhttp.dll) : évite de lier net/http et crypto/tls, qui doublaient la taille de l'exe.

import (
	"errors"
	"net/url"
	"strconv"
	"syscall"
	"unsafe"
)

var (
	winhttp                 = syscall.NewLazyDLL("winhttp.dll")
	pWinHttpOpen            = winhttp.NewProc("WinHttpOpen")
	pWinHttpSetTimeouts     = winhttp.NewProc("WinHttpSetTimeouts")
	pWinHttpConnect         = winhttp.NewProc("WinHttpConnect")
	pWinHttpOpenRequest     = winhttp.NewProc("WinHttpOpenRequest")
	pWinHttpSendRequest     = winhttp.NewProc("WinHttpSendRequest")
	pWinHttpReceiveResponse = winhttp.NewProc("WinHttpReceiveResponse")
	pWinHttpQueryHeaders    = winhttp.NewProc("WinHttpQueryHeaders")
	pWinHttpReadData        = winhttp.NewProc("WinHttpReadData")
	pWinHttpCloseHandle     = winhttp.NewProc("WinHttpCloseHandle")
)

const (
	winhttpAutoProxy  = 4               // WINHTTP_ACCESS_TYPE_AUTOMATIC_PROXY
	winhttpSecure     = 0x00800000      // WINHTTP_FLAG_SECURE
	winhttpStatusCode = 19 | 0x20000000 // WINHTTP_QUERY_STATUS_CODE | WINHTTP_QUERY_FLAG_NUMBER
	httpTimeoutMs     = 5000
	httpMaxBody       = 1 << 20
)

// httpGet : corps de la réponse (statut 200 seulement, 1 Mo au plus).
func httpGet(addr, agent, headers string) ([]byte, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	var flags uintptr
	port := 80
	switch u.Scheme {
	case "https":
		flags, port = winhttpSecure, 443
	case "http":
	default:
		return nil, errors.New("schéma non pris en charge : " + u.Scheme)
	}
	if p := u.Port(); p != "" {
		if port, err = strconv.Atoi(p); err != nil {
			return nil, err
		}
	}

	ses, _, err := pWinHttpOpen.Call(uintptr(unsafe.Pointer(w16(agent))), winhttpAutoProxy, 0, 0, 0)
	if ses == 0 {
		return nil, err
	}
	defer pWinHttpCloseHandle.Call(ses)
	pWinHttpSetTimeouts.Call(ses, httpTimeoutMs, httpTimeoutMs, httpTimeoutMs, httpTimeoutMs)
	con, _, err := pWinHttpConnect.Call(ses, uintptr(unsafe.Pointer(w16(u.Hostname()))), uintptr(port), 0)
	if con == 0 {
		return nil, err
	}
	defer pWinHttpCloseHandle.Call(con)
	req, _, err := pWinHttpOpenRequest.Call(con, uintptr(unsafe.Pointer(w16("GET"))),
		uintptr(unsafe.Pointer(w16(u.RequestURI()))), 0, 0, 0, flags)
	if req == 0 {
		return nil, err
	}
	defer pWinHttpCloseHandle.Call(req)

	h := syscall.StringToUTF16(headers)
	if ok, _, err := pWinHttpSendRequest.Call(req, uintptr(unsafe.Pointer(&h[0])), uintptr(len(h)-1), 0, 0, 0, 0); ok == 0 {
		return nil, err
	}
	if ok, _, err := pWinHttpReceiveResponse.Call(req, 0); ok == 0 {
		return nil, err
	}
	var status, size uint32 = 0, 4
	if ok, _, err := pWinHttpQueryHeaders.Call(req, winhttpStatusCode, 0,
		uintptr(unsafe.Pointer(&status)), uintptr(unsafe.Pointer(&size)), 0); ok == 0 {
		return nil, err
	}
	if status != 200 {
		return nil, errors.New("HTTP " + strconv.Itoa(int(status)))
	}

	var body []byte
	buf := make([]byte, 16<<10)
	for {
		var n uint32
		if ok, _, err := pWinHttpReadData.Call(req, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)),
			uintptr(unsafe.Pointer(&n))); ok == 0 {
			return nil, err
		}
		if n == 0 {
			return body, nil
		}
		body = append(body, buf[:n]...)
		if len(body) > httpMaxBody {
			return nil, errors.New("réponse trop longue")
		}
	}
}
