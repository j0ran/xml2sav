// +build windows

package main

import (
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

func getExePath() (string, error) {
	var GetModuleFileNameW = syscall.MustLoadDLL("kernel32.dll").MustFindProc("GetModuleFileNameW")
	b := make([]uint16, syscall.MAX_PATH)
	r, _, err := GetModuleFileNameW.Call(0, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)))
	n := uint32(r)
	if n == 0 {
		return "", err
	}
	return string(utf16.Decode(b[0:n])), nil
}

func setDefaultRegKey(key registry.Key, path string, value string) {
	k, _, err := registry.CreateKey(key, path, registry.ALL_ACCESS)
	if err != nil {
		return
	}
	defer k.Close()

	err = k.SetStringValue("", value)
	if err != nil {
		return
	}
}

func registerWindows() {
	exe, err := getExePath()
	if err != nil {
		return
	}

	setDefaultRegKey(registry.CURRENT_USER, `Software\Classes\xml2sav.v2\shell\open\command`, `"`+exe+`" -pause "%1"`)
	setDefaultRegKey(registry.CURRENT_USER, `Software\Classes\.xsav`, "xml2sav.v2")
}

func init() {
	register = registerWindows
}
