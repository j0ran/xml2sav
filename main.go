package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type Label struct {
	Value string
	Desc  string
}

type Var struct {
	Name     string
	Type     byte
	Width    byte
	Decimals byte
	Print    int32
	Write    int32
	Label    string
	Default  string
	Labels   []Label
	Value    string
	HasValue bool
}

var endian = binary.LittleEndian

type SpssWriter struct {
	io.Writer
	Dict    []*Var
	DictMap map[string]*Var
}

func stob(s string, l int) []byte {
	if len(s) > l {
		s = s[:l]
	} else if len(s) < l {
		s += strings.Repeat(" ", l-len(s))
	}
	return []byte(s)
}

func (out *SpssWriter) headerRecord(fileLabel string) {
	c := time.Now()
	out.Write(stob("$FL2", 4))                               // rec_tyoe
	out.Write(stob("@(#) SPSS DATA FILE - xml2sav 2.0", 60)) // prod_name
	binary.Write(out, endian, int32(2))                      // layout_code
	binary.Write(out, endian, int32(-1))                     // nominal_case_size
	binary.Write(out, endian, int32(0))                      // compression
	binary.Write(out, endian, int32(0))                      // weight_index
	binary.Write(out, endian, int32(-1))                     // ncases
	binary.Write(out, endian, float64(100))                  // bias
	out.Write(stob(c.Format("02 Jan 06"), 9))                // creation_date
	out.Write(stob(c.Format("15:04:05"), 8))                 // creation_time
	out.Write(stob(fileLabel, 64))                           // file_label
	out.Write(stob("\x00\x00\x00", 3))                       // padding
}

func (out *SpssWriter) variableRecords() {

}

func (out *SpssWriter) encodingRecord() {
	binary.Write(out, endian, int32(7))  // rec_type
	binary.Write(out, endian, int32(20)) // filler
	binary.Write(out, endian, int32(1))  // size
	binary.Write(out, endian, int32(5))  // filler
	out.Write(stob("UTF-8", 5))          // encoding
}

func (out *SpssWriter) terminationRecord() {
	binary.Write(out, endian, int32(999)) // rec_type
	binary.Write(out, endian, int32(0))   // filler
}

func (out *SpssWriter) addVar(v *Var) {
	out.Dict = append(out.Dict, v)
	out.DictMap[v.Name] = v
}

func (out *SpssWriter) clearCase() {
	for _, v := range out.Dict {
		v.Value = ""
		v.HasValue = false
	}
}

func (out *SpssWriter) setVar(name, value string) {
	v, found := out.DictMap[name]
	if !found {
		log.Fatalln("Can not find the variable named", name)
	}
	v.Value = value
	v.HasValue = true
}

func (out *SpssWriter) writeCase() {

}

func main() {
	fmt.Println("xml2sav")

	file, err := os.Create("test.sav")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	bufout := bufio.NewWriter(file)
	defer bufout.Flush()

	out := &SpssWriter{Writer: bufout}

	out.headerRecord("Export from example.xsav")
}
