/*
xml2sav - converts a custom xml document to a SPSS binary file.
Copyright (C) 2016 A.J. Jessurun

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
	"bufio"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type CsvVar struct {
	Name  string
	Value string
}

type CsvWriter struct {
	*csv.Writer
	BufIO *bufio.Writer
	Dict  []*CsvVar
	Vars  map[string]*CsvVar
}

func NewCsvWriter(writer io.Writer) *CsvWriter {
	w := new(CsvWriter)
	w.BufIO = bufio.NewWriter(writer)
	w.Writer = csv.NewWriter(w.BufIO)
	w.Vars = make(map[string]*CsvVar)
	return w
}

func (c *CsvWriter) Flush() error {
	return c.BufIO.Flush()
}

func parseXSavToCsv(reader io.Reader, filename string) error {
	basename := strings.TrimSuffix(filename, filepath.Ext(filename))
	var csv *CsvWriter
	var f *os.File
	decoder := xml.NewDecoder(reader)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "sav":
				savname := getAttr(&t, "name")
				csvfilename := fmt.Sprintf("%s_%s.csv", basename, savname)
				log.Println("Writing", csvfilename)
				f, err = os.Create(csvfilename)
				if err != nil {
					return err
				}
				csv = NewCsvWriter(f)
			case "var":
				v := new(CsvVar)
				v.Name = getAttr(&t, "name")
				csv.Dict = append(csv.Dict, v)
				if _, found := csv.Vars[v.Name]; found {
					return fmt.Errorf("Variable %s already defined", v.Name)
				}
				csv.Vars[v.Name] = v
			case "case":
				for _, v := range csv.Dict {
					v.Value = ""
				}
			case "val":
				var valxml valXML
				if err = decoder.DecodeElement(&valxml, &t); err != nil {
					return err
				}
				v, found := csv.Vars[valxml.Name]
				if !found {
					return fmt.Errorf("Found a value for non existing variable %s", valxml.Name)
				}
				v.Value = valxml.Value
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "sav":
				csv.Flush()
				f.Close()
			case "dict":
				header := make([]string, len(csv.Dict))
				for i := range csv.Dict {
					header[i] = csv.Dict[i].Name
				}
				csv.Write(header)
			case "case":
				record := make([]string, len(csv.Dict))
				for i := range csv.Dict {
					record[i] = csv.Dict[i].Value
				}
				csv.Write(record)
			}
		}
	}

	return nil
}
