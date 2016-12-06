package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type Label struct {
	Value string
	Desc  string
}

type Var struct {
	Name       string
	ShortName  string
	Type       int32
	Print      byte
	Width      byte
	Decimals   byte
	Label      string
	Default    string
	HasDefault bool
	Labels     []Label
	Value      string
	HasValue   bool
}

var endian = binary.LittleEndian

type Flusher interface {
	Flush() error
}

type SpssWriter struct {
	io.Writer
	Dict     []*Var
	DictMap  map[string]*Var
	ShortMap map[string]*Var
	Count    int32
}

func NewSpssWriter(w io.Writer) *SpssWriter {
	return &SpssWriter{
		Writer:   w,
		DictMap:  make(map[string]*Var),
		ShortMap: make(map[string]*Var),
	}
}

func stob(s string, l int) []byte {
	if len(s) > l {
		s = s[:l]
	} else if len(s) < l {
		s += strings.Repeat(" ", l-len(s))
	}
	return []byte(s)
}

func ftoa(f float64) string {
	return strconv.FormatFloat(f, 'E', -1, 64)
}

func (out *SpssWriter) caseSize() int32 {
	size := int32(0)
	for range out.Dict {
		size++
	}
	return size
}

func (out *SpssWriter) headerRecord(fileLabel string) {
	c := time.Now()
	out.Write(stob("$FL2", 4))                               // rec_tyoe
	out.Write(stob("@(#) SPSS DATA FILE - xml2sav 2.0", 60)) // prod_name
	binary.Write(out, endian, int32(2))                      // layout_code
	binary.Write(out, endian, out.caseSize())                // nominal_case_size
	binary.Write(out, endian, int32(0))                      // compression
	binary.Write(out, endian, int32(0))                      // weight_index
	binary.Write(out, endian, int32(-1))                     // ncases
	binary.Write(out, endian, float64(100))                  // bias
	out.Write(stob(c.Format("02 Jan 06"), 9))                // creation_date
	out.Write(stob(c.Format("15:04:05"), 8))                 // creation_time
	out.Write(stob(fileLabel, 64))                           // file_label
	out.Write(stob("\x00\x00\x00", 3))                       // padding
}

// If you use a buffer, supply it as the flusher argument
// After this close the file
func (out *SpssWriter) updateHeaderNCases(flusher Flusher, seeker io.WriteSeeker) {
	if flusher != nil {
		flusher.Flush()
	}
	seeker.Seek(80, io.SeekStart)
	binary.Write(seeker, endian, out.Count) // ncases in headerRecord
}

func (out *SpssWriter) variableRecords() {
	for _, v := range out.Dict {
		binary.Write(out, endian, int32(2)) // rec_type
		binary.Write(out, endian, v.Type)   // type (0 or strlen)
		if len(v.Label) > 0 {
			binary.Write(out, endian, int32(1)) // has_var_label
		} else {
			binary.Write(out, endian, int32(0)) // has_var_label
		}
		binary.Write(out, endian, int32(0)) // n_missing_values
		format := int32(v.Print)<<16 | int32(v.Width)<<8 | int32(v.Decimals)
		binary.Write(out, endian, format) // print
		binary.Write(out, endian, format) // write
		out.Write(stob(v.ShortName, 8))   // name
		if len(v.Label) > 0 {
			binary.Write(out, endian, int32(len(v.Label))) // label_len
			out.Write([]byte(v.Label))                     // label
			for i, pad := 0, len(v.Label)%4; i < pad; i++ {
				out.Write([]byte{0}) // pad out until multiple of 32 bit
			}
		}
	}
}

func (out *SpssWriter) longVarNameRecords() {
	binary.Write(out, endian, int32(7))  // rec_type
	binary.Write(out, endian, int32(13)) // subtype
	binary.Write(out, endian, int32(1))  // size

	buf := bytes.Buffer{}
	for _, v := range out.Dict {
		buf.Write([]byte(v.ShortName))
		buf.Write([]byte("="))
		buf.Write([]byte(v.Name))
		buf.Write([]byte{9})
	}
	if buf.Len() > 0 {
		buf.UnreadByte() // remove last byte
	}
	binary.Write(out, endian, int32(buf.Len()))
	out.Write(buf.Bytes())
}

func (out *SpssWriter) encodingRecord() {
	binary.Write(out, endian, int32(7))  // rec_type
	binary.Write(out, endian, int32(20)) // subtype
	binary.Write(out, endian, int32(1))  // size
	binary.Write(out, endian, int32(5))  // filler
	out.Write(stob("UTF-8", 5))          // encoding
}

func (out *SpssWriter) terminationRecord() {
	binary.Write(out, endian, int32(999)) // rec_type
	binary.Write(out, endian, int32(0))   // filler
}

func (out *SpssWriter) addVar(v *Var) {
	// Trim long name
	if len(v.Name) > 64 {
		v.Name = v.Name[:64]
	}

	if _, found := out.DictMap[v.Name]; found {
		log.Fatalln("Adding duplicate variable named", v.Name)
	}

	// Create unique short variable name
	short := strings.ToUpper(v.Name)
	count := 1
	if len(short) > 8 {
		short = short[:8]
	}
	for {
		_, found := out.ShortMap[short]
		if !found {
			break
		}
		count++
		s := strconv.Itoa(count)
		if len(short)+len(s) > 8 {
			short = short[:8-len(s)]
		}
		short = short + s
	}
	v.ShortName = short

	out.Dict = append(out.Dict, v)
	out.DictMap[v.Name] = v
	out.ShortMap[v.ShortName] = v
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
	for _, v := range out.Dict {
		if v.HasValue {
			f, err := strconv.ParseFloat(v.Value, 64)
			if err != nil {
				log.Fatalln(err)
			}
			binary.Write(out, endian, f)
		} else if v.HasDefault {
			f, err := strconv.ParseFloat(v.Default, 64)
			if err != nil {
				log.Fatalln(err)
			}
			binary.Write(out, endian, f)
		} else {
			binary.Write(out, endian, -math.MaxFloat64)
		}
	}
	out.Count++
}

func main() {
	fmt.Println("xml2sav")

	file, err := os.Create("test.sav")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	bufout := bufio.NewWriter(file)
	out := NewSpssWriter(bufout)

	out.addVar(&Var{
		Name:     "eenhelelangevarname1",
		Type:     0,
		Print:    5,
		Width:    8,
		Decimals: 2,
		Label:    "Test label",
	})
	out.addVar(&Var{
		Name:     "eenhelelangevarname2",
		Type:     0,
		Print:    5,
		Width:    8,
		Decimals: 2,
		Label:    "Test label",
	})
	out.headerRecord("Export from example.xsav")
	out.variableRecords()
	out.longVarNameRecords()
	out.encodingRecord()
	out.terminationRecord()
	for i := float64(0.0); i < 10; i += 0.1 {
		out.clearCase()
		out.setVar("eenhelelangevarname1", ftoa(i))
		out.setVar("eenhelelangevarname2", ftoa(i+0.03))
		out.writeCase()
	}
	out.updateHeaderNCases(bufout, file)
}
