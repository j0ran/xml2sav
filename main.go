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
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"
)

var maxStringLength = 1024 * 50
var defaultStringLength = 2048
var maxPrintStringWidth = 40
var pause = false
var noLogToFile = false
var singlePass = false
var toCsv = false
var ignoreMissingVar = false
var register func()

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: xml2sav [options] <file.xsav>")
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}
	flag.BoolVar(&pause, "pause", pause, "pause and wait for enter after finsishing")
	flag.BoolVar(&noLogToFile, "nolog", noLogToFile, "don't write log to file")
	flag.BoolVar(&singlePass, "single", singlePass, "don't determine lengths of string variables")
	flag.BoolVar(&toCsv, "csv", toCsv, "convert to csv")
	flag.BoolVar(&ignoreMissingVar, "ignore", ignoreMissingVar, "ignore values in cases that are not declared in dictronary")
}

func main() {
	startTime := time.Now()

	fmt.Println("xml2sav 2.1  Copyright (C) 2016-2017  A.J. Jessurun")
	fmt.Println("This program comes with ABSOLUTELY NO WARRANTY.")
	fmt.Println("This is free software, and you are welcome to redistribute it")
	fmt.Println("under certain conditions. See the file COPYING.txt.")

	flag.Parse()
	if len(flag.Args()) != 1 {
		if register != nil { // register file association
			register()
		}
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}
	filename := flag.Arg(0)

	if !noLogToFile {
		logfile, err := os.Create(filename[:len(filename)-len(path.Ext(filename))] + ".log")
		if err != nil {
			log.Fatalln(err)
		}
		log.SetOutput(io.MultiWriter(os.Stderr, logfile))
	}

	in, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer in.Close()

	log.Println("Reading", filename)
	var lengths VarLengths
	if !singlePass && !toCsv {
		log.Println("Pass 1, determining maximum length of strings")
		if lengths, err = findVarLengths(in); err != nil {
			log.Fatalln(err)
		}
		in.Seek(0, io.SeekStart) // Rewind for second read
		log.Println("Pass 2, generating sav files")
	}

	if toCsv {
		if err = parseXSavToCsv(in, filename); err != nil {
			log.Fatalln(err)
		}
	} else {
		if err = parseXSav(in, filename, lengths); err != nil {
			log.Fatalln(err)
		}
	}

	log.Printf("Done in %v\n", time.Now().Sub(startTime))

	if pause {
		fmt.Println("Press enter to continue.")
		var line string
		fmt.Scanln(&line)
	}
}
