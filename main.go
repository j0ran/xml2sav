package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var defaultStringLength = 255
var maxPrintStringWidth = 40
var pause = false

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: xml2sav [options] <file.xsav>")
		flag.PrintDefaults()
	}
	flag.BoolVar(&pause, "pause", pause, "pause and wait for enter after finsishing")
}

func main() {
	startTime := time.Now()

	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}
	filename := flag.Arg(0)

	fmt.Println("xml2sav 2.0  Copyright (C) 2009-2016  A.J. Jessurun")
	fmt.Println("This program comes with ABSOLUTELY NO WARRANTY.")
	fmt.Println("This is free software, and you are welcome to redistribute it")
	fmt.Println("under certain conditions. See the file COPYING.txt.")

	in, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer in.Close()

	log.Println("Reading", filename)

	if err = parseXSav(in, filename); err != nil {
		log.Fatalln(err)
	}

	log.Printf("Done in %v\n", time.Now().Sub(startTime))

	if pause {
		fmt.Println("Press enter to continue.")
		var line string
		fmt.Scanln(&line)
	}
}
