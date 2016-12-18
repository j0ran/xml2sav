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
	"encoding/xml"
	"fmt"
	"io"
)

type VarLengths map[string]map[string]int

func (l VarLengths) GetVarLength(savname, varname string) (int, error) {
	sav, found := l[savname]
	if !found {
		return 0, fmt.Errorf("Can not find sav section with name %s\n", savname)
	}
	le, found := sav[varname]
	if !found {
		return 0, fmt.Errorf("Can not find variable %s in sav section %s\n", varname, savname)
	}
	return le, nil
}

func findVarLengths(r io.Reader) (VarLengths, error) {
	v := make(VarLengths)
	var savName string
	var lengths map[string]int

	decoder := xml.NewDecoder(r)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "sav" {
				savName = getAttr(&t, "name")
				lengths = make(map[string]int)
			} else if t.Name.Local == "val" {
				var valxml valXML
				if err = decoder.DecodeElement(&valxml, &t); err != nil {
					return nil, err
				}
				l, found := lengths[valxml.Name]
				if !found {
					l = 0
				}
				if l < len(valxml.Value) || !found {
					lengths[valxml.Name] = len(valxml.Value)
				}
			}
		case xml.EndElement:
			v[savName] = lengths
		}
	}

	return v, nil
}
