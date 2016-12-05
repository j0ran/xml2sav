package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

func stob(s string, l int) []byte {
	if len(s) > l {
		s = s[:l]
	}
	for len(s) < l {
		s += " "
	}
	return []byte(s)
}

func header(out io.Writer, fileLabel string) {
	c := time.Now()
	out.Write(stob("$FL2", 4))                               // rec_tyoe
	out.Write(stob("@(#) SPSS DATA FILE - xml2sav 2.0", 60)) // prod_name
	binary.Write(out, binary.LittleEndian, int32(2))         // layout_code
	binary.Write(out, binary.LittleEndian, int32(-1))        // nominal_case_size
	binary.Write(out, binary.LittleEndian, int32(0))         // compression
	binary.Write(out, binary.LittleEndian, int32(0))         // weight_index
	binary.Write(out, binary.LittleEndian, int32(-1))        // ncases
	binary.Write(out, binary.LittleEndian, float64(100))     // bias
	out.Write(stob(c.Format("02 Jan 06"), 9))                // creation_date
	out.Write(stob(c.Format("15:04:05"), 8))                 // creation_time
	out.Write(stob(fileLabel, 64))                           // file_label
	out.Write(stob("\x00\x00\x00", 3))                       // padding
}

func main() {
	fmt.Println("xml2sav")

	file, err := os.Create("test.sav")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	out := bufio.NewWriter(file)
	defer out.Flush()

	header(out, "Export from example.xsav")
}
