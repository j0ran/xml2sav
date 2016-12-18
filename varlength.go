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
