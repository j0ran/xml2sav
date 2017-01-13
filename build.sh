#!/bin/sh
GOOS=windows GOARCH=amd64 go build -o xml2sav-win64.exe
GOOS=windows GOARCH=386 go build -o xml2sav-win32.exe
GOOS=linux GOARCH=amd64 go build -o xml2sav-linux64
