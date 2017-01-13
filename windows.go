//+build windows

/*
xml2sav - converts a custom xml document to a SPSS binary file.
Copyright (C) 2016-2017 A.J. Jessurun

This file is part of xml2sav.

Xml2sav is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Xml2sav is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with xml2sav.  If not, see <http://www.gnu.org/licenses/>.
*/

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
