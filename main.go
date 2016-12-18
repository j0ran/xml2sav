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
}

func main() {
	startTime := time.Now()

	flag.Parse()
	if len(flag.Args()) != 1 {
		if register != nil { // register file association
			register()
		}
		flag.Usage()
		os.Exit(1)
	}
	filename := flag.Arg(0)

	fmt.Println("xml2sav 2.0  Copyright (C) 2009-2016  A.J. Jessurun")
	fmt.Println("This program comes with ABSOLUTELY NO WARRANTY.")
	fmt.Println("This is free software, and you are welcome to redistribute it")
	fmt.Println("under certain conditions. See the file COPYING.txt.")

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
	if !singlePass {
		log.Println("Pass 1, determining maximum length of strings")
		if lengths, err = findVarLengths(in); err != nil {
			log.Fatalln(err)
		}
		in.Seek(0, io.SeekStart) // Rewind for second read
		log.Println("Pass 2, generating sav files")
	}

	if err = parseXSav(in, filename, lengths); err != nil {
		log.Fatalln(err)
	}

	log.Printf("Done in %v\n", time.Now().Sub(startTime))

	if pause {
		fmt.Println("Press enter to continue.")
		var line string
		fmt.Scanln(&line)
	}
}

// Support very long string
// Two pass, determine string lengths
