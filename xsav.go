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
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type labelXML struct {
	Value string `xml:"value,attr"`
	Desc  string `xml:",chardata"`
}

type varXML struct {
	Type     string      `xml:"type,attr"`
	Name     string      `xml:"name,attr"`
	Measure  string      `xml:"measure,attr"`
	Decimals byte        `xml:"decimals,attr"`
	Width    int         `xml:"width,attr"`
	Label    string      `xml:"label,attr"`
	Default  string      `xml:"default,attr"`
	Labels   []*labelXML `xml:"label"`
}

type valXML struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

func getAttr(element *xml.StartElement, name string) string {
	for _, a := range element.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	log.Fatalf("%s element does not have a %s attribute\n", element.Name.Local, name)
	return ""
}

func hasAttr(element *xml.StartElement, name string) bool {
	for _, a := range element.Attr {
		if a.Name.Local == name {
			return true
		}
	}
	return false
}

func parseXSav(in io.Reader, basename string, lengths VarLengths) error {
	bareBasename := strings.TrimSuffix(basename, filepath.Ext(basename))
	var filename string
	var f *os.File
	var out *SpssWriter
	var dictDone bool
	var savname string

	decoder := xml.NewDecoder(in)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "sav":
				savname = getAttr(&t, "name")
				filename = fmt.Sprintf("%s_%s.sav", bareBasename, savname)
				f, err = os.Create(filename)
				if err != nil {
					return err
				}
				out = NewSpssWriter(f)
				log.Println("Writing", filename)
			case "var":
				if dictDone {
					return errors.New("Adding variables while the dictionary already finished")
				}
				if out == nil {
					return errors.New("Adding variables without knowing to which sav file they belong")
				}

				varxml := new(varXML)
				if err = decoder.DecodeElement(varxml, &t); err != nil {
					return err
				}

				v := new(Var)
				v.Name = varxml.Name
				v.Type = SPSS_NUMERIC
				v.Measure = SPSS_MLVL_NOM
				switch varxml.Type {
				case "numeric":
					v.Decimals = varxml.Decimals
					v.Print = SPSS_FMT_F
					v.Width = 8
					if hasAttr(&t, "width") {
						v.Width = byte(varxml.Width)
					}
					v.Decimals = 2
					if hasAttr(&t, "decimals") {
						v.Decimals = byte(varxml.Decimals)
					}
				case "date":
					v.Print = SPSS_FMT_DATE
					v.Width = 11
					v.Decimals = 0
					v.Measure = SPSS_MLVL_RAT
				case "datetime":
					v.Print = SPSS_FMT_DATE_TIME
					v.Width = 20
					v.Decimals = 0
					v.Measure = SPSS_MLVL_RAT
				default: // string
					width := defaultStringLength
					if hasAttr(&t, "width") {
						width = varxml.Width
					} else if lengths != nil {
						width, err = lengths.GetVarLength(savname, v.Name)
						if err != nil {
							return err
						}
					}
					v.Type = int32(width)
					v.Print = SPSS_FMT_A
					v.Width = byte(width)
					if width > 40 {
						v.Width = 40
					}
					v.Decimals = 0
				}
				v.Default = varxml.Default
				v.HasDefault = hasAttr(&t, "default")
				v.Label = varxml.Label
				if hasAttr(&t, "measure") {
					switch varxml.Measure {
					case "scale":
						v.Measure = SPSS_MLVL_RAT
					case "nominal":
						v.Measure = SPSS_MLVL_NOM
					case "ordinal":
						v.Measure = SPSS_MLVL_ORD
					default:
						return fmt.Errorf("Unknown value for measure %s", varxml.Measure)
					}
				}
				for _, l := range varxml.Labels {
					v.Labels = append(v.Labels, Label{l.Value, l.Desc})
				}
				out.AddVar(v)
			case "case":
				out.ClearCase()
			case "val":
				var valxml valXML
				if err = decoder.DecodeElement(&valxml, &t); err != nil {
					return err
				}
				out.SetVar(valxml.Name, valxml.Value)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "dict":
				dictDone = true
				out.Start(fmt.Sprintf("Export with xml2sav: %s", basename))
			case "case":
				out.WriteCase()
			case "sav":
				out.Finish()
				f.Close()
				f = nil
				filename = ""
				savname = ""
				out = nil
				dictDone = false
			}
		}
	}

	return nil
}
